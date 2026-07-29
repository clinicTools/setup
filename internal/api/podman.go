package api

import (
	"context"
	"errors"
	"net/http"
	"os/exec"

	"github.com/clinictools/setup/internal/system"
	"github.com/go-chi/chi/v5"
)

// --- Lesende Endpunkte ---

// PodmanStatus liefert Verfügbarkeit und Eckdaten der Podman-Installation.
func (a *API) PodmanStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, system.PodmanGetStatus(a.Runner(r)))
}

// Containers listet alle Container (inkl. gestoppter).
func (a *API) Containers(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	handle(w, func() (any, error) { return system.Containers(runner, true) })
}

// ContainerImages listet lokale Images.
func (a *API) ContainerImages(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	handle(w, func() (any, error) { return system.ContainerImages(runner) })
}

// ContainerVolumes listet Podman-Volumes.
func (a *API) ContainerVolumes(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	handle(w, func() (any, error) { return system.ContainerVolumes(runner) })
}

// ContainerLogs liefert die jüngsten Logzeilen eines Containers.
func (a *API) ContainerLogs(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	id := chi.URLParam(r, "id")
	lines := atoiDefault(r.URL.Query().Get("lines"), 200)
	handle(w, func() (any, error) {
		out, err := system.ContainerLogs(runner, id, lines)
		return map[string]string{"logs": out}, err
	})
}

// --- Container-Aktionen ---

type containerActionRequest struct {
	Action string `json:"action"`
}

// ContainerAction führt start/stop/restart/kill/pause/unpause/rm aus.
func (a *API) ContainerAction(w http.ResponseWriter, r *http.Request) {
	var req containerActionRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.ContainerAction(a.Runner(r), chi.URLParam(r, "id"), req.Action); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PodmanInstall installiert Podman samt Compose-Unterstützung als Live-Job.
func (a *API) PodmanInstall(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	job := a.Jobs.Start("Podman installieren", func(ctx context.Context) *exec.Cmd {
		// Zusätzlich den API-Socket aktivieren: Wird später „podman compose"
		// (Delegation an docker-compose) verwendet, ist er zwingend nötig.
		script := "apt-get install -y podman podman-compose && " +
			"(systemctl enable --now podman.socket || true)"
		cmd := runner.BuildCommand(ctx, true, "sh", "-c", script)
		cmd.Env = append(cmd.Environ(), "DEBIAN_FRONTEND=noninteractive")
		return cmd
	})
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

// --- Compose-Stacks ---

// Stacks listet die verwalteten Compose-Projekte.
func (a *API) Stacks(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	handle(w, func() (any, error) { return system.Stacks(runner) })
}

// StackGet liefert den Inhalt der compose.yaml eines Stacks.
func (a *API) StackGet(w http.ResponseWriter, r *http.Request) {
	runner := a.Runner(r)
	name := chi.URLParam(r, "name")
	handle(w, func() (any, error) {
		content, err := system.StackRead(runner, name)
		return map[string]string{"name": name, "compose": content}, err
	})
}

type stackWriteRequest struct {
	Compose string `json:"compose"`
}

// StackWrite legt einen Stack an bzw. aktualisiert seine compose.yaml.
func (a *API) StackWrite(w http.ResponseWriter, r *http.Request) {
	var req stackWriteRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("ungültige Anfrage"))
		return
	}
	if err := system.StackWrite(a.Runner(r), chi.URLParam(r, "name"), req.Compose); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// StackDelete entfernt ein Stack-Verzeichnis.
func (a *API) StackDelete(w http.ResponseWriter, r *http.Request) {
	if err := system.StackDelete(a.Runner(r), chi.URLParam(r, "name")); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// StackAction startet eine Compose-Aktion (up/down/restart/pull/…) als Job,
// dessen Ausgabe live über den WebSocket-Channel „jobs" gestreamt wird.
func (a *API) StackAction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	action := chi.URLParam(r, "action")

	cmdName, args, err := system.StackComposeArgs(name, action)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	runner := a.Runner(r)
	label := map[string]string{
		"up": "Stack starten", "down": "Stack stoppen", "restart": "Stack neu starten",
		"pull": "Images aktualisieren", "stop": "Stack anhalten", "start": "Stack fortsetzen",
	}[action]

	job := a.Jobs.Start(label+": "+name, func(ctx context.Context) *exec.Cmd {
		return runner.BuildCommand(ctx, true, cmdName, args...)
	})
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}
