package auth

import (
	"sync"
	"time"
)

// RateLimiter bremst Brute-Force-Anmeldeversuche: Nach max Fehlversuchen
// innerhalb von window wird der Schlüssel (z. B. Benutzer+IP) für block
// gesperrt. Erfolgreiche Anmeldungen setzen den Zähler zurück.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attemptInfo
	max      int
	window   time.Duration
	block    time.Duration
}

type attemptInfo struct {
	count        int
	first        time.Time
	blockedUntil time.Time
}

// NewRateLimiter erzeugt einen Limiter (z. B. 5 Versuche / 5 min, 15 min Sperre).
func NewRateLimiter(max int, window, block time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: map[string]*attemptInfo{},
		max:      max,
		window:   window,
		block:    block,
	}
}

// Allowed prüft, ob ein Versuch erlaubt ist; liefert ggf. die Restsperrzeit.
func (r *RateLimiter) Allowed(key string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := r.attempts[key]
	if info == nil {
		return true, 0
	}
	now := time.Now()
	if now.Before(info.blockedUntil) {
		return false, time.Until(info.blockedUntil)
	}
	// Fenster abgelaufen → Zähler verfällt.
	if now.Sub(info.first) > r.window {
		delete(r.attempts, key)
	}
	return true, 0
}

// Fail registriert einen Fehlversuch und sperrt bei Überschreitung.
func (r *RateLimiter) Fail(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	info := r.attempts[key]
	if info == nil || now.Sub(info.first) > r.window {
		info = &attemptInfo{first: now}
		r.attempts[key] = info
	}
	info.count++
	if info.count >= r.max {
		info.blockedUntil = now.Add(r.block)
	}
}

// Reset löscht den Zähler nach erfolgreicher Anmeldung.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, key)
}
