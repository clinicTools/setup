package system

import (
	"encoding/json"
	"fmt"
	"strings"
)

// cronManagedPrefix kennzeichnet von dieser Anwendung verwaltete cron.d-Dateien.
const cronManagedPrefix = "debian-admin-"

// CreateCronJob legt einen System-Cron-Job als verwaltete Datei unter
// /etc/cron.d an. name dient als Dateibasis (nur [a-z0-9-]).
func CreateCronJob(r *Runner, name, schedule, user, command string) error {
	if !validCronName(name) {
		return fmt.Errorf("ungültiger Name: %q", name)
	}
	if !validCronSchedule(schedule) {
		return fmt.Errorf("ungültiger Zeitplan: %q", schedule)
	}
	if !validUsername(user) {
		return fmt.Errorf("ungültiger Benutzer: %q", user)
	}
	if command == "" || strings.ContainsAny(command, "\n\r") {
		return fmt.Errorf("ungültiges Kommando")
	}

	path := "/etc/cron.d/" + cronManagedPrefix + name
	content := fmt.Sprintf("# Verwaltet von debian-admin\n%s %s %s\n", schedule, user, command)
	// Privilegiert schreiben: tee liest den Inhalt (hinter der sudo-Passwortzeile)
	// von stdin und schreibt ihn nach /etc/cron.d.
	_, err := r.sudoStdin(content, "tee", path)
	return err
}

// DeleteCronJob entfernt einen zuvor angelegten verwalteten Cron-Job.
func DeleteCronJob(r *Runner, name string) error {
	if !validCronName(name) {
		return fmt.Errorf("ungültiger Name: %q", name)
	}
	_, err := r.sudo("rm", "-f", "/etc/cron.d/"+cronManagedPrefix+name)
	return err
}

func validCronName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

// validCronSchedule prüft ein 5-Feld-Cron-Muster mit erlaubten Zeichen.
func validCronSchedule(s string) bool {
	fields := strings.Fields(s)
	if len(fields) != 5 {
		return false
	}
	for _, f := range fields {
		for _, r := range f {
			switch {
			case r >= '0' && r <= '9':
			case r == '*' || r == '/' || r == ',' || r == '-':
			default:
				return false
			}
		}
	}
	return true
}

// Timer beschreibt einen systemd-Timer (moderne Cron-Alternative).
type Timer struct {
	Unit      string `json:"unit"`
	Next      string `json:"next"`
	Left      string `json:"left"`
	Last      string `json:"last"`
	Passed    string `json:"passed"`
	Activates string `json:"activates"`
}

// CronJob beschreibt eine Zeile aus einer crontab/cron.d-Datei.
type CronJob struct {
	Source   string `json:"source"`
	Schedule string `json:"schedule"`
	User     string `json:"user,omitempty"`
	Command  string `json:"command"`
}

// ScheduledTasks bündelt systemd-Timer und klassische Cron-Jobs.
type ScheduledTasks struct {
	Timers []Timer   `json:"timers"`
	Cron   []CronJob `json:"cron"`
}

// Scheduled sammelt geplante Aufgaben (Timer + system-cron).
func Scheduled() (*ScheduledTasks, error) {
	return &ScheduledTasks{
		Timers: systemdTimers(),
		Cron:   systemCron(),
	}, nil
}

func systemdTimers() []Timer {
	if !commandExists("systemctl") {
		return nil
	}
	out, err := run("systemctl", "list-timers", "--all", "--no-pager",
		"--output=json")
	if err != nil {
		return nil
	}
	var raw []struct {
		Unit      string `json:"unit"`
		Next      string `json:"next"`
		Left      string `json:"left"`
		Last      string `json:"last"`
		Passed    string `json:"passed"`
		Activates string `json:"activates"`
	}
	if json.Unmarshal([]byte(out), &raw) != nil {
		return nil
	}
	timers := make([]Timer, 0, len(raw))
	for _, t := range raw {
		timers = append(timers, Timer(t))
	}
	return timers
}

// systemCron parst /etc/crontab und /etc/cron.d/* (mit User-Feld).
func systemCron() []CronJob {
	var jobs []CronJob
	sources := []string{"/etc/crontab"}
	if entries, err := readDir("/etc/cron.d"); err == nil {
		for _, e := range entries {
			sources = append(sources, "/etc/cron.d/"+e)
		}
	}

	for _, src := range sources {
		data, err := readFile(src)
		if err != nil {
			continue
		}
		for _, l := range lines(data) {
			if strings.HasPrefix(l, "#") || strings.Contains(l, "=") {
				continue // Kommentare und Umgebungsvariablen überspringen
			}
			fields := strings.Fields(l)
			if len(fields) < 7 {
				continue
			}
			schedule := strings.Join(fields[0:5], " ")
			jobs = append(jobs, CronJob{
				Source:   src,
				Schedule: schedule,
				User:     fields[5],
				Command:  strings.Join(fields[6:], " "),
			})
		}
	}
	return jobs
}
