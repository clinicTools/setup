package system

import (
	"strings"
	"testing"
)

func TestRootRunnerExecutes(t *testing.T) {
	out, err := RootRunner().run("echo", "hallo")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out != "hallo" {
		t.Errorf("out = %q, erwartet hallo", out)
	}
}

func TestRunnerWrapEscalation(t *testing.T) {
	// Runner mit Credential → privilegierte Kommandos werden über sudo geführt.
	user := NewRunner(1000, 1000, []int{1000}, "alice", "geheim")
	name, args, needsPw := user.wrap(true, "systemctl", []string{"start", "nginx"})
	if name != "sudo" || !needsPw {
		t.Fatalf("erwartet sudo-Eskalation, got name=%q needsPw=%v", name, needsPw)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-- systemctl start nginx") {
		t.Errorf("sudo-Argumente unerwartet: %q", joined)
	}

	// Nicht-privilegiert → direkt.
	n2, a2, pw2 := user.wrap(false, "journalctl", []string{"-n", "10"})
	if n2 != "journalctl" || pw2 || len(a2) != 2 {
		t.Errorf("unprivilegiert sollte direkt ausgeführt werden, got %q %v", n2, a2)
	}
}

func TestRootRunnerNoSudo(t *testing.T) {
	// Ohne Credential (Dienst läuft als aktueller Prozess) wird auch
	// „privilegiert" direkt ausgeführt (kein sudo-Wrapper).
	name, _, needsPw := RootRunner().wrap(true, "systemctl", []string{"restart", "x"})
	if name != "systemctl" || needsPw {
		t.Errorf("RootRunner sollte ohne sudo ausführen, got name=%q needsPw=%v", name, needsPw)
	}
}
