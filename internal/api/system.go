package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/clinictools/setup/internal/system"
	"github.com/go-chi/chi/v5"
)

// --- Lese-Endpunkte (Dashboard & Module) ---

func (a *API) Info(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Info() })
}

func (a *API) Users(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Users() })
}

func (a *API) Groups(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Groups() })
}

func (a *API) Services(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Services() })
}

func (a *API) ServiceStatus(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	handle(w, func() (any, error) {
		out, err := system.ServiceStatus(name)
		return map[string]string{"status": out}, err
	})
}

func (a *API) Packages(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Packages() })
}

func (a *API) InstalledPackages(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.InstalledPackages() })
}

func (a *API) Network(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Network() })
}

func (a *API) Storage(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Storage() })
}

func (a *API) Firewall(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Firewall() })
}

func (a *API) Logs(w http.ResponseWriter, r *http.Request) {
	q := system.LogQuery{
		Unit:     r.URL.Query().Get("unit"),
		Priority: r.URL.Query().Get("priority"),
		Lines:    atoiDefault(r.URL.Query().Get("lines"), 200),
	}
	handle(w, func() (any, error) { return system.Logs(q) })
}

func (a *API) Scheduled(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Scheduled() })
}

func (a *API) Time(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.Time() })
}

func (a *API) Timezones(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.ListTimezones() })
}

func (a *API) Processes(w http.ResponseWriter, r *http.Request) {
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	handle(w, func() (any, error) { return system.Processes(limit) })
}

// --- Schreib-/Aktions-Endpunkte (erfordern Admin-Gruppe) ---

type serviceActionRequest struct {
	Action string `json:"action"`
}

func (a *API) ServiceAction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req serviceActionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.ServiceAction(name, req.Action); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req system.CreateUserRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.CreateUser(req); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	removeHome := r.URL.Query().Get("removeHome") == "true"
	if err := system.DeleteUser(name, removeHome); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type passwordRequest struct {
	Password string `json:"password"`
}

func (a *API) SetPassword(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req passwordRequest
	if err := decode(r, &req); err != nil || req.Password == "" {
		writeError(w, http.StatusBadRequest, errors.New("Passwort erforderlich"))
		return
	}
	if err := system.SetPassword(name, req.Password); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type enabledRequest struct {
	Enabled bool `json:"enabled"`
}

func (a *API) FirewallSet(w http.ResponseWriter, r *http.Request) {
	var req enabledRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.FirewallSetEnabled(req.Enabled); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type timezoneRequest struct {
	Timezone string `json:"timezone"`
}

func (a *API) SetTimezone(w http.ResponseWriter, r *http.Request) {
	var req timezoneRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.SetTimezone(req.Timezone); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) SetNTP(w http.ResponseWriter, r *http.Request) {
	var req enabledRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.SetNTP(req.Enabled); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) AptUpdate(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) {
		out, err := system.AptUpdate()
		return map[string]string{"output": out}, err
	})
}

type powerRequest struct {
	Action string `json:"action"`
}

func (a *API) Power(w http.ResponseWriter, r *http.Request) {
	var req powerRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.PowerAction(req.Action); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- Firewall-Regeln ---

type firewallRuleRequest struct {
	Action   string `json:"action"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
}

func (a *API) FirewallAddRule(w http.ResponseWriter, r *http.Request) {
	var req firewallRuleRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.FirewallAddRule(req.Action, req.Port, req.Protocol); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) FirewallDeleteRule(w http.ResponseWriter, r *http.Request) {
	var req firewallRuleRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.FirewallDeleteRule(req.Action, req.Port, req.Protocol); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- Gruppen & Benutzerbearbeitung ---

type groupRequest struct {
	Name   string `json:"name"`
	System bool   `json:"system"`
}

func (a *API) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req groupRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.CreateGroup(req.Name, req.System); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (a *API) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if err := system.DeleteGroup(chi.URLParam(r, "name")); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type modifyUserRequest struct {
	Groups []string `json:"groups"`
	Shell  string   `json:"shell"`
}

func (a *API) ModifyUser(w http.ResponseWriter, r *http.Request) {
	var req modifyUserRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.ModifyUser(chi.URLParam(r, "name"), req.Groups, req.Shell); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- Prozesse ---

type killRequest struct {
	Signal string `json:"signal"`
}

func (a *API) KillProcess(w http.ResponseWriter, r *http.Request) {
	pid := atoiDefault(chi.URLParam(r, "pid"), 0)
	var req killRequest
	_ = decode(r, &req)
	if err := system.KillProcess(pid, req.Signal); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- Cron ---

type cronRequest struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	User     string `json:"user"`
	Command  string `json:"command"`
}

func (a *API) CreateCron(w http.ResponseWriter, r *http.Request) {
	var req cronRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.CreateCronJob(req.Name, req.Schedule, req.User, req.Command); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (a *API) DeleteCron(w http.ResponseWriter, r *http.Request) {
	if err := system.DeleteCronJob(chi.URLParam(r, "name")); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- Netzwerk ---

type hostnameRequest struct {
	Hostname string `json:"hostname"`
}

func (a *API) SetHostname(w http.ResponseWriter, r *http.Request) {
	var req hostnameRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.SetHostname(req.Hostname); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type ifaceStateRequest struct {
	Up bool `json:"up"`
}

func (a *API) SetInterfaceState(w http.ResponseWriter, r *http.Request) {
	var req ifaceStateRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.SetInterfaceState(chi.URLParam(r, "iface"), req.Up); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
