package auth

import (
	"errors"
	"fmt"
	"os/user"
	"strconv"

	"github.com/msteinert/pam/v2"
)

// ErrInvalidCredentials wird bei fehlgeschlagener PAM-Authentifizierung zurückgegeben.
var ErrInvalidCredentials = errors.New("ungültiger Benutzername oder Passwort")

// Authenticator authentifiziert Benutzer gegen das Betriebssystem.
type Authenticator struct {
	service     string
	adminGroups map[string]struct{}
}

// NewAuthenticator erzeugt einen PAM-Authenticator für den angegebenen
// PAM-Service (z. B. "login") und die Liste privilegierter Gruppen.
func NewAuthenticator(service string, adminGroups []string) *Authenticator {
	groups := make(map[string]struct{}, len(adminGroups))
	for _, g := range adminGroups {
		groups[g] = struct{}{}
	}
	return &Authenticator{service: service, adminGroups: groups}
}

// Authenticate prüft username/password über PAM. Bei Erfolg wird das
// aufgelöste Benutzerprofil (inkl. Gruppen) zurückgegeben.
func (a *Authenticator) Authenticate(username, password string) (*User, error) {
	tx, err := pam.StartFunc(a.service, username, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff, pam.PromptEchoOn:
			return password, nil
		case pam.ErrorMsg, pam.TextInfo:
			return "", nil
		default:
			return "", fmt.Errorf("unbekannter PAM-Prompt-Stil: %v", s)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("PAM-Transaktion fehlgeschlagen: %w", err)
	}
	defer func() { _ = tx.End() }()

	if err := tx.Authenticate(0); err != nil {
		return nil, ErrInvalidCredentials
	}
	// Kontovalidierung (abgelaufen/gesperrt) durchsetzen.
	if err := tx.AcctMgmt(0); err != nil {
		return nil, ErrInvalidCredentials
	}

	return a.Lookup(username)
}

// Lookup löst ein Benutzerprofil ohne Passwortprüfung auf. Wird für die
// Anreicherung der Session-Daten und den /me-Endpunkt genutzt.
func (a *Authenticator) Lookup(username string) (*User, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("Benutzer %q nicht gefunden: %w", username, err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	out := &User{
		Username: u.Username,
		UID:      uid,
		GID:      gid,
		FullName: u.Name,
		HomeDir:  u.HomeDir,
	}

	if gids, err := u.GroupIds(); err == nil {
		for _, g := range gids {
			grp, err := user.LookupGroupId(g)
			if err != nil {
				continue
			}
			out.Groups = append(out.Groups, grp.Name)
			if _, ok := a.adminGroups[grp.Name]; ok {
				out.Admin = true
			}
		}
	}

	out.Shell = lookupShell(username)
	return out, nil
}
