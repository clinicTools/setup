package system

import (
	"encoding/json"
	"strconv"
	"strings"
)

// LogEntry ist eine einzelne Journal-Zeile.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Unit      string `json:"unit"`
	Priority  int    `json:"priority"`
	Message   string `json:"message"`
}

// LogQuery parametrisiert eine journalctl-Abfrage.
type LogQuery struct {
	Unit     string // optionaler Unit-Filter
	Priority string // optional: emerg..debug oder 0..7
	Lines    int    // Anzahl der jüngsten Zeilen (Default 200)
}

// Logs liest die jüngsten Journal-Einträge via journalctl (JSON-Ausgabe).
func Logs(q LogQuery) ([]LogEntry, error) {
	if !commandExists("journalctl") {
		return nil, errMissingTool("journalctl")
	}
	if q.Lines <= 0 || q.Lines > 2000 {
		q.Lines = 200
	}
	args := []string{"--no-pager", "-o", "json", "-n", strconv.Itoa(q.Lines)}
	if q.Unit != "" {
		if err := validUnit(q.Unit); err != nil {
			return nil, err
		}
		args = append(args, "-u", q.Unit)
	}
	if q.Priority != "" && isSafeToken(q.Priority) {
		args = append(args, "-p", q.Priority)
	}

	out, err := run("journalctl", args...)
	if err != nil {
		return nil, err
	}

	var entries []LogEntry
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		var raw map[string]json.RawMessage
		if json.Unmarshal([]byte(l), &raw) != nil {
			continue
		}
		entries = append(entries, LogEntry{
			Timestamp: usecToTime(jsonString(raw["__REALTIME_TIMESTAMP"])),
			Hostname:  jsonString(raw["_HOSTNAME"]),
			Unit:      jsonString(raw["_SYSTEMD_UNIT"]),
			Priority:  atoiSafe(jsonString(raw["PRIORITY"])),
			Message:   jsonString(raw["MESSAGE"]),
		})
	}
	return entries, nil
}

// jsonString dekodiert einen JSON-Wert zu String; journald liefert MESSAGE
// gelegentlich als Byte-Array (Zahlenliste) — dieser Fall wird abgefangen.
func jsonString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var bytes []int
	if json.Unmarshal(raw, &bytes) == nil {
		b := make([]byte, len(bytes))
		for i, v := range bytes {
			b[i] = byte(v)
		}
		return string(b)
	}
	return ""
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// usecToTime wandelt einen Mikrosekunden-Timestamp in RFC3339 (UTC) um.
func usecToTime(usec string) string {
	n, err := strconv.ParseInt(usec, 10, 64)
	if err != nil {
		return ""
	}
	return unixMicroRFC3339(n)
}

// isSafeToken erlaubt nur alphanumerische Filterwerte ohne Sonderzeichen.
func isSafeToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
