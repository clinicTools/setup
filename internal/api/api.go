// Package api stellt die HTTP-Handler der REST-Schnittstelle bereit. Alle
// Antworten sind JSON; Fehler folgen einem einheitlichen Schema {"error": …}.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/clinictools/setup/internal/audit"
	"github.com/clinictools/setup/internal/auth"
	"github.com/clinictools/setup/internal/jobs"
	"github.com/clinictools/setup/internal/system"
)

// API bündelt die Abhängigkeiten der Handler.
type API struct {
	Auth     *auth.Authenticator
	Sessions *auth.SessionManager
	Jobs     *jobs.Manager
	Audit    *audit.Logger
	Limiter  *auth.RateLimiter
	Creds    *auth.CredentialStore
}

// New erzeugt eine API-Instanz. Der Login-Limiter erlaubt 5 Fehlversuche je
// 5 Minuten und sperrt danach 15 Minuten.
func New(a *auth.Authenticator, s *auth.SessionManager, jm *jobs.Manager, al *audit.Logger, cs *auth.CredentialStore) *API {
	return &API{
		Auth:     a,
		Sessions: s,
		Jobs:     jm,
		Audit:    al,
		Limiter:  auth.NewRateLimiter(5, 5*time.Minute, 15*time.Minute),
		Creds:    cs,
	}
}

// Runner baut den Ausführungskontext einer Anfrage: Läuft der Dienst als root,
// werden Kommandos unter UID/GID des angemeldeten Benutzers ausgeführt
// (privilegierte via sudo mit dem in der Session gehaltenen Passwort). Läuft der
// Dienst nicht als root (Entwicklung), wird direkt als aktueller Prozess
// ausgeführt.
func (a *API) Runner(r *http.Request) *system.Runner {
	u := auth.UserFromContext(r.Context())
	if u == nil || os.Geteuid() != 0 {
		return system.RootRunner()
	}
	password, _ := a.Creds.Get(u.SessionID)
	return system.NewRunner(u.UID, u.GID, u.GIDs, u.Username, u.HomeDir, password)
}

// writeJSON serialisiert v als JSON mit dem angegebenen Statuscode.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError mappt Fehler auf einen Statuscode und liefert {"error": msg}.
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// handle reduziert Boilerplate: ein Collector liefert (data, error); Fehler
// werden als 500 abgebildet, Erfolg als 200.
func handle(w http.ResponseWriter, fn func() (any, error)) {
	data, err := fn()
	if err != nil {
		// Eingabefehler des Aufrufers → 400, fehlgeschlagene Systemkommandos → 502.
		var invalid *system.ErrInvalidInput
		if errors.As(err, &invalid) {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		var failed *system.ErrCommandFailed
		if errors.As(err, &failed) {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// decode liest einen JSON-Request-Body in v (begrenzt auf 1 MiB gegen DoS).
func decode(r *http.Request, v any) error {
	defer func() { _ = r.Body.Close() }()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
