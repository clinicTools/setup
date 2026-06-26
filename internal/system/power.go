package system

import "fmt"

// PowerAction führt einen System-Power-Vorgang aus.
// action ∈ {reboot, poweroff}. Der Aufruf erfolgt über systemctl, sodass
// laufende Dienste regulär heruntergefahren werden.
func PowerAction(r *Runner, action string) error {
	switch action {
	case "reboot", "poweroff":
	default:
		return fmt.Errorf("unbekannte Power-Aktion: %q", action)
	}
	if !commandExists("systemctl") {
		return errMissingTool("systemctl")
	}
	_, err := r.sudo("systemctl", action)
	return err
}
