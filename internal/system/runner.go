package system

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Runner führt Kommandos im Kontext eines bestimmten Betriebssystem-Benutzers
// aus — analog zum cockpit-bridge, der unter der UID des angemeldeten Benutzers
// läuft. Lesende Kommandos laufen direkt als dieser Benutzer; privilegierte
// Kommandos werden über `sudo -S` mit dem (in der Session gehaltenen) Passwort
// eskaliert, sodass die Autorisierung vom Betriebssystem (sudoers) und nicht
// von der Anwendung entschieden wird.
//
// Ist cred == nil (Dienst läuft selbst nicht als root, z. B. Entwicklung), wird
// direkt als aktueller Prozess ausgeführt — ohne UID-Wechsel und ohne sudo.
type Runner struct {
	cred     *syscall.Credential
	password string
	username string
}

// NewRunner erzeugt einen Runner für einen Benutzer. uid < 0 erzeugt einen
// Runner ohne Credential (Direktausführung als aktueller Prozess).
func NewRunner(uid, gid int, groups []int, username, password string) *Runner {
	r := &Runner{password: password, username: username}
	if uid >= 0 {
		g := make([]uint32, 0, len(groups))
		for _, v := range groups {
			g = append(g, uint32(v))
		}
		r.cred = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), Groups: g}
	}
	return r
}

// RootRunner führt Kommandos als aktueller Prozess aus (Bootstrap/Tests/Dev).
func RootRunner() *Runner { return &Runner{} }

// dropsPrivileges meldet, ob der Runner auf einen Benutzer wechselt.
func (r *Runner) dropsPrivileges() bool { return r.cred != nil }

// run führt ein Kommando als der Benutzer aus und liefert stdout (getrimmt).
func (r *Runner) run(name string, args ...string) (string, error) {
	return r.exec("", false, name, args...)
}

// runStdin führt ein Kommando als der Benutzer mit stdin aus.
func (r *Runner) runStdin(stdin, name string, args ...string) (string, error) {
	return r.exec(stdin, false, name, args...)
}

// sudo führt ein Kommando privilegiert aus (über sudo, falls UID gewechselt wird).
func (r *Runner) sudo(name string, args ...string) (string, error) {
	return r.exec("", true, name, args...)
}

// sudoStdin führt ein privilegiertes Kommando mit zusätzlichem stdin aus.
func (r *Runner) sudoStdin(stdin, name string, args ...string) (string, error) {
	return r.exec(stdin, true, name, args...)
}

// wrap bestimmt das tatsächlich auszuführende Kommando: Bei privilegierter
// Ausführung mit UID-Wechsel wird sudo vorangestellt; needsPwStdin signalisiert,
// dass das Passwort als erste stdin-Zeile erwartet wird.
func (r *Runner) wrap(privileged bool, name string, args []string) (cmdName string, cmdArgs []string, needsPwStdin bool) {
	if privileged && r.dropsPrivileges() {
		return "sudo", append([]string{"-S", "-k", "-p", "", "--", name}, args...), true
	}
	return name, args, false
}

// BuildCommand erzeugt ein konfiguriertes *exec.Cmd (inkl. sudo-Eskalation und
// Benutzer-Credential), dessen Ausgabe der Aufrufer streamen kann (Jobs). Bei
// privilegierter Ausführung wird das Passwort bereits als stdin gesetzt.
func (r *Runner) BuildCommand(ctx context.Context, privileged bool, name string, args ...string) *exec.Cmd {
	cmdName, cmdArgs, needsPw := r.wrap(privileged, name, args)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	if r.cred != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: r.cred}
	}
	if needsPw {
		cmd.Stdin = strings.NewReader(r.password + "\n")
	}
	return cmd
}

// exec ist die zentrale Ausführungsroutine.
func (r *Runner) exec(stdin string, privileged bool, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmdName, cmdArgs, needsPw := r.wrap(privileged, name, args)
	var stdinBuf bytes.Buffer
	if needsPw {
		stdinBuf.WriteString(r.password)
		stdinBuf.WriteByte('\n')
	}
	stdinBuf.WriteString(stdin)

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	if r.cred != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: r.cred}
	}
	if stdinBuf.Len() > 0 {
		cmd.Stdin = &stdinBuf
	}

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

// CanEscalate prüft, ob der Benutzer privilegierte Kommandos ausführen darf
// (echte sudo-Berechtigung laut OS) — ersetzt die reine Gruppenheuristik.
func (r *Runner) CanEscalate() bool {
	if !r.dropsPrivileges() {
		return true // Dienst läuft bereits als root.
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// `sudo -S -v` validiert die Anmeldedaten und die sudo-Berechtigung.
	cmd := exec.CommandContext(ctx, "sudo", "-S", "-k", "-p", "", "-v")
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: r.cred}
	cmd.Stdin = strings.NewReader(r.password + "\n")
	return cmd.Run() == nil
}
