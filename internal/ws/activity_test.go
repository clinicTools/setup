package ws

import (
	"testing"
	"time"
)

func TestActivityIdle(t *testing.T) {
	a := NewActivity()

	// Aktive Verbindung verhindert Idle.
	a.Connect()
	if a.Idle(time.Nanosecond) {
		t.Error("mit aktiver Verbindung darf nicht idle sein")
	}
	if a.Connections() != 1 {
		t.Errorf("Connections = %d, erwartet 1", a.Connections())
	}

	// Nach Trennung ist der Tracker je nach Timeout idle.
	a.Disconnect()
	if a.Connections() != 0 {
		t.Errorf("Connections = %d, erwartet 0", a.Connections())
	}
	time.Sleep(2 * time.Millisecond)
	if !a.Idle(time.Millisecond) {
		t.Error("ohne Verbindung und nach Ablauf sollte idle sein")
	}

	// Großzügiges Fenster: noch nicht idle.
	if a.Idle(time.Hour) {
		t.Error("innerhalb des Fensters darf nicht idle sein")
	}

	// Touch verschiebt das Fenster.
	a.Touch()
	if a.Idle(time.Minute) {
		t.Error("nach Touch darf nicht sofort idle sein")
	}
}
