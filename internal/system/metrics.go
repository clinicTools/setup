package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// CPUTimes hält aggregierte CPU-Jiffies aus /proc/stat für die Delta-Berechnung.
type CPUTimes struct {
	Total uint64
	Idle  uint64
}

// ReadCPUTimes liest die Summenzeile aus /proc/stat. Die CPU-Auslastung ergibt
// sich aus der Differenz zweier Messungen: 1 - ΔIdle/ΔTotal.
func ReadCPUTimes() (CPUTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return CPUTimes{}, err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}
		var t CPUTimes
		for i := 1; i < len(fields); i++ {
			v, _ := strconv.ParseUint(fields[i], 10, 64)
			t.Total += v
			// Feld 4 = idle, Feld 5 = iowait (beide als „untätig" gewertet).
			if i == 4 || i == 5 {
				t.Idle += v
			}
		}
		return t, nil
	}
	return CPUTimes{}, errMissingTool("/proc/stat")
}

// CPUPercent berechnet die Auslastung in Prozent zwischen zwei Messungen.
func CPUPercent(prev, cur CPUTimes) float64 {
	totalDelta := float64(cur.Total - prev.Total)
	idleDelta := float64(cur.Idle - prev.Idle)
	if totalDelta <= 0 {
		return 0
	}
	usage := (1 - idleDelta/totalDelta) * 100
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}
	return usage
}

// NetCounter hält die übertragenen Bytes einer Schnittstelle.
type NetCounter struct {
	RxBytes uint64
	TxBytes uint64
}

// ReadNetDev liest die Byte-Zähler aller Schnittstellen aus /proc/net/dev.
func ReadNetDev() (map[string]NetCounter, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	out := map[string]NetCounter{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		name, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue // Kopfzeilen ohne ":"
		}
		name = strings.TrimSpace(name)
		if name == "lo" {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		out[name] = NetCounter{RxBytes: rx, TxBytes: tx}
	}
	return out, nil
}

// Meminfo exportiert die RAM-/Swap-Auslastung (Wrapper für ws-Channels).
func Meminfo() (MemoryInfo, MemoryInfo) { return readMemInfo() }

// LoadAvg exportiert den Lastdurchschnitt (Wrapper für ws-Channels).
func LoadAvg() [3]float64 { return readLoadAvg() }
