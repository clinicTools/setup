package auth

// User beschreibt einen am Betriebssystem authentifizierten Benutzer.
type User struct {
	// Username ist der Unix-Login-Name.
	Username string `json:"username"`
	// UID/GID stammen aus /etc/passwd.
	UID int `json:"uid"`
	GID int `json:"gid"`
	// FullName ist das GECOS-Feld (sofern gesetzt).
	FullName string `json:"fullName"`
	// HomeDir und Shell aus /etc/passwd.
	HomeDir string `json:"homeDir"`
	Shell   string `json:"shell"`
	// Groups sind die aufgelösten Gruppennamen des Benutzers.
	Groups []string `json:"groups"`
	// Admin kennzeichnet Mitglieder einer privilegierten Gruppe.
	Admin bool `json:"admin"`
}
