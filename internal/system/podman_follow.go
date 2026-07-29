package system

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"sync"
)

// FollowContainerLogs streamt die Ausgabe eines Containers (`podman logs -f`).
// Der Subprozess läuft im Benutzerkontext (privilegiert, da System-Container)
// und wird über ctx beendet: Bricht die Subscription ab, endet auch der
// Prozess — es läuft nichts ohne aktiven Zuhörer.
func FollowContainerLogs(ctx context.Context, r *Runner, id string, tail int, emit func(string)) error {
	if !PodmanInstalled() {
		return errMissingTool("podman")
	}
	if !validContainerRef(id) {
		return fmt.Errorf("ungültige Container-Kennung: %q", id)
	}
	if tail <= 0 || tail > 1000 {
		tail = 100
	}

	cmd := r.BuildCommand(ctx, true, "podman", "logs", "-f", "--tail", itoa(tail), id)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// podman schreibt Container-stderr auf stderr — beide Ströme mitlesen.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	scan := func(r *bufio.Scanner) {
		defer wg.Done()
		r.Buffer(make([]byte, 64*1024), 1024*1024)
		for r.Scan() {
			if line := strings.TrimRight(r.Text(), "\r"); line != "" {
				emit(line)
			}
		}
	}
	wg.Add(2)
	go scan(bufio.NewScanner(stdout))
	go scan(bufio.NewScanner(stderr))
	wg.Wait()

	_ = cmd.Wait()
	return ctx.Err()
}
