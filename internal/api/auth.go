package api

import (
	"errors"
	"net/http"

	"github.com/clinictools/setup/internal/auth"
)

// loginRequest ist der Body des Login-Endpunkts.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login authentifiziert gegen das Betriebssystem (PAM) und setzt bei Erfolg
// das Session-Cookie.
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, errors.New("Benutzername und Passwort erforderlich"))
		return
	}

	user, err := a.Auth.Authenticate(req.Username, req.Password)
	if err != nil {
		// Einheitliche Fehlermeldung – keine Auskunft über Existenz des Kontos.
		writeError(w, http.StatusUnauthorized, auth.ErrInvalidCredentials)
		return
	}

	if err := a.Sessions.Issue(w, user); err != nil {
		writeError(w, http.StatusInternalServerError, errors.New("Session konnte nicht erstellt werden"))
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// Logout löscht das Session-Cookie.
func (a *API) Logout(w http.ResponseWriter, _ *http.Request) {
	a.Sessions.Clear(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Me liefert das angereicherte Profil des angemeldeten Benutzers.
func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, errors.New("nicht authentifiziert"))
		return
	}
	// Gruppen/Details frisch aus dem System nachladen (Cookie kann veralten).
	if enriched, err := a.Auth.Lookup(u.Username); err == nil {
		writeJSON(w, http.StatusOK, enriched)
		return
	}
	writeJSON(w, http.StatusOK, u)
}
