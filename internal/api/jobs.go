package api

import (
	"errors"
	"net/http"

	"github.com/clinictools/setup/internal/system"
	"github.com/go-chi/chi/v5"
)

// Jobs listet alle bekannten Jobs.
func (a *API) JobsList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.Jobs.List())
}

// JobGet liefert einen einzelnen Job.
func (a *API) JobGet(w http.ResponseWriter, r *http.Request) {
	job, ok := a.Jobs.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("Job nicht gefunden"))
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// JobCancel bricht einen laufenden Job ab.
func (a *API) JobCancel(w http.ResponseWriter, r *http.Request) {
	job, ok := a.Jobs.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("Job nicht gefunden"))
		return
	}
	job.Cancel()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// packagesRequest ist der Body für Paket-Installation/-Entfernung.
type packagesRequest struct {
	Packages []string `json:"packages"`
}

// PackageInstall startet `apt-get install` als Job.
func (a *API) PackageInstall(w http.ResponseWriter, r *http.Request) {
	a.aptPackageJob(w, r, "install", "Pakete installieren")
}

// PackageRemove startet `apt-get remove` als Job.
func (a *API) PackageRemove(w http.ResponseWriter, r *http.Request) {
	a.aptPackageJob(w, r, "remove", "Pakete entfernen")
}

// aptPackageJob validiert Paketnamen und startet das jeweilige apt-Kommando.
func (a *API) aptPackageJob(w http.ResponseWriter, r *http.Request, action, name string) {
	var req packagesRequest
	if err := decode(r, &req); err != nil || len(req.Packages) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("keine Pakete angegeben"))
		return
	}
	for _, p := range req.Packages {
		if !system.ValidPackageName(p) {
			writeError(w, http.StatusBadRequest, errors.New("ungültiger Paketname: "+p))
			return
		}
	}
	args := append([]string{"-y", action}, req.Packages...)
	job := a.Jobs.StartEnv(name, "apt-get", args, aptEnv)
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

// PackageUpgrade startet ein vollständiges System-Upgrade als Job.
func (a *API) PackageUpgrade(w http.ResponseWriter, _ *http.Request) {
	job := a.Jobs.StartEnv("System aktualisieren", "apt-get", []string{"-y", "upgrade"}, aptEnv)
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

// aptEnv sorgt für nicht-interaktive apt-Läufe (keine Debconf-Prompts).
var aptEnv = []string{"DEBIAN_FRONTEND=noninteractive"}
