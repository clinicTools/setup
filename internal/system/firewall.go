package system

import (
	"fmt"
	"strings"
)

// FirewallStatus beschreibt den Zustand der UFW-Firewall.
type FirewallStatus struct {
	Available bool          `json:"available"`
	Enabled   bool          `json:"enabled"`
	Rules     []FirewallRule `json:"rules"`
}

// FirewallRule ist eine einzelne UFW-Regel.
type FirewallRule struct {
	To     string `json:"to"`
	Action string `json:"action"`
	From   string `json:"from"`
}

// Firewall liest den UFW-Status. Ist ufw nicht installiert, wird Available=false
// gemeldet (kein Fehler). `ufw status` benötigt root und wird daher privilegiert
// im Benutzerkontext ausgeführt.
func Firewall(r *Runner) (*FirewallStatus, error) {
	st := &FirewallStatus{}
	if !commandExists("ufw") {
		return st, nil
	}
	st.Available = true

	out, err := r.sudo("ufw", "status", "verbose")
	if err != nil {
		return st, nil
	}
	for _, l := range lines(out) {
		if strings.HasPrefix(l, "Status:") {
			st.Enabled = strings.Contains(l, "active")
			continue
		}
		// Regelzeilen: "<to>   <action>   <from>"
		if strings.Contains(l, "ALLOW") || strings.Contains(l, "DENY") ||
			strings.Contains(l, "REJECT") || strings.Contains(l, "LIMIT") {
			fields := strings.Fields(l)
			if len(fields) >= 3 && !strings.EqualFold(fields[0], "To") {
				st.Rules = append(st.Rules, FirewallRule{
					To:     fields[0],
					Action: fields[1],
					From:   strings.Join(fields[2:], " "),
				})
			}
		}
	}
	return st, nil
}

// FirewallAddRule fügt eine UFW-Regel hinzu. action ∈ {allow, deny, reject,
// limit}, port ist numerisch (1–65535), proto ∈ {"", tcp, udp}.
func FirewallAddRule(r *Runner, action, port, proto string) error {
	if !commandExists("ufw") {
		return errMissingTool("ufw")
	}
	if !validFirewallAction(action) {
		return fmt.Errorf("ungültige Aktion: %q", action)
	}
	if !validPort(port) {
		return fmt.Errorf("ungültiger Port: %q", port)
	}
	target := port
	if proto == "tcp" || proto == "udp" {
		target = port + "/" + proto
	} else if proto != "" {
		return fmt.Errorf("ungültiges Protokoll: %q", proto)
	}
	_, err := r.sudo("ufw", action, target)
	return err
}

// FirewallDeleteRule entfernt eine UFW-Regel anhand ihrer Spezifikation.
func FirewallDeleteRule(r *Runner, action, port, proto string) error {
	if !commandExists("ufw") {
		return errMissingTool("ufw")
	}
	if !validFirewallAction(action) || !validPort(port) {
		return fmt.Errorf("ungültige Regelangabe")
	}
	target := port
	if proto == "tcp" || proto == "udp" {
		target = port + "/" + proto
	}
	_, err := r.sudo("ufw", "--force", "delete", action, target)
	return err
}

func validFirewallAction(a string) bool {
	switch a {
	case "allow", "deny", "reject", "limit":
		return true
	}
	return false
}

func validPort(p string) bool {
	if p == "" {
		return false
	}
	n := 0
	for _, r := range p {
		if r < '0' || r > '9' {
			return false
		}
		n = n*10 + int(r-'0')
	}
	return n >= 1 && n <= 65535
}

// FirewallSetEnabled aktiviert oder deaktiviert UFW.
func FirewallSetEnabled(r *Runner, enabled bool) error {
	if !commandExists("ufw") {
		return errMissingTool("ufw")
	}
	action := "disable"
	if enabled {
		action = "enable"
	}
	_, err := r.sudo("ufw", "--force", action)
	return err
}
