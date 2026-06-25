package system

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HostInfo fasst Eckdaten des Systems für das Dashboard zusammen.
type HostInfo struct {
	Hostname      string   `json:"hostname"`
	OS            string   `json:"os"`
	PrettyName    string   `json:"prettyName"`
	Kernel        string   `json:"kernel"`
	Architecture  string   `json:"architecture"`
	Virtualization string  `json:"virtualization,omitempty"`
	UptimeSeconds int64    `json:"uptimeSeconds"`
	BootTime      string   `json:"bootTime"`
	CPUModel      string   `json:"cpuModel"`
	CPUCores      int      `json:"cpuCores"`
	LoadAvg       [3]float64 `json:"loadAvg"`
	Memory        MemoryInfo `json:"memory"`
	Swap          MemoryInfo `json:"swap"`
}

// MemoryInfo beschreibt Arbeitsspeicher- oder Swap-Auslastung in Bytes.
type MemoryInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// Info sammelt die Host-Übersicht aus /proc, /etc/os-release und uname.
func Info() (*HostInfo, error) {
	info := &HostInfo{
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
	}

	info.Hostname, _ = os.Hostname()

	rel := parseOSRelease()
	info.OS = rel["ID"]
	info.PrettyName = rel["PRETTY_NAME"]
	if info.PrettyName == "" {
		info.PrettyName = "Debian GNU/Linux"
	}

	if out, err := run("uname", "-r"); err == nil {
		info.Kernel = out
	}

	info.UptimeSeconds = readUptime()
	info.BootTime = time.Now().Add(-time.Duration(info.UptimeSeconds) * time.Second).Format(time.RFC3339)
	info.LoadAvg = readLoadAvg()
	info.CPUModel = readCPUModel()
	info.Memory, info.Swap = readMemInfo()

	if commandExists("systemd-detect-virt") {
		if v, err := run("systemd-detect-virt"); err == nil && v != "none" {
			info.Virtualization = v
		}
	}

	return info, nil
}

func parseOSRelease() map[string]string {
	out := map[string]string{}
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return out
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[key] = strings.Trim(val, `"`)
	}
	return out
}

func readUptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	f, _ := strconv.ParseFloat(fields[0], 64)
	return int64(f)
}

func readLoadAvg() [3]float64 {
	var out [3]float64
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return out
	}
	fields := strings.Fields(string(data))
	for i := 0; i < 3 && i < len(fields); i++ {
		out[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return out
}

func readCPUModel() string {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, val, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) == "model name" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// readMemInfo liest RAM und Swap aus /proc/meminfo (Werte dort in kB).
func readMemInfo() (mem MemoryInfo, swap MemoryInfo) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	vals := map[string]uint64{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, val, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		fields := strings.Fields(val)
		if len(fields) == 0 {
			continue
		}
		n, _ := strconv.ParseUint(fields[0], 10, 64)
		vals[strings.TrimSpace(key)] = n * 1024 // kB -> Bytes
	}

	mem.Total = vals["MemTotal"]
	mem.Free = vals["MemAvailable"]
	mem.Used = mem.Total - mem.Free

	swap.Total = vals["SwapTotal"]
	swap.Free = vals["SwapFree"]
	swap.Used = swap.Total - swap.Free
	return
}
