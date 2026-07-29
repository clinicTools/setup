// Package server verdrahtet Router, Middleware, API-Handler und das
// eingebettete Frontend zu einem lauffähigen HTTP-Dienst.
package server

import (
	"net/http"
	"time"

	"github.com/clinictools/setup/internal/api"
	"github.com/clinictools/setup/internal/audit"
	"github.com/clinictools/setup/internal/auth"
	"github.com/clinictools/setup/internal/config"
	"github.com/clinictools/setup/internal/jobs"
	"github.com/clinictools/setup/internal/web"
	"github.com/clinictools/setup/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// New baut den vollständigen HTTP-Handler des Dienstes und liefert zusätzlich
// den Activity-Tracker zurück, über den der Aufrufer (main) einen Cockpit-
// artigen Idle-Shutdown realisieren kann.
func New(cfg *config.Config) (http.Handler, *ws.Activity) {
	authn := auth.NewAuthenticator(cfg.PAMService, cfg.AdminGroups)
	sessions := auth.NewSessionManager(cfg.SessionSecret, cfg.SessionTTL, !cfg.AllowInsecureCookie)
	jobMgr := jobs.NewManager()
	auditLog := audit.New(cfg.AuditLogPath)
	credStore := auth.NewCredentialStore(cfg.SessionTTL)
	a := api.New(authn, sessions, jobMgr, auditLog, credStore)

	activity := ws.NewActivity()
	// a.Runner baut den Benutzerkontext einer Verbindung (für Channels wie journal).
	wsHandler := ws.NewHandler(ws.NewRegistry(jobMgr), activity, a.Runner, cfg.Dev)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	// X-Forwarded-For/X-Real-IP nur hinter vertrauenswürdigem Proxy auswerten,
	// sonst ist die Quell-IP fälschbar (Rate-Limit-Umgehung, Audit-Manipulation).
	if cfg.TrustProxy {
		r.Use(middleware.RealIP)
	}
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)
	// Jede HTTP-Aktivität verlängert das Idle-Fenster.
	r.Use(activityMiddleware(activity))

	if cfg.Dev {
		r.Use(devCORS)
	}

	r.Route("/api", func(r chi.Router) {
		// Öffentliche Auth-Endpunkte.
		r.Post("/auth/login", a.Login)
		r.Post("/auth/logout", a.Logout)
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Geschützte Endpunkte (gültige Session erforderlich).
		r.Group(func(r chi.Router) {
			r.Use(sessions.Middleware)

			// Live-WebSocket (gemultiplexte Channels). KEIN Request-Timeout,
			// da die Verbindung langlebig ist.
			r.Get("/ws", wsHandler.ServeHTTP)

			// REST-Endpunkte mit 60-Sekunden-Timeout.
			r.Group(func(r chi.Router) {
				r.Use(middleware.Timeout(60 * time.Second))
				r.Get("/auth/me", a.Me)

				// Lesende System-Endpunkte.
				r.Get("/system/info", a.Info)
				r.Get("/system/users", a.Users)
				r.Get("/system/groups", a.Groups)
				r.Get("/system/services", a.Services)
				r.Get("/system/services/{name}/status", a.ServiceStatus)
				r.Get("/system/packages", a.Packages)
				r.Get("/system/packages/installed", a.InstalledPackages)
				r.Get("/system/network", a.Network)
				r.Get("/system/storage", a.Storage)
				r.Get("/system/firewall", a.Firewall)
				r.Get("/system/logs", a.Logs)
				r.Get("/system/scheduled", a.Scheduled)
				r.Get("/system/time", a.Time)
				r.Get("/system/timezones", a.Timezones)
				r.Get("/system/processes", a.Processes)
				r.Get("/system/audit", a.AuditList)
				r.Get("/system/jobs", a.JobsList)
				r.Get("/system/jobs/{id}", a.JobGet)
				// Podman (Lesen)
				r.Get("/system/podman/status", a.PodmanStatus)
				r.Get("/system/podman/containers", a.Containers)
				r.Get("/system/podman/containers/{id}/logs", a.ContainerLogs)
				r.Get("/system/podman/images", a.ContainerImages)
				r.Get("/system/podman/volumes", a.ContainerVolumes)

				// Schreibende Endpunkte: Admin-Gruppe + CSRF + Audit.
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Use(auth.CSRFMiddleware)
					r.Use(a.AuditMiddleware)

					// Dienste
					r.Post("/system/services/{name}/action", a.ServiceAction)
					// Benutzer & Gruppen
					r.Post("/system/users", a.CreateUser)
					r.Delete("/system/users/{name}", a.DeleteUser)
					r.Put("/system/users/{name}/password", a.SetPassword)
					r.Put("/system/users/{name}", a.ModifyUser)
					r.Post("/system/groups", a.CreateGroup)
					r.Delete("/system/groups/{name}", a.DeleteGroup)
					// Firewall
					r.Put("/system/firewall", a.FirewallSet)
					r.Post("/system/firewall/rules", a.FirewallAddRule)
					r.Delete("/system/firewall/rules", a.FirewallDeleteRule)
					// Zeit
					r.Put("/system/time/timezone", a.SetTimezone)
					r.Put("/system/time/ntp", a.SetNTP)
					// Pakete (langlaufend → Jobs)
					r.Post("/system/packages/update", a.AptUpdate)
					r.Post("/system/packages/install", a.PackageInstall)
					r.Post("/system/packages/remove", a.PackageRemove)
					r.Post("/system/packages/upgrade", a.PackageUpgrade)
					// Prozesse
					r.Post("/system/processes/{pid}/kill", a.KillProcess)
					// Cron
					r.Post("/system/cron", a.CreateCron)
					r.Delete("/system/cron/{name}", a.DeleteCron)
					// Netzwerk
					r.Put("/system/network/hostname", a.SetHostname)
					r.Put("/system/network/interfaces/{iface}", a.SetInterfaceState)
					// Jobs
					r.Post("/system/jobs/{id}/cancel", a.JobCancel)
					// Podman & Compose-Stacks (compose.yaml kann Zugangsdaten
					// enthalten → auch das Lesen ist Admins vorbehalten)
					r.Post("/system/podman/install", a.PodmanInstall)
					r.Post("/system/podman/containers/{id}/action", a.ContainerAction)
					r.Get("/system/podman/stacks", a.Stacks)
					r.Get("/system/podman/stacks/{name}", a.StackGet)
					r.Put("/system/podman/stacks/{name}", a.StackWrite)
					r.Delete("/system/podman/stacks/{name}", a.StackDelete)
					r.Post("/system/podman/stacks/{name}/validate", a.StackValidate)
					r.Post("/system/podman/stacks/{name}/{action}", a.StackAction)
					// Energie
					r.Post("/system/power", a.Power)
				})
			})
		})
	})

	// Frontend (SPA) als Fallback für alle übrigen Pfade.
	r.NotFound(web.SPAHandler().ServeHTTP)

	return r, activity
}

// activityMiddleware markiert jede HTTP-Anfrage als Aktivität, damit der
// Idle-Shutdown nur bei tatsächlicher Inaktivität greift.
func activityMiddleware(a *ws.Activity) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.Touch()
			next.ServeHTTP(w, r)
		})
	}
}

// securityHeaders setzt defensive HTTP-Header analog zur KIS-Oberfläche.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// devCORS erlaubt im Entwicklungsmodus Requests des Vite-Dev-Servers.
func devCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "http://localhost:5173")
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
