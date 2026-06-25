package system

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// KillProcess sendet ein Signal an einen Prozess. signal ∈ {TERM, KILL, HUP,
// INT}. PIDs ≤ 1 sind nicht erlaubt (Schutz von init/Kernel-Threads).
func KillProcess(pid int, signal string) error {
	if pid <= 1 {
		return fmt.Errorf("ungültige PID: %d", pid)
	}
	var sig syscall.Signal
	switch signal {
	case "", "TERM":
		sig = syscall.SIGTERM
	case "KILL":
		sig = syscall.SIGKILL
	case "HUP":
		sig = syscall.SIGHUP
	case "INT":
		sig = syscall.SIGINT
	default:
		return fmt.Errorf("unzulässiges Signal: %q", signal)
	}
	return syscall.Kill(pid, sig)
}

// Process beschreibt einen laufenden Prozess (aus /proc/<pid>).
type Process struct {
	PID     int     `json:"pid"`
	PPID    int     `json:"ppid"`
	User    string  `json:"user"`
	State   string  `json:"state"`
	Command string  `json:"command"`
	RSSMB   float64 `json:"rssMB"`
	Threads int     `json:"threads"`
}

// pageSize wird für die RSS-Umrechnung (Seiten -> Bytes) benötigt.
var pageSize = int64(os.Getpagesize())

// Processes liest die laufenden Prozesse aus /proc und sortiert nach
// Speicherverbrauch absteigend. limit begrenzt die Ergebnismenge (0 = alle).
func Processes(limit int) ([]Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	uidNames := uidNameIndex()

	var procs []Process
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if p, ok := readProcess(pid, uidNames); ok {
			procs = append(procs, p)
		}
	}

	sort.Slice(procs, func(i, j int) bool { return procs[i].RSSMB > procs[j].RSSMB })
	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}
	return procs, nil
}

func readProcess(pid int, uidNames map[string]string) (Process, bool) {
	base := filepath.Join("/proc", strconv.Itoa(pid))

	statm, err := os.ReadFile(filepath.Join(base, "statm"))
	if err != nil {
		return Process{}, false
	}
	var rssPages int64
	if f := strings.Fields(string(statm)); len(f) >= 2 {
		rssPages, _ = strconv.ParseInt(f[1], 10, 64)
	}

	p := Process{
		PID:     pid,
		RSSMB:   float64(rssPages*pageSize) / (1024 * 1024),
		Command: processCommand(base),
		User:    uidNames[fileOwnerUID(base)],
	}

	if status, err := os.ReadFile(filepath.Join(base, "status")); err == nil {
		for _, l := range strings.Split(string(status), "\n") {
			key, val, ok := strings.Cut(l, ":")
			if !ok {
				continue
			}
			val = strings.TrimSpace(val)
			switch key {
			case "PPid":
				p.PPID, _ = strconv.Atoi(val)
			case "State":
				p.State = val
			case "Threads":
				p.Threads, _ = strconv.Atoi(val)
			}
		}
	}
	return p, true
}

// processCommand liefert die Kommandozeile (cmdline) oder den comm-Namen.
func processCommand(base string) string {
	if cmdline, err := os.ReadFile(filepath.Join(base, "cmdline")); err == nil {
		s := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	if comm, err := os.ReadFile(filepath.Join(base, "comm")); err == nil {
		return strings.TrimSpace(string(comm))
	}
	return ""
}

func fileOwnerUID(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		return ""
	}
	if st, ok := statUID(fi); ok {
		return strconv.FormatUint(uint64(st), 10)
	}
	return ""
}

// uidNameIndex bildet UID(string) -> Username ab.
func uidNameIndex() map[string]string {
	out := map[string]string{}
	data, err := readFile("/etc/passwd")
	if err != nil {
		return out
	}
	for _, l := range lines(data) {
		f := strings.Split(l, ":")
		if len(f) >= 3 {
			out[f[2]] = f[0]
		}
	}
	return out
}
