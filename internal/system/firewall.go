package system

import (
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
// gemeldet (kein Fehler).
func Firewall() (*FirewallStatus, error) {
	st := &FirewallStatus{}
	if !commandExists("ufw") {
		return st, nil
	}
	st.Available = true

	out, err := run("ufw", "status", "verbose")
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

// FirewallSetEnabled aktiviert oder deaktiviert UFW.
func FirewallSetEnabled(enabled bool) error {
	if !commandExists("ufw") {
		return errMissingTool("ufw")
	}
	action := "disable"
	if enabled {
		action = "enable"
	}
	_, err := run("ufw", "--force", action)
	return err
}
