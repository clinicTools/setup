package system

import (
	"fmt"
	"strings"
	"time"
)

// TimeInfo beschreibt Zeit-/Zeitzonen-Einstellungen.
type TimeInfo struct {
	LocalTime  string `json:"localTime"`
	UniversalTime string `json:"universalTime"`
	Timezone   string `json:"timezone"`
	NTPEnabled bool   `json:"ntpEnabled"`
	NTPSynced  bool   `json:"ntpSynced"`
	RTCInLocal bool   `json:"rtcInLocalTime"`
}

// Time liest die Zeit-Einstellungen via timedatectl (Fallback: Go-Stdlib).
func Time() (*TimeInfo, error) {
	info := &TimeInfo{
		LocalTime:     time.Now().Format(time.RFC3339),
		UniversalTime: time.Now().UTC().Format(time.RFC3339),
	}
	zone, _ := time.Now().Zone()
	info.Timezone = zone

	if !commandExists("timedatectl") {
		return info, nil
	}
	out, err := run("timedatectl", "show")
	if err != nil {
		return info, nil
	}
	for _, l := range lines(out) {
		key, val, ok := strings.Cut(l, "=")
		if !ok {
			continue
		}
		switch key {
		case "Timezone":
			info.Timezone = val
		case "NTP":
			info.NTPEnabled = val == "yes"
		case "NTPSynchronized":
			info.NTPSynced = val == "yes"
		case "LocalRTC":
			info.RTCInLocal = val == "yes"
		}
	}
	return info, nil
}

// ListTimezones liefert die verfügbaren Zeitzonen (timedatectl list-timezones).
func ListTimezones() ([]string, error) {
	if !commandExists("timedatectl") {
		return nil, errMissingTool("timedatectl")
	}
	out, err := run("timedatectl", "list-timezones")
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

// SetTimezone setzt die Systemzeitzone (privilegiert).
func SetTimezone(r *Runner, tz string) error {
	if !commandExists("timedatectl") {
		return errMissingTool("timedatectl")
	}
	if !validTimezone(tz) {
		return fmt.Errorf("ungültige Zeitzone: %q", tz)
	}
	_, err := r.sudo("timedatectl", "set-timezone", tz)
	return err
}

// SetNTP aktiviert/deaktiviert die NTP-Zeitsynchronisation (privilegiert).
func SetNTP(r *Runner, enabled bool) error {
	if !commandExists("timedatectl") {
		return errMissingTool("timedatectl")
	}
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := r.sudo("timedatectl", "set-ntp", val)
	return err
}

// validTimezone erlaubt nur das Schema "Region/Stadt" bzw. "UTC".
func validTimezone(tz string) bool {
	if tz == "UTC" {
		return true
	}
	if strings.HasPrefix(tz, "-") || strings.ContainsAny(tz, " \t;|&$`") {
		return false
	}
	// Pfad-Traversal und absolute Pfade ausschließen.
	if strings.Contains(tz, "..") || strings.HasPrefix(tz, "/") {
		return false
	}
	return strings.Contains(tz, "/")
}
