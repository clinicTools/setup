// Package config lädt die Laufzeitkonfiguration des Debian-Admin-Dienstes
// aus Umgebungsvariablen. Sinnvolle Defaults erlauben den Start ohne jegliche
// Konfiguration; sicherheitsrelevante Werte (Session-Secret) werden bei Bedarf
// zur Laufzeit generiert.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config bündelt alle einstellbaren Parameter des Dienstes.
type Config struct {
	// Addr ist die Listen-Adresse des HTTP-Servers (Host:Port).
	Addr string
	// SessionSecret signiert die JWT-Session-Cookies (HS256).
	SessionSecret []byte
	// SessionTTL bestimmt die Gültigkeitsdauer einer Anmeldung.
	SessionTTL time.Duration
	// PAMService ist der Name des verwendeten PAM-Stacks (z. B. "login").
	PAMService string
	// AdminGroups listet Unix-Gruppen, deren Mitglieder als privilegiert gelten.
	AdminGroups []string
	// AllowInsecureCookie erlaubt Cookies ohne Secure-Flag (nur für lokale Entwicklung).
	AllowInsecureCookie bool
	// Dev aktiviert CORS für den Vite-Dev-Server (Port 5173).
	Dev bool
	// IdleTimeout beendet den Prozess, wenn so lange keine Verbindung/Aktivität
	// vorliegt (0 = deaktiviert). In Kombination mit systemd-Socket-Activation
	// ergibt sich das Cockpit-Verhalten „läuft nur bei Verbindung".
	IdleTimeout time.Duration
}

// Load liest die Konfiguration aus der Umgebung.
func Load() *Config {
	c := &Config{
		Addr:                env("DA_ADDR", ":8088"),
		SessionTTL:          envDuration("DA_SESSION_TTL", 8*time.Hour),
		PAMService:          env("DA_PAM_SERVICE", "login"),
		AdminGroups:         envList("DA_ADMIN_GROUPS", "sudo,admin,wheel"),
		AllowInsecureCookie: envBool("DA_ALLOW_INSECURE_COOKIE", false),
		Dev:                 envBool("DA_DEV", false),
		IdleTimeout:         envDuration("DA_IDLE_TIMEOUT", 0),
	}

	if secret := os.Getenv("DA_SESSION_SECRET"); secret != "" {
		c.SessionSecret = []byte(secret)
	} else {
		// Flüchtiges Secret: Neustarts invalidieren bestehende Sessions.
		buf := make([]byte, 32)
		_, _ = rand.Read(buf)
		c.SessionSecret = []byte(hex.EncodeToString(buf))
	}

	return c
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func envList(key, def string) []string {
	raw := env(key, def)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
