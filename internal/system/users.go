package system

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SystemUser spiegelt einen Eintrag aus /etc/passwd inkl. Gruppen wider.
type SystemUser struct {
	Username string   `json:"username"`
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
	FullName string   `json:"fullName"`
	HomeDir  string   `json:"homeDir"`
	Shell    string   `json:"shell"`
	System   bool     `json:"system"` // UID < 1000 (außer root)
	Groups   []string `json:"groups"`
}

// SystemGroup spiegelt einen Eintrag aus /etc/group wider.
type SystemGroup struct {
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	System  bool     `json:"system"`
	Members []string `json:"members"`
}

// Users liest alle Benutzerkonten aus /etc/passwd und reichert sie um die
// zugehörigen Gruppen aus /etc/group an.
func Users() ([]SystemUser, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	groupsByUser := userGroupIndex()

	var users []SystemUser
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Split(sc.Text(), ":")
		if len(fields) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(fields[2])
		gid, _ := strconv.Atoi(fields[3])
		u := SystemUser{
			Username: fields[0],
			UID:      uid,
			GID:      gid,
			FullName: strings.Split(fields[4], ",")[0],
			HomeDir:  fields[5],
			Shell:    fields[6],
			System:   uid != 0 && uid < 1000,
			Groups:   groupsByUser[fields[0]],
		}
		sort.Strings(u.Groups)
		users = append(users, u)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].UID < users[j].UID })
	return users, nil
}

// Groups liest alle Gruppen aus /etc/group.
func Groups() ([]SystemGroup, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var groups []SystemGroup
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Split(sc.Text(), ":")
		if len(fields) < 4 {
			continue
		}
		gid, _ := strconv.Atoi(fields[2])
		var members []string
		if fields[3] != "" {
			members = strings.Split(fields[3], ",")
		}
		groups = append(groups, SystemGroup{
			Name:    fields[0],
			GID:     gid,
			System:  gid < 1000,
			Members: members,
		})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].GID < groups[j].GID })
	return groups, nil
}

// userGroupIndex baut eine Map Benutzer -> Gruppennamen aus /etc/group.
func userGroupIndex() map[string][]string {
	out := map[string][]string{}
	f, err := os.Open("/etc/group")
	if err != nil {
		return out
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Split(sc.Text(), ":")
		if len(fields) < 4 || fields[3] == "" {
			continue
		}
		for _, m := range strings.Split(fields[3], ",") {
			out[m] = append(out[m], fields[0])
		}
	}
	return out
}

// CreateUserRequest beschreibt die Parameter zum Anlegen eines Benutzers.
type CreateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Shell    string `json:"shell"`
	Groups   []string `json:"groups"`
	System   bool   `json:"system"`
	Password string `json:"password"`
}

// CreateUser legt einen Benutzer via useradd an und setzt optional ein Passwort.
func CreateUser(r *Runner, req CreateUserRequest) error {
	if !validUsername(req.Username) {
		return fmt.Errorf("ungültiger Benutzername: %q", req.Username)
	}
	args := []string{"-m"}
	if req.FullName != "" {
		args = append(args, "-c", req.FullName)
	}
	if req.Shell != "" {
		args = append(args, "-s", req.Shell)
	}
	if len(req.Groups) > 0 {
		args = append(args, "-G", strings.Join(sanitizeNames(req.Groups), ","))
	}
	if req.System {
		args = append(args, "-r")
	}
	args = append(args, req.Username)

	if _, err := r.sudo("useradd", args...); err != nil {
		return err
	}
	if req.Password != "" {
		return setPassword(r, req.Username, req.Password)
	}
	return nil
}

// DeleteUser entfernt einen Benutzer; removeHome löscht zusätzlich das Home.
func DeleteUser(r *Runner, username string, removeHome bool) error {
	if !validUsername(username) {
		return fmt.Errorf("ungültiger Benutzername: %q", username)
	}
	args := []string{}
	if removeHome {
		args = append(args, "-r")
	}
	args = append(args, username)
	_, err := r.sudo("userdel", args...)
	return err
}

// SetPassword setzt das Passwort eines Benutzers via chpasswd.
func SetPassword(r *Runner, username, password string) error {
	if !validUsername(username) {
		return fmt.Errorf("ungültiger Benutzername: %q", username)
	}
	return setPassword(r, username, password)
}

func setPassword(r *Runner, username, password string) error {
	// chpasswd liest „user:passwort" von stdin (hinter der sudo-Passwortzeile).
	_, err := r.sudoStdin(username+":"+password+"\n", "chpasswd")
	return err
}

// ModifyUser ändert Gruppenmitgliedschaft und/oder Login-Shell eines Benutzers.
func ModifyUser(r *Runner, username string, groups []string, shell string) error {
	if !validUsername(username) {
		return fmt.Errorf("ungültiger Benutzername: %q", username)
	}
	if groups != nil {
		// -G ersetzt die Zusatzgruppen vollständig.
		if _, err := r.sudo("usermod", "-G", strings.Join(sanitizeNames(groups), ","), username); err != nil {
			return err
		}
	}
	if shell != "" {
		if !validShell(shell) {
			return fmt.Errorf("ungültige Shell: %q", shell)
		}
		if _, err := r.sudo("usermod", "-s", shell, username); err != nil {
			return err
		}
	}
	return nil
}

// CreateGroup legt eine Gruppe an (system = Systemgruppe via -r).
func CreateGroup(r *Runner, name string, systemGroup bool) error {
	if !validUsername(name) {
		return fmt.Errorf("ungültiger Gruppenname: %q", name)
	}
	args := []string{}
	if systemGroup {
		args = append(args, "-r")
	}
	args = append(args, name)
	_, err := r.sudo("groupadd", args...)
	return err
}

// DeleteGroup entfernt eine Gruppe.
func DeleteGroup(r *Runner, name string) error {
	if !validUsername(name) {
		return fmt.Errorf("ungültiger Gruppenname: %q", name)
	}
	_, err := r.sudo("groupdel", name)
	return err
}

// validShell erlaubt nur absolute Pfade ohne Sonderzeichen.
func validShell(shell string) bool {
	if !strings.HasPrefix(shell, "/") || strings.ContainsAny(shell, " \t;|&$`\n") {
		return false
	}
	return true
}

// validUsername erzwingt konservative Unix-Namensregeln, um Command-Injection
// über useradd-Argumente auszuschließen.
func validUsername(name string) bool {
	if name == "" || len(name) > 32 {
		return false
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9' && i > 0:
		case (r == '_' || r == '-' || r == '.') && i > 0:
		default:
			return false
		}
	}
	return true
}

func sanitizeNames(in []string) []string {
	out := make([]string, 0, len(in))
	for _, n := range in {
		if validUsername(n) {
			out = append(out, n)
		}
	}
	return out
}
