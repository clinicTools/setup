package auth

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterMax(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute, time.Hour)
	key := "user|1.2.3.4"

	for i := 0; i < 2; i++ {
		if ok, _ := rl.Allowed(key); !ok {
			t.Fatalf("Versuch %d sollte erlaubt sein", i)
		}
		rl.Fail(key)
	}
	// Dritter Fehlversuch erreicht das Limit → Sperre.
	rl.Fail(key)
	if ok, retry := rl.Allowed(key); ok || retry <= 0 {
		t.Errorf("nach max Fehlversuchen erwartet gesperrt, got ok=%v retry=%v", ok, retry)
	}
}

func TestRateLimiterResetOnSuccess(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute, time.Hour)
	key := "u|ip"
	rl.Fail(key)
	rl.Reset(key)
	if ok, _ := rl.Allowed(key); !ok {
		t.Error("nach Reset sollte wieder erlaubt sein")
	}
}
