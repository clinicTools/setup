package auth

import (
	"context"
	"net/http"
)

type ctxKey int

const userCtxKey ctxKey = 0

// Middleware schützt nachgelagerte Handler: ohne gültige Session wird mit
// 401 abgebrochen, andernfalls landet der Benutzer im Request-Context.
func (m *SessionManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := m.Verify(r)
		if err != nil {
			http.Error(w, `{"error":"nicht authentifiziert"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin erzwingt, dass der angemeldete Benutzer einer privilegierten
// Gruppe angehört. Schreibende Operationen werden damit abgesichert.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil || !u.Admin {
			http.Error(w, `{"error":"unzureichende Berechtigung"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserFromContext liefert den authentifizierten Benutzer aus dem Context.
func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userCtxKey).(*User)
	return u
}
