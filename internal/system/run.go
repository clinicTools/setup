// Package system kapselt das Auslesen und Verändern des Debian-Hostzustands.
// Die meisten Operationen delegieren an Standard-Systemwerkzeuge (systemctl,
// apt, ip, …) bzw. lesen direkt aus /proc und /sys.
package system

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

// readFile liest eine Datei als String (Wrapper um os.ReadFile).
func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// osHostname ist ein dünner Wrapper um os.Hostname für die Testbarkeit.
func osHostname() (string, error) {
	return os.Hostname()
}

// readDir liefert die Einträge eines Verzeichnisses als Namensliste.
func readDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out, nil
}

// unixMicroRFC3339 formatiert einen Unix-Mikrosekunden-Zeitstempel als RFC3339.
func unixMicroRFC3339(usec int64) string {
	return time.UnixMicro(usec).UTC().Format(time.RFC3339)
}

// defaultTimeout begrenzt die Laufzeit externer Kommandos.
const defaultTimeout = 30 * time.Second

// ErrCommandFailed signalisiert einen nicht-null Exit-Code mit stderr-Auszug.
type ErrCommandFailed struct {
	Cmd    string
	Stderr string
	Err    error
}

func (e *ErrCommandFailed) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = e.Err.Error()
	}
	return e.Cmd + ": " + msg
}

func (e *ErrCommandFailed) Unwrap() error { return e.Err }

// run führt ein Kommando mit Timeout aus und liefert stdout (getrimmt).
func run(name string, args ...string) (string, error) {
	return runCtx(context.Background(), name, args...)
}

func runCtx(parent context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "", &ErrCommandFailed{Cmd: name, Stderr: "Zeitüberschreitung", Err: ctx.Err()}
	}
	if err != nil {
		return strings.TrimSpace(stdout.String()), &ErrCommandFailed{
			Cmd: name, Stderr: stderr.String(), Err: err,
		}
	}
	return strings.TrimSpace(stdout.String()), nil
}

// commandExists prüft, ob ein Programm im PATH auffindbar ist. Damit lassen
// sich optionale Werkzeuge (ufw, timedatectl) elegant überspringen.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// lines zerlegt eine Kommandoausgabe in nicht-leere, getrimmte Zeilen.
func lines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// errMissingTool beschreibt ein nicht installiertes optionales Werkzeug.
func errMissingTool(tool string) error {
	return errors.New("Werkzeug nicht verfügbar: " + tool)
}
