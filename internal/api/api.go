// Package api stellt die HTTP-Handler der REST-Schnittstelle bereit. Alle
// Antworten sind JSON; Fehler folgen einem einheitlichen Schema {"error": …}.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/clinictools/setup/internal/auth"
	"github.com/clinictools/setup/internal/system"
)

// API bündelt die Abhängigkeiten der Handler.
type API struct {
	Auth     *auth.Authenticator
	Sessions *auth.SessionManager
}

// New erzeugt eine API-Instanz.
func New(a *auth.Authenticator, s *auth.SessionManager) *API {
	return &API{Auth: a, Sessions: s}
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
		var missing *system.ErrCommandFailed
		if errors.As(err, &missing) {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// decode liest einen JSON-Request-Body in v.
func decode(r *http.Request, v any) error {
	defer func() { _ = r.Body.Close() }()
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
