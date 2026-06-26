package auth

import (
	"testing"
	"time"
)

func TestCredentialStoreRoundTrip(t *testing.T) {
	s := NewCredentialStore(time.Minute)
	s.Put("sess-1", "geheim")

	pw, ok := s.Get("sess-1")
	if !ok || pw != "geheim" {
		t.Fatalf("Get = %q, %v; erwartet geheim, true", pw, ok)
	}

	s.Delete("sess-1")
	if _, ok := s.Get("sess-1"); ok {
		t.Error("nach Delete sollte kein Eintrag mehr existieren")
	}
}

func TestCredentialStoreExpiry(t *testing.T) {
	s := NewCredentialStore(time.Millisecond)
	s.Put("sess", "pw")
	time.Sleep(3 * time.Millisecond)
	if _, ok := s.Get("sess"); ok {
		t.Error("abgelaufener Eintrag sollte nicht mehr geliefert werden")
	}
}

func TestCredentialStoreEncryptsAtRest(t *testing.T) {
	s := NewCredentialStore(time.Minute)
	s.Put("sess", "supersecret")
	// Das Klartext-Passwort darf nicht im gespeicherten Chiffrat auftauchen.
	e := s.values["sess"]
	if e == nil {
		t.Fatal("Eintrag fehlt")
	}
	if string(e.cipher) == "supersecret" || contains(e.cipher, "supersecret") {
		t.Error("Passwort liegt im Klartext vor")
	}
}

func contains(haystack []byte, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		string(haystack) != "" && indexOf(haystack, needle) >= 0
}

func indexOf(h []byte, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if string(h[i:i+len(n)]) == n {
			return i
		}
	}
	return -1
}
