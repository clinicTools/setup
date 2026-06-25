package auth

import (
	"bufio"
	"os"
	"strings"
)

// lookupShell liest die Login-Shell eines Benutzers aus /etc/passwd, da das
// os/user-Paket dieses Feld nicht bereitstellt. Bei Fehlern wird "" geliefert.
func lookupShell(username string) string {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		// Format: name:passwd:uid:gid:gecos:home:shell
		fields := strings.Split(line, ":")
		if len(fields) >= 7 && fields[0] == username {
			return fields[6]
		}
	}
	return ""
}
