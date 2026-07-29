// Package ws implementiert eine gemultiplexte WebSocket-Schnittstelle nach dem
// Vorbild von Cockpit: Über eine einzige Verbindung eröffnet der Client
// beliebig viele „Channels" (Subscriptions). Der serverseitige Collector eines
// Channels läuft ausschließlich, solange die Subscription besteht — wird sie
// beendet oder bricht die Verbindung ab, stoppt die zugehörige Goroutine (und
// ggf. ihr Subprozess wie `journalctl -f`) sofort.
package ws

import "encoding/json"

// ClientMessage ist eine vom Browser gesendete Steuernachricht.
type ClientMessage struct {
	// Type ∈ {subscribe, unsubscribe, ping}.
	Type string `json:"type"`
	// ID identifiziert die Subscription eindeutig (vom Client vergeben).
	ID string `json:"id"`
	// Channel benennt den gewünschten Datenstrom (nur bei subscribe).
	Channel string `json:"channel"`
	// Params sind channel-spezifische Parameter (optional).
	Params json.RawMessage `json:"params,omitempty"`
}

// ServerMessage ist eine an den Browser gesendete Nachricht.
type ServerMessage struct {
	// Type ∈ {ready, message, error, closed, pong}.
	Type string `json:"type"`
	// ID referenziert die zugehörige Subscription.
	ID string `json:"id,omitempty"`
	// Channel benennt den Datenstrom (Komfort fürs Frontend-Routing).
	Channel string `json:"channel,omitempty"`
	// Payload ist die eigentliche Nutzlast eines message-Events.
	Payload any `json:"payload,omitempty"`
	// Error beschreibt einen Fehlerzustand.
	Error string `json:"error,omitempty"`
}
