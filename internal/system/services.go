package system

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Service beschreibt eine systemd-Unit (Typ .service).
type Service struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LoadState   string `json:"loadState"`
	ActiveState string `json:"activeState"` // active, inactive, failed …
	SubState    string `json:"subState"`    // running, exited, dead …
	UnitFile    string `json:"unitFileState,omitempty"` // enabled, disabled, static …
}

// Services listet alle .service-Units (geladen + installiert) zusammen mit
// ihrem Enablement-Status.
func Services() ([]Service, error) {
	if !commandExists("systemctl") {
		return nil, errMissingTool("systemctl")
	}

	// Aktive/geladene Units mit Zustand.
	out, err := run("systemctl", "list-units", "--type=service", "--all",
		"--no-legend", "--no-pager", "--plain", "--output=json")
	if err != nil {
		return nil, err
	}

	var raw []struct {
		Unit        string `json:"unit"`
		Load        string `json:"load"`
		Active      string `json:"active"`
		Sub         string `json:"sub"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("systemctl-JSON nicht lesbar: %w", err)
	}

	enablement := unitFileStates()

	services := make([]Service, 0, len(raw))
	for _, u := range raw {
		services = append(services, Service{
			Name:        u.Unit,
			Description: u.Description,
			LoadState:   u.Load,
			ActiveState: u.Active,
			SubState:    u.Sub,
			UnitFile:    enablement[u.Unit],
		})
	}
	return services, nil
}

// unitFileStates liefert eine Map Unit -> enabled/disabled/static.
func unitFileStates() map[string]string {
	out := map[string]string{}
	res, err := run("systemctl", "list-unit-files", "--type=service",
		"--no-legend", "--no-pager", "--plain", "--output=json")
	if err != nil {
		return out
	}
	var raw []struct {
		UnitFile string `json:"unit_file"`
		State    string `json:"state"`
	}
	if json.Unmarshal([]byte(res), &raw) == nil {
		for _, u := range raw {
			out[u.UnitFile] = u.State
		}
	}
	return out
}

// ServiceStatus liefert die ausführliche Statusausgabe einer Unit.
func ServiceStatus(name string) (string, error) {
	if err := validUnit(name); err != nil {
		return "", err
	}
	// status liefert bei inaktiven Units Exit-Code 3 — Ausgabe trotzdem nutzen.
	out, _ := run("systemctl", "status", name, "--no-pager", "--lines=40")
	return out, nil
}

// ServiceAction führt eine Lifecycle-Operation auf einer Unit aus.
// action ∈ {start, stop, restart, reload, enable, disable}. Die Ausführung
// erfolgt privilegiert (sudo) im Kontext des angemeldeten Benutzers.
func ServiceAction(r *Runner, name, action string) error {
	if err := validUnit(name); err != nil {
		return err
	}
	switch action {
	case "start", "stop", "restart", "reload", "enable", "disable":
	default:
		return fmt.Errorf("unbekannte Aktion: %q", action)
	}
	_, err := r.sudo("systemctl", action, name)
	return err
}

// validUnit verhindert, dass beliebige Argumente an systemctl durchgereicht
// werden (Schutz vor Flag-/Command-Injection).
func validUnit(name string) error {
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, " \t\n;|&$`") {
		return fmt.Errorf("ungültiger Unit-Name: %q", name)
	}
	return nil
}
