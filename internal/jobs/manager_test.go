package jobs

import (
	"strings"
	"testing"
	"time"
)

func TestJobRunsAndStreams(t *testing.T) {
	m := NewManager()
	job := m.Start("echo", "echo", []string{"hallo-welt"})

	// Auf Abschluss warten.
	select {
	case <-job.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Job nicht rechtzeitig beendet")
	}

	if job.Status != StatusSucceeded {
		t.Errorf("Status = %q, erwartet succeeded", job.Status)
	}

	replay, _, _ := job.Subscribe()
	found := false
	for _, l := range replay {
		if strings.Contains(l.Text, "hallo-welt") {
			found = true
		}
	}
	if !found {
		t.Error("erwartete Ausgabezeile 'hallo-welt' nicht gefunden")
	}
}

func TestJobFailureExitCode(t *testing.T) {
	m := NewManager()
	job := m.Start("false", "false", nil)
	select {
	case <-job.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Job nicht beendet")
	}
	if job.Status != StatusFailed {
		t.Errorf("Status = %q, erwartet failed", job.Status)
	}
}

func TestManagerListAndGet(t *testing.T) {
	m := NewManager()
	job := m.Start("echo", "echo", []string{"x"})
	<-job.Done()

	if got, ok := m.Get(job.ID); !ok || got.ID != job.ID {
		t.Error("Get lieferte den Job nicht")
	}
	if len(m.List()) != 1 {
		t.Errorf("List-Länge = %d, erwartet 1", len(m.List()))
	}
}
