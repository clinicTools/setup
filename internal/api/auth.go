package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/clinictools/setup/internal/audit"
	"github.com/clinictools/setup/internal/auth"
	"github.com/clinictools/setup/internal/system"
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

	ip := clientIP(r)
	key := req.Username + "|" + ip

	// Brute-Force-Schutz.
	if ok, retry := a.Limiter.Allowed(key); !ok {
		a.Audit.Record(audit.Entry{User: req.Username, Action: "login", Success: false,
			Detail: "gesperrt (Rate-Limit)", IP: ip})
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retry.Seconds())))
		writeError(w, http.StatusTooManyRequests,
			fmt.Errorf("zu viele Fehlversuche, bitte in %d s erneut versuchen", int(retry.Seconds())))
		return
	}

	user, err := a.Auth.Authenticate(req.Username, req.Password)
	if err != nil {
		a.Limiter.Fail(key)
		a.Audit.Record(audit.Entry{User: req.Username, Action: "login", Success: false,
			Detail: "ungültige Anmeldedaten", IP: ip})
		// Einheitliche Fehlermeldung – keine Auskunft über Existenz des Kontos.
		writeError(w, http.StatusUnauthorized, auth.ErrInvalidCredentials)
		return
	}

	a.Limiter.Reset(key)

	// Session-ID erzeugen und das Passwort für die Dauer der Sitzung im
	// Credential-Store hinterlegen — es autorisiert spätere privilegierte
	// Aktionen via sudo im Benutzerkontext (Cockpit-Modell).
	user.SessionID = newSessionID()
	a.Creds.Put(user.SessionID, req.Password)

	// Admin-Status aus der tatsächlichen sudo-Berechtigung ableiten (OS als
	// Autorität), nicht nur aus der Gruppenzugehörigkeit. Nur möglich, wenn der
	// Dienst als root läuft und damit in den Benutzerkontext wechseln kann.
	if os.Geteuid() == 0 {
		runner := system.NewRunner(user.UID, user.GID, user.GIDs, user.Username, user.HomeDir, req.Password)
		user.Admin = runner.CanEscalate()
	}

	if err := a.Sessions.Issue(w, user); err != nil {
		a.Creds.Delete(user.SessionID)
		writeError(w, http.StatusInternalServerError, errors.New("Session konnte nicht erstellt werden"))
		return
	}
	a.Sessions.IssueCSRF(w)
	a.Audit.Record(audit.Entry{User: user.Username, Action: "login", Success: true, IP: ip})
	writeJSON(w, http.StatusOK, user)
}

// Logout löscht Session- und CSRF-Cookie und verwirft die Anmeldedaten.
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	if u, err := a.Sessions.Verify(r); err == nil && u.SessionID != "" {
		a.Creds.Delete(u.SessionID)
	}
	a.Sessions.Clear(w)
	a.Sessions.ClearCSRF(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// newSessionID erzeugt eine zufällige Session-Kennung.
func newSessionID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// clientIP extrahiert die Quell-IP (RemoteAddr ohne Port).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
