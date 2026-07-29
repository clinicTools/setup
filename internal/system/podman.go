package system

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// --- Datentypen ---

// PodmanStatus beschreibt Verfügbarkeit und Eckdaten der Podman-Installation.
type PodmanStatus struct {
	Installed          bool   `json:"installed"`
	Version            string `json:"version,omitempty"`
	ComposeAvailable   bool   `json:"composeAvailable"`
	ComposeCommand     string `json:"composeCommand,omitempty"`
	ComposeNeedsSocket bool   `json:"composeNeedsSocket"`
	SocketActive       bool   `json:"socketActive"`
	Containers         int    `json:"containers"`
	Running            int    `json:"running"`
	Images             int    `json:"images"`
}

// Container ist eine vereinfachte Sicht auf einen Podman-Container.
type Container struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created string   `json:"created,omitempty"`
	Ports   []string `json:"ports"`
	// Stack/Service stammen aus den Compose-Labels (sofern gesetzt).
	Stack   string `json:"stack,omitempty"`
	Service string `json:"service,omitempty"`
}

// ContainerImage beschreibt ein lokales Image.
type ContainerImage struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Size    int64    `json:"size"`
	Created string   `json:"created,omitempty"`
}

// ContainerVolume beschreibt ein Podman-Volume.
type ContainerVolume struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// Stack ist ein per compose.yaml verwalteter Verbund von Containern.
type Stack struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Containers int    `json:"containers"`
	Running    int    `json:"running"`
	Status     string `json:"status"` // running | partial | stopped | unknown
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

// stacksDir ist das Verzeichnis der verwalteten Compose-Projekte. Es gehört
// root und ist nur privilegiert lesbar — Compose-Dateien enthalten häufig
// Zugangsdaten (Umgebungsvariablen, Passwörter).
const stacksDir = "/etc/debian-admin/stacks"

// composeFileName ist der Dateiname innerhalb eines Stack-Verzeichnisses.
const composeFileName = "compose.yaml"

// --- Verfügbarkeit & Status ---

// PodmanInstalled meldet, ob das podman-Binary vorhanden ist.
func PodmanInstalled() bool { return commandExists("podman") }

// ComposeCommand ermittelt das verfügbare Compose-Werkzeug. Zurückgegeben
// werden Programmname und voranzustellende Argumente, z. B.
// ("podman-compose", nil) oder ("podman", ["compose"]).
//
// Reihenfolge bewusst so gewählt:
//  1. `podman-compose` — spricht direkt die Podman-CLI an und funktioniert
//     ohne laufenden API-Socket.
//  2. `podman compose` — delegiert an einen externen Provider (docker-compose),
//     der eine aktive `podman.socket` VORAUSSETZT; sonst schlägt es mit
//     „failed to connect to the docker API" fehl.
//  3. `docker compose` — nur als letzte Rückfallebene.
func ComposeCommand() (string, []string, bool) {
	if commandExists("podman-compose") {
		return "podman-compose", nil, true
	}
	if commandExists("podman") {
		// `podman compose version` schlägt fehl, wenn kein Provider installiert ist.
		if err := exec.Command("podman", "compose", "version").Run(); err == nil {
			return "podman", []string{"compose"}, true
		}
	}
	if commandExists("docker") {
		if err := exec.Command("docker", "compose", "version").Run(); err == nil {
			return "docker", []string{"compose"}, true
		}
	}
	return "", nil, false
}

// ComposeNeedsSocket meldet, ob das gewählte Compose-Werkzeug den Podman-API-
// Socket benötigt (Delegation an docker-compose). Die Oberfläche weist dann auf
// `systemctl enable --now podman.socket` hin.
func ComposeNeedsSocket() bool {
	name, _, ok := ComposeCommand()
	return ok && name != "podman-compose"
}

// PodmanGetStatus liefert den aktuellen Podman-Zustand.
func PodmanGetStatus(r *Runner) *PodmanStatus {
	st := &PodmanStatus{Installed: PodmanInstalled()}
	if !st.Installed {
		return st
	}

	if out, err := r.run("podman", "--version"); err == nil {
		// Ausgabe: „podman version 4.9.3"
		fields := strings.Fields(out)
		if len(fields) > 0 {
			st.Version = fields[len(fields)-1]
		}
	}

	name, args, ok := ComposeCommand()
	st.ComposeAvailable = ok
	if ok {
		st.ComposeCommand = strings.TrimSpace(name + " " + strings.Join(args, " "))
		st.ComposeNeedsSocket = name != "podman-compose"
	}

	if commandExists("systemctl") {
		if out, _ := r.run("systemctl", "is-active", "podman.socket"); out == "active" {
			st.SocketActive = true
		}
	}

	if containers, err := Containers(r, true); err == nil {
		st.Containers = len(containers)
		for _, c := range containers {
			if c.State == "running" {
				st.Running++
			}
		}
	}
	if images, err := ContainerImages(r); err == nil {
		st.Images = len(images)
	}
	return st
}

// --- Container ---

// podmanContainerJSON spiegelt die relevanten Felder von `podman ps --format json`.
type podmanContainerJSON struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Image  string            `json:"Image"`
	State  string            `json:"State"`
	Status string            `json:"Status"`
	Labels map[string]string `json:"Labels"`
	Ports  []struct {
		HostIP        string `json:"host_ip"`
		HostPort      int    `json:"host_port"`
		ContainerPort int    `json:"container_port"`
		Protocol      string `json:"protocol"`
	} `json:"Ports"`
	CreatedAt string `json:"CreatedAt"`
}

// Containers listet Container (rootful/System-Ebene, daher privilegiert).
func Containers(r *Runner, all bool) ([]Container, error) {
	if !PodmanInstalled() {
		return nil, errMissingTool("podman")
	}
	args := []string{"ps", "--format", "json"}
	if all {
		args = append(args, "--all")
	}
	out, err := r.sudo("podman", args...)
	if err != nil {
		return nil, err
	}

	var raw []podmanContainerJSON
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("podman-JSON nicht lesbar: %w", err)
	}

	containers := make([]Container, 0, len(raw))
	for _, c := range raw {
		item := Container{
			ID:      shortID(c.ID),
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.CreatedAt,
			Stack:   composeProject(c.Labels),
			Service: composeService(c.Labels),
			Ports:   []string{},
		}
		if len(c.Names) > 0 {
			item.Name = c.Names[0]
		}
		for _, p := range c.Ports {
			if p.HostPort > 0 {
				item.Ports = append(item.Ports,
					fmt.Sprintf("%d→%d/%s", p.HostPort, p.ContainerPort, p.Protocol))
			} else {
				item.Ports = append(item.Ports, fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
			}
		}
		containers = append(containers, item)
	}
	sort.Slice(containers, func(i, j int) bool { return containers[i].Name < containers[j].Name })
	return containers, nil
}

// composeProject liest den Projektnamen aus den bekannten Compose-Labels.
func composeProject(labels map[string]string) string {
	for _, key := range []string{"com.docker.compose.project", "io.podman.compose.project"} {
		if v := labels[key]; v != "" {
			return v
		}
	}
	return ""
}

func composeService(labels map[string]string) string {
	for _, key := range []string{"com.docker.compose.service", "io.podman.compose.service"} {
		if v := labels[key]; v != "" {
			return v
		}
	}
	return ""
}

// ContainerAction führt eine Lifecycle-Operation auf einem Container aus.
// action ∈ {start, stop, restart, kill, pause, unpause, rm}.
func ContainerAction(r *Runner, id, action string) error {
	if !validContainerRef(id) {
		return invalidInput("ungültige Container-Kennung: %q", id)
	}
	var args []string
	switch action {
	case "start", "stop", "restart", "kill", "pause", "unpause":
		args = []string{action, id}
	case "rm":
		args = []string{"rm", "-f", id}
	default:
		return invalidInput("unbekannte Aktion: %q", action)
	}
	_, err := r.sudo("podman", args...)
	return err
}

// ContainerLogs liefert die jüngsten Logzeilen eines Containers.
func ContainerLogs(r *Runner, id string, lines int) (string, error) {
	if !validContainerRef(id) {
		return "", invalidInput("ungültige Container-Kennung: %q", id)
	}
	if lines <= 0 || lines > 5000 {
		lines = 200
	}
	// podman logs schreibt teilweise auf stderr; Fehler daher tolerant behandeln.
	out, _ := r.sudo("podman", "logs", "--tail", itoa(lines), id)
	return out, nil
}

// ContainerImages listet lokale Images.
func ContainerImages(r *Runner) ([]ContainerImage, error) {
	if !PodmanInstalled() {
		return nil, errMissingTool("podman")
	}
	out, err := r.sudo("podman", "images", "--format", "json")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		ID      string   `json:"Id"`
		Names   []string `json:"Names"`
		Size    int64    `json:"Size"`
		Created int64    `json:"Created"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("podman-JSON nicht lesbar: %w", err)
	}
	images := make([]ContainerImage, 0, len(raw))
	for _, i := range raw {
		img := ContainerImage{ID: shortID(i.ID), Names: i.Names, Size: i.Size}
		if i.Created > 0 {
			img.Created = unixSecRFC3339(i.Created)
		}
		images = append(images, img)
	}
	return images, nil
}

// ContainerVolumes listet Podman-Volumes.
func ContainerVolumes(r *Runner) ([]ContainerVolume, error) {
	if !PodmanInstalled() {
		return nil, errMissingTool("podman")
	}
	out, err := r.sudo("podman", "volume", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Name       string `json:"Name"`
		Driver     string `json:"Driver"`
		Mountpoint string `json:"Mountpoint"`
		CreatedAt  string `json:"CreatedAt"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("podman-JSON nicht lesbar: %w", err)
	}
	vols := make([]ContainerVolume, 0, len(raw))
	for _, v := range raw {
		vols = append(vols, ContainerVolume(v))
	}
	return vols, nil
}

// --- Compose-Stacks ---

// StackPath liefert den Pfad der compose.yaml eines Stacks.
func StackPath(name string) string {
	return stacksDir + "/" + name + "/" + composeFileName
}

// Stacks listet die verwalteten Compose-Projekte samt Container-Zustand.
func Stacks(r *Runner) ([]Stack, error) {
	// Verzeichnisliste privilegiert lesen (Stack-Verzeichnis ist root-only).
	out, err := r.sudo("sh", "-c", "ls -1 "+stacksDir+" 2>/dev/null || true")
	if err != nil {
		return nil, err
	}

	containers, _ := Containers(r, true)
	byStack := map[string][]Container{}
	for _, c := range containers {
		if c.Stack != "" {
			byStack[c.Stack] = append(byStack[c.Stack], c)
		}
	}

	var stacks []Stack
	for _, name := range lines(out) {
		if !validStackName(name) {
			continue
		}
		s := Stack{Name: name, Path: StackPath(name), Status: "stopped"}
		for _, c := range byStack[name] {
			s.Containers++
			if c.State == "running" {
				s.Running++
			}
		}
		switch {
		case s.Containers == 0:
			s.Status = "stopped"
		case s.Running == s.Containers:
			s.Status = "running"
		case s.Running > 0:
			s.Status = "partial"
		default:
			s.Status = "stopped"
		}
		stacks = append(stacks, s)
	}
	sort.Slice(stacks, func(i, j int) bool { return stacks[i].Name < stacks[j].Name })
	return stacks, nil
}

// StackRead liefert den Inhalt der compose.yaml eines Stacks.
func StackRead(r *Runner, name string) (string, error) {
	if !validStackName(name) {
		return "", invalidInput("ungültiger Stack-Name: %q", name)
	}
	return r.sudo("cat", StackPath(name))
}

// StackWrite legt einen Stack an bzw. aktualisiert seine compose.yaml.
func StackWrite(r *Runner, name, content string) error {
	if !validStackName(name) {
		return invalidInput("ungültiger Stack-Name: %q (erlaubt: a–z, 0–9, -, _)", name)
	}
	if strings.TrimSpace(content) == "" {
		return invalidInput("compose.yaml darf nicht leer sein")
	}
	dir := stacksDir + "/" + name
	// Verzeichnis anlegen und restriktiv absichern (Compose-Dateien enthalten
	// häufig Zugangsdaten).
	if _, err := r.sudo("install", "-d", "-m", "0700", dir); err != nil {
		return err
	}
	// Inhalt über stdin schreiben — keine Shell-Interpolation, kein Injection-Risiko.
	if _, err := r.sudoStdin(content, "tee", StackPath(name)); err != nil {
		return err
	}
	_, err := r.sudo("chmod", "0600", StackPath(name))
	return err
}

// StackDelete entfernt das Stack-Verzeichnis (nach vorherigem „down").
func StackDelete(r *Runner, name string) error {
	if !validStackName(name) {
		return invalidInput("ungültiger Stack-Name: %q", name)
	}
	_, err := r.sudo("rm", "-rf", stacksDir+"/"+name)
	return err
}

// StackComposeArgs baut die Argumentliste für eine Compose-Aktion.
// action ∈ {up, down, restart, pull, stop, start}.
func StackComposeArgs(name, action string) (string, []string, error) {
	if !validStackName(name) {
		return "", nil, invalidInput("ungültiger Stack-Name: %q", name)
	}
	cmd, base, ok := ComposeCommand()
	if !ok {
		return "", nil, fmt.Errorf("kein Compose-Werkzeug gefunden (podman compose / podman-compose / docker compose)")
	}

	args := append([]string{}, base...)
	args = append(args, "-f", StackPath(name), "-p", name)
	switch action {
	case "up":
		args = append(args, "up", "-d", "--remove-orphans")
	case "down":
		args = append(args, "down")
	case "restart":
		args = append(args, "restart")
	case "pull":
		args = append(args, "pull")
	case "stop":
		args = append(args, "stop")
	case "start":
		args = append(args, "start")
	default:
		return "", nil, invalidInput("unbekannte Stack-Aktion: %q", action)
	}
	return cmd, args, nil
}

// --- Validierung & Helfer ---

// validStackName erlaubt nur Compose-taugliche Projektnamen.
func validStackName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case (r == '-' || r == '_') && i > 0:
		default:
			return false
		}
	}
	return true
}

// validContainerRef erlaubt Container-IDs und -Namen ohne Sonderzeichen.
func validContainerRef(ref string) bool {
	if ref == "" || len(ref) > 128 || strings.HasPrefix(ref, "-") {
		return false
	}
	for _, r := range ref {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return false
		}
	}
	return true
}

// shortID kürzt eine Container-/Image-ID auf die üblichen 12 Zeichen.
func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
