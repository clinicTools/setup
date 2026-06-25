package ws

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/clinictools/setup/internal/jobs"
)

// Emitter sendet eine Nutzlast an den Client. Aufrufe sind nicht-blockierend;
// bei überlastetem Client werden Nachrichten verworfen statt zu stauen.
type Emitter func(payload any)

// RunFunc ist der Collector eines Channels. Er wird je Subscription EINMAL
// aufgerufen und blockiert, bis ctx abgebrochen wird (Unsubscribe oder
// Verbindungsabbruch). Sämtlicher Zustand (z. B. vorherige Metrik-Messung) lebt
// als lokale Variable und ist damit pro Subscription isoliert.
type RunFunc func(ctx context.Context, params json.RawMessage, emit Emitter) error

// Registry bildet Channel-Namen auf ihre Collector-Funktion ab.
type Registry map[string]RunFunc

// NewRegistry bündelt die Cockpit-orientierten Live-Channels. Der Job-Manager
// wird für den „jobs"-Channel (Live-Log-Verfolgung) benötigt.
func NewRegistry(jm *jobs.Manager) Registry {
	return Registry{
		"metrics":   metricsChannel,
		"journal":   journalChannel,
		"services":  servicesChannel,
		"processes": processesChannel,
		"jobs":      jobsChannel(jm),
	}
}

// jobsChannel verfolgt die Ausgabe eines Jobs: zunächst der gepufferte Verlauf,
// danach neue Zeilen in Echtzeit, abschließend ein Status-Event.
func jobsChannel(jm *jobs.Manager) RunFunc {
	return func(ctx context.Context, params json.RawMessage, emit Emitter) error {
		var p struct {
			JobID string `json:"jobId"`
		}
		if len(params) > 0 {
			_ = json.Unmarshal(params, &p)
		}
		job, ok := jm.Get(p.JobID)
		if !ok {
			return errors.New("Job nicht gefunden")
		}

		replay, ch, unsubscribe := job.Subscribe()
		defer unsubscribe()
		for _, line := range replay {
			emit(line)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case line, open := <-ch:
				if !open {
					emit(map[string]any{
						"finished": true,
						"status":   job.Status,
						"exitCode": job.ExitCode,
						"error":    job.Error,
					})
					return nil
				}
				emit(line)
			}
		}
	}
}
