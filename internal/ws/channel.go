package ws

import (
	"context"
	"encoding/json"
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

// DefaultRegistry bündelt die Cockpit-orientierten Live-Channels.
func DefaultRegistry() Registry {
	return Registry{
		"metrics":   metricsChannel,
		"journal":   journalChannel,
		"services":  servicesChannel,
		"processes": processesChannel,
	}
}
