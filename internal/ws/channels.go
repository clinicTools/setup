package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clinictools/setup/internal/system"
)

// metricsChannel liefert Live-Kennzahlen (CPU, RAM, Swap, Last, Netzdurchsatz)
// im Sekundentakt. CPU- und Netz-Raten werden aus der Differenz zweier
// /proc-Messungen berechnet; der Zustand lebt als lokale Variable und ist damit
// pro Subscription isoliert.
func metricsChannel(ctx context.Context, _ json.RawMessage, emit Emitter) error {
	prevCPU, _ := system.ReadCPUTimes()
	prevNet, _ := system.ReadNetDev()
	prevTime := time.Now()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			curCPU, _ := system.ReadCPUTimes()
			curNet, _ := system.ReadNetDev()
			mem, swap := system.Meminfo()
			elapsed := now.Sub(prevTime).Seconds()

			interfaces := make([]map[string]any, 0, len(curNet))
			for name, cur := range curNet {
				prev, ok := prevNet[name]
				if !ok || elapsed <= 0 {
					continue
				}
				interfaces = append(interfaces, map[string]any{
					"name":   name,
					"rxRate": float64(cur.RxBytes-prev.RxBytes) / elapsed,
					"txRate": float64(cur.TxBytes-prev.TxBytes) / elapsed,
				})
			}

			emit(map[string]any{
				"timestamp":  now.UnixMilli(),
				"cpuPercent": system.CPUPercent(prevCPU, curCPU),
				"memory":     mem,
				"swap":       swap,
				"loadAvg":    system.LoadAvg(),
				"interfaces": interfaces,
			})

			prevCPU, prevNet, prevTime = curCPU, curNet, now
		}
	}
}

// journalChannel streamt neue Journal-Einträge (journalctl -f). Der Subprozess
// existiert nur, solange die Subscription aktiv ist.
func journalChannel(ctx context.Context, params json.RawMessage, emit Emitter) error {
	var q struct {
		Unit     string `json:"unit"`
		Priority string `json:"priority"`
		Lines    int    `json:"lines"`
	}
	if len(params) > 0 {
		_ = json.Unmarshal(params, &q)
	}
	return system.FollowJournal(ctx, system.LogQuery{
		Unit:     q.Unit,
		Priority: q.Priority,
		Lines:    q.Lines,
	}, func(e system.LogEntry) {
		emit(e)
	})
}

// servicesChannel sendet alle 3 s eine aktuelle Dienst-Momentaufnahme.
func servicesChannel(ctx context.Context, _ json.RawMessage, emit Emitter) error {
	return pollLoop(ctx, 3*time.Second, func() (any, error) {
		return system.Services()
	}, emit)
}

// processesChannel sendet alle 2 s die Top-Prozesse nach Speicherverbrauch.
func processesChannel(ctx context.Context, _ json.RawMessage, emit Emitter) error {
	return pollLoop(ctx, 2*time.Second, func() (any, error) {
		return system.Processes(40)
	}, emit)
}

// pollLoop ruft fn sofort und danach im Intervall auf und emittiert das
// Ergebnis. Schlägt bereits der erste Aufruf fehl, wird der Fehler zurück-
// gegeben (der Client erhält eine error-Nachricht); spätere Fehler werden
// übersprungen, um transiente Aussetzer zu tolerieren.
func pollLoop(ctx context.Context, interval time.Duration, fn func() (any, error), emit Emitter) error {
	data, err := fn()
	if err != nil {
		return err
	}
	emit(data)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if data, err := fn(); err == nil {
				emit(data)
			}
		}
	}
}
