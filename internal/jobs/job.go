// Package jobs verwaltet langlaufende Kommandos (z. B. `apt upgrade`,
// k3s-Installation) als asynchrone Jobs mit Live-Log-Stream. Ein Job kapselt
// einen Subprozess; sein Output wird zeilenweise gepuffert und an alle
// Abonnenten verteilt. Über den WebSocket-Channel „jobs" verfolgt das Frontend
// einen Job in Echtzeit — analog zu Cockpits Terminal-/Task-Ausgaben.
package jobs

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Builder erzeugt das auszuführende Kommando für einen Job mit dem übergebenen
// (abbrechbaren) Context. So entscheidet der Aufrufer über Benutzerkontext,
// sudo-Eskalation und Umgebung (via system.Runner), während der Job nur den
// Lebenszyklus und das Streaming verwaltet.
type Builder func(ctx context.Context) *exec.Cmd

// Status beschreibt den Lebenszyklus eines Jobs.
type Status string

const (
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// maxLines begrenzt den Ringpuffer pro Job.
const maxLines = 5000

// LogLine ist eine einzelne Ausgabezeile eines Jobs.
type LogLine struct {
	Seq    int    `json:"seq"`
	Time   string `json:"time"`
	Stream string `json:"stream"` // stdout | stderr | system
	Text   string `json:"text"`
}

// Job repräsentiert einen laufenden oder abgeschlossenen Vorgang.
type Job struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     Status `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt,omitempty"`
	ExitCode   int    `json:"exitCode"`
	Error      string `json:"error,omitempty"`

	mu     sync.Mutex
	lines  []LogLine
	seq    int
	subs   map[int]chan LogLine
	nextID int
	done   chan struct{}
	cancel context.CancelFunc
}

// newJob initialisiert einen Job-Datensatz.
func newJob(id, name string) *Job {
	return &Job{
		ID:        id,
		Name:      name,
		Status:    StatusRunning,
		StartedAt: time.Now().Format(time.RFC3339),
		subs:      map[int]chan LogLine{},
		done:      make(chan struct{}),
	}
}

// run startet den Subprozess und verteilt dessen Ausgabe. Läuft in eigener
// Goroutine bis zum Prozessende.
func (j *Job) run(parent context.Context, build Builder) {
	ctx, cancel := context.WithCancel(parent)
	j.mu.Lock()
	j.cancel = cancel
	j.mu.Unlock()
	defer cancel()

	cmd := build(ctx)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	j.appendLine("system", "$ "+strings.Join(redactArgs(cmd.Args), " "))

	if err := cmd.Start(); err != nil {
		j.finish(StatusFailed, 1, err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go j.pump(&wg, stdout, "stdout")
	go j.pump(&wg, stderr, "stderr")
	wg.Wait()

	err := cmd.Wait()
	switch {
	case ctx.Err() == context.Canceled:
		j.finish(StatusCanceled, -1, "abgebrochen")
	case err != nil:
		code := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		j.finish(StatusFailed, code, err.Error())
	default:
		j.finish(StatusSucceeded, 0, "")
	}
}

// pump liest einen Stream zeilenweise.
func (j *Job) pump(wg *sync.WaitGroup, r io.Reader, stream string) {
	defer wg.Done()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		j.appendLine(stream, sc.Text())
	}
}

// appendLine puffert eine Zeile und verteilt sie an alle Abonnenten.
func (j *Job) appendLine(stream, text string) {
	j.mu.Lock()
	j.seq++
	line := LogLine{Seq: j.seq, Time: time.Now().Format(time.RFC3339), Stream: stream, Text: text}
	j.lines = append(j.lines, line)
	if len(j.lines) > maxLines {
		j.lines = j.lines[len(j.lines)-maxLines:]
	}
	for _, ch := range j.subs {
		select {
		case ch <- line:
		default: // langsamer Abonnent: Zeile verwerfen
		}
	}
	j.mu.Unlock()
}

// finish setzt den Endzustand und weckt Abonnenten.
func (j *Job) finish(status Status, code int, errMsg string) {
	j.mu.Lock()
	j.Status = status
	j.ExitCode = code
	j.Error = errMsg
	j.FinishedAt = time.Now().Format(time.RFC3339)
	for _, ch := range j.subs {
		close(ch)
	}
	j.subs = map[int]chan LogLine{}
	close(j.done)
	j.mu.Unlock()
}

// Subscribe liefert den bisherigen Puffer sowie einen Kanal für neue Zeilen.
// Ist der Job bereits fertig, ist der Kanal geschlossen (nur Replay). Die
// zurückgegebene Funktion meldet den Abonnenten wieder ab.
func (j *Job) Subscribe() (replay []LogLine, ch <-chan LogLine, unsubscribe func()) {
	j.mu.Lock()
	defer j.mu.Unlock()

	replay = append([]LogLine(nil), j.lines...)
	out := make(chan LogLine, 256)
	if j.Status != StatusRunning {
		close(out)
		return replay, out, func() {}
	}
	id := j.nextID
	j.nextID++
	j.subs[id] = out
	return replay, out, func() {
		j.mu.Lock()
		if c, ok := j.subs[id]; ok {
			delete(j.subs, id)
			close(c)
		}
		j.mu.Unlock()
	}
}

// Done liefert einen Kanal, der bei Jobende geschlossen wird.
func (j *Job) Done() <-chan struct{} { return j.done }

// Cancel bricht den laufenden Job ab.
func (j *Job) Cancel() {
	j.mu.Lock()
	cancel := j.cancel
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// redactArgs bereitet die Kommandozeile für die Anzeige auf: Eine
// sudo-Eskalation wird auf „sudo <kommando> …" verkürzt (die Flags -S/-k/-p
// sind nur Mechanik; das Passwort steht ohnehin auf stdin, nie in den Args).
func redactArgs(args []string) []string {
	if len(args) > 0 && args[0] == "sudo" {
		for i, a := range args {
			if a == "--" {
				return append([]string{"sudo"}, args[i+1:]...)
			}
		}
	}
	return args
}
