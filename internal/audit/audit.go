// Package audit protokolliert administrative (schreibende) Aktionen
// nachvollziehbar: wer, wann, was, mit welchem Ergebnis. Einträge werden im
// Speicher als Ringpuffer gehalten (für die UI) und optional zeilenweise als
// JSON in eine Datei geschrieben (revisionssicher, überdauert Neustarts).
package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry ist ein einzelner Audit-Datensatz.
type Entry struct {
	Time    string `json:"time"`
	User    string `json:"user"`
	Action  string `json:"action"`
	Target  string `json:"target,omitempty"`
	Success bool   `json:"success"`
	Detail  string `json:"detail,omitempty"`
	IP      string `json:"ip,omitempty"`
}

// Logger sammelt Audit-Einträge.
type Logger struct {
	mu      sync.Mutex
	entries []Entry
	max     int
	file    *os.File
}

// New erzeugt einen Logger. Ist path gesetzt, wird zusätzlich in diese Datei
// geschrieben (Anhängen). Fehler beim Öffnen werden toleriert (nur In-Memory).
func New(path string) *Logger {
	l := &Logger{max: 2000}
	if path != "" {
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			l.file = f
		}
	}
	return l
}

// Record hält einen Eintrag fest (Zeitstempel wird gesetzt, falls leer).
func (l *Logger) Record(e Entry) {
	if e.Time == "" {
		e.Time = time.Now().Format(time.RFC3339)
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, e)
	if len(l.entries) > l.max {
		l.entries = l.entries[len(l.entries)-l.max:]
	}
	if l.file != nil {
		if data, err := json.Marshal(e); err == nil {
			_, _ = l.file.Write(append(data, '\n'))
		}
	}
}

// List liefert die jüngsten Einträge (neueste zuerst), begrenzt durch limit.
func (l *Logger) List(limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}
	out := make([]Entry, 0, limit)
	for i := len(l.entries) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, l.entries[i])
	}
	return out
}
