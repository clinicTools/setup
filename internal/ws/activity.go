package ws

import (
	"sync/atomic"
	"time"
)

// Activity verfolgt aktive WebSocket-Verbindungen und den Zeitpunkt der letzten
// Aktivität. Damit lässt sich — wie bei Cockpit — der Prozess beenden, sobald
// für eine gewisse Zeit niemand mehr verbunden ist (Socket-Activation startet
// ihn bei der nächsten Verbindung neu).
type Activity struct {
	conns      atomic.Int64
	lastActive atomic.Int64 // Unix-Nanosekunden
}

// NewActivity erzeugt einen Tracker und markiert den Start als „aktiv".
func NewActivity() *Activity {
	a := &Activity{}
	a.Touch()
	return a
}

// Touch aktualisiert den Zeitstempel der letzten Aktivität.
func (a *Activity) Touch() { a.lastActive.Store(time.Now().UnixNano()) }

// Connect/Disconnect zählen aktive Verbindungen.
func (a *Activity) Connect() { a.conns.Add(1); a.Touch() }
func (a *Activity) Disconnect() {
	a.conns.Add(-1)
	a.Touch()
}

// Connections liefert die Anzahl aktiver Verbindungen.
func (a *Activity) Connections() int64 { return a.conns.Load() }

// Idle ist wahr, wenn keine Verbindung besteht und seit der letzten Aktivität
// mindestens d vergangen ist.
func (a *Activity) Idle(d time.Duration) bool {
	if a.conns.Load() > 0 {
		return false
	}
	last := time.Unix(0, a.lastActive.Load())
	return time.Since(last) >= d
}
