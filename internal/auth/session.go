package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// cookieName ist der Name des Session-Cookies.
const cookieName = "da_session"

// SessionManager signiert und verifiziert Session-Tokens (JWT/HS256) und
// verwaltet das zugehörige HTTP-Cookie.
type SessionManager struct {
	secret         []byte
	ttl            time.Duration
	secureCookie   bool
}

// claims ist die JWT-Nutzlast einer Session.
type claims struct {
	Username  string   `json:"username"`
	UID       int      `json:"uid"`
	GID       int      `json:"gid"`
	Admin     bool     `json:"admin"`
	Groups    []string `json:"groups"`
	GIDs      []int    `json:"gids"`
	SessionID string   `json:"sid"`
	jwt.RegisteredClaims
}

// NewSessionManager erzeugt einen SessionManager. secureCookie sollte in
// Produktion (HTTPS) true sein.
func NewSessionManager(secret []byte, ttl time.Duration, secureCookie bool) *SessionManager {
	return &SessionManager{secret: secret, ttl: ttl, secureCookie: secureCookie}
}

// Issue setzt ein signiertes Session-Cookie für den Benutzer. u.SessionID muss
// gesetzt sein (verknüpft die Session mit dem Credential-Store).
func (m *SessionManager) Issue(w http.ResponseWriter, u *User) error {
	now := time.Now()
	c := claims{
		Username:  u.Username,
		UID:       u.UID,
		GID:       u.GID,
		Admin:     u.Admin,
		Groups:    u.Groups,
		GIDs:      u.GIDs,
		SessionID: u.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  now.Add(m.ttl),
		MaxAge:   int(m.ttl.Seconds()),
	})
	return nil
}

// Clear löscht das Session-Cookie (Logout).
func (m *SessionManager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// Verify liest und validiert das Session-Cookie eines Requests.
func (m *SessionManager) Verify(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, errors.New("keine Session vorhanden")
	}

	var c claims
	token, err := jwt.ParseWithClaims(cookie.Value, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unerwartete Signatur-Methode")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("ungültige Session")
	}

	return &User{
		Username:  c.Username,
		UID:       c.UID,
		GID:       c.GID,
		Admin:     c.Admin,
		Groups:    c.Groups,
		GIDs:      c.GIDs,
		SessionID: c.SessionID,
	}, nil
}
