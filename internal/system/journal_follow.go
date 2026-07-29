package system

import (
	"bufio"
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

// FollowJournal streamt neue Journal-Einträge via `journalctl -f` und ruft für
// jede Zeile emit auf. Der Subprozess läuft im Benutzerkontext (r) und wird über
// ctx gesteuert: Wird ctx abgebrochen (Unsubscribe/Verbindungsabbruch), beendet
// CommandContext den journalctl-Prozess — es läuft also nichts ohne aktive
// Subscription.
func FollowJournal(ctx context.Context, r *Runner, q LogQuery, emit func(LogEntry)) error {
	if !commandExists("journalctl") {
		return errMissingTool("journalctl")
	}

	n := q.Lines
	if n <= 0 || n > 1000 {
		n = 50
	}
	args := []string{"-f", "--no-pager", "-o", "json", "-n", strconv.Itoa(n)}
	if q.Unit != "" {
		if err := validUnit(q.Unit); err != nil {
			return err
		}
		args = append(args, "-u", q.Unit)
	}
	if q.Priority != "" && isSafeToken(q.Priority) {
		args = append(args, "-p", q.Priority)
	}

	cmd := r.BuildCommand(ctx, false, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]json.RawMessage
		if json.Unmarshal([]byte(line), &raw) != nil {
			continue
		}
		emit(LogEntry{
			Timestamp: usecToTime(jsonString(raw["__REALTIME_TIMESTAMP"])),
			Hostname:  jsonString(raw["_HOSTNAME"]),
			Unit:      jsonString(raw["_SYSTEMD_UNIT"]),
			Priority:  atoiSafe(jsonString(raw["PRIORITY"])),
			Message:   jsonString(raw["MESSAGE"]),
		})
	}
	_ = cmd.Wait()
	// Abbruch durch ctx ist kein Fehler.
	return ctx.Err()
}
