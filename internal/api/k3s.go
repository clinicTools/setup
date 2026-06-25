package api

import (
	"errors"
	"net/http"

	"github.com/clinictools/setup/internal/system"
)

// K3sStatus liefert Installations-/Laufzustand von k3s.
func (a *API) K3sStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, system.K3sGetStatus())
}

// K3sNodes liefert die Cluster-Knoten.
func (a *API) K3sNodes(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.K3sNodes() })
}

// K3sPods liefert die Pods aller Namespaces.
func (a *API) K3sPods(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) { return system.K3sPods() })
}

// K3sKubeconfig liefert die kubeconfig im Klartext.
func (a *API) K3sKubeconfig(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) {
		cfg, err := system.K3sKubeconfig()
		return map[string]string{"kubeconfig": cfg}, err
	})
}

// K3sToken liefert den Join-Token für Agent-Knoten.
func (a *API) K3sToken(w http.ResponseWriter, _ *http.Request) {
	handle(w, func() (any, error) {
		token, err := system.K3sNodeToken()
		return map[string]string{"token": token}, err
	})
}

// k3sInstallRequest steuert die Installationsoptionen.
type k3sInstallRequest struct {
	// DisableTraefik lässt den mitgelieferten Ingress-Controller weg.
	DisableTraefik bool `json:"disableTraefik"`
	// WriteKubeconfigMode setzt die Dateirechte der kubeconfig (z. B. "644").
	WriteKubeconfigMode string `json:"writeKubeconfigMode"`
}

// K3sInstall startet die k3s-Installation als Live-Job (offizieller Installer).
func (a *API) K3sInstall(w http.ResponseWriter, r *http.Request) {
	var req k3sInstallRequest
	_ = decode(r, &req)

	// INSTALL_K3S_EXEC bündelt die Serveroptionen; nur feste, validierte Flags.
	execArgs := "server"
	if req.DisableTraefik {
		execArgs += " --disable=traefik"
	}
	mode := "644"
	if req.WriteKubeconfigMode == "600" || req.WriteKubeconfigMode == "640" {
		mode = req.WriteKubeconfigMode
	}

	// Server-definierte Befehlszeile (keine Client-Strings) — Pipe erfordert sh.
	script := "curl -sfL https://get.k3s.io | " +
		"INSTALL_K3S_EXEC='" + execArgs + "' " +
		"K3S_KUBECONFIG_MODE='" + mode + "' sh -"
	job := a.Jobs.Start("k3s installieren", "sh", []string{"-c", script})
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

// K3sUninstall startet das k3s-Deinstallationsskript als Job.
func (a *API) K3sUninstall(w http.ResponseWriter, _ *http.Request) {
	if !system.K3sInstalled() {
		writeError(w, http.StatusBadRequest, errors.New("k3s ist nicht installiert"))
		return
	}
	job := a.Jobs.Start("k3s deinstallieren", "/usr/local/bin/k3s-uninstall.sh", nil)
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}
