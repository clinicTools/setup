package api

import (
	"fmt"
	"net/http"

	"github.com/clinictools/setup/internal/audit"
	"github.com/clinictools/setup/internal/auth"
)

// statusRecorder fängt den HTTP-Statuscode ab, um Erfolg/Fehler zu protokollieren.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// AuditMiddleware protokolliert jede (schreibende) Anfrage zentral: Benutzer,
// Methode/Pfad, Ergebnis und Quell-IP.
func (a *API) AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		user := "?"
		if u := auth.UserFromContext(r.Context()); u != nil {
			user = u.Username
		}
		a.Audit.Record(audit.Entry{
			User:    user,
			Action:  r.Method + " " + r.URL.Path,
			Success: rec.status < 400,
			Detail:  fmt.Sprintf("HTTP %d", rec.status),
			IP:      r.RemoteAddr,
		})
	})
}

// AuditList liefert die jüngsten Audit-Einträge (nur Admins).
func (a *API) AuditList(w http.ResponseWriter, r *http.Request) {
	limit := atoiDefault(r.URL.Query().Get("limit"), 200)
	writeJSON(w, http.StatusOK, a.Audit.List(limit))
}
