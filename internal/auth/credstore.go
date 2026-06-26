package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"sync"
	"time"
)

// CredentialStore hält die Anmeldepasswörter aktiver Sitzungen im Speicher, um
// privilegierte Aktionen per `sudo -S` im Benutzerkontext zu autorisieren
// (Cockpit-Modell). Die Passwörter werden mit einem flüchtigen Prozess-Schlüssel
// AES-GCM-verschlüsselt abgelegt — als Schutz gegen versehentliche Preisgabe in
// Logs/Core-Dumps; der Schlüssel existiert nur im Speicher und überlebt keinen
// Neustart (dann ist ohnehin eine Neuanmeldung nötig).
type CredentialStore struct {
	mu     sync.Mutex
	gcm    cipher.AEAD
	ttl    time.Duration
	values map[string]*encEntry
}

type encEntry struct {
	nonce   []byte
	cipher  []byte
	expires time.Time
}

// NewCredentialStore erzeugt einen Store mit gegebener Lebensdauer.
func NewCredentialStore(ttl time.Duration) *CredentialStore {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	return &CredentialStore{gcm: gcm, ttl: ttl, values: map[string]*encEntry{}}
}

// Put hinterlegt das Passwort einer Sitzung (verschlüsselt).
func (s *CredentialStore) Put(sessionID, password string) {
	nonce := make([]byte, s.gcm.NonceSize())
	_, _ = rand.Read(nonce)
	ct := s.gcm.Seal(nil, nonce, []byte(password), nil)

	s.mu.Lock()
	s.values[sessionID] = &encEntry{nonce: nonce, cipher: ct, expires: time.Now().Add(s.ttl)}
	s.mu.Unlock()
}

// Get liefert das Passwort einer Sitzung (oder false, wenn fehlend/abgelaufen).
func (s *CredentialStore) Get(sessionID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.values[sessionID]
	if !ok {
		return "", false
	}
	if time.Now().After(e.expires) {
		delete(s.values, sessionID)
		return "", false
	}
	pt, err := s.gcm.Open(nil, e.nonce, e.cipher, nil)
	if err != nil {
		return "", false
	}
	return string(pt), true
}

// Delete entfernt die Anmeldedaten einer Sitzung (Logout).
func (s *CredentialStore) Delete(sessionID string) {
	s.mu.Lock()
	delete(s.values, sessionID)
	s.mu.Unlock()
}
