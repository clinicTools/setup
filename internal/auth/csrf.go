package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const (
	csrfCookie = "da_csrf"
	csrfHeader = "X-CSRF-Token"
)

// IssueCSRF setzt ein nicht-HttpOnly-CSRF-Cookie (Double-Submit-Muster). Das
// Frontend liest es aus und sendet den Wert bei mutierenden Anfragen als Header.
func (m *SessionManager) IssueCSRF(w http.ResponseWriter) {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // muss für JS lesbar sein
		Secure:   m.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.ttl.Seconds()),
	})
}

// ClearCSRF entfernt das CSRF-Cookie.
func (m *SessionManager) ClearCSRF(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: csrfCookie, Value: "", Path: "/", MaxAge: -1,
		Secure: m.secureCookie, SameSite: http.SameSiteLaxMode,
	})
}

// CSRFMiddleware erzwingt bei zustandsändernden Methoden, dass Header- und
// Cookie-Token übereinstimmen (Double-Submit-Cookie gegen CSRF).
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(csrfCookie)
		header := r.Header.Get(csrfHeader)
		if err != nil || header == "" ||
			subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
			http.Error(w, `{"error":"CSRF-Token fehlt oder ungültig"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
