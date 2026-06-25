# Debian Admin

Eine moderne **Web-Administrationsoberfläche für Debian-/Ubuntu-Systeme**.
Backend in **Go**, Frontend in **Vue 3** — ausgeliefert als ein einzelnes,
abhängigkeitsarmes Binary (das Frontend wird per `go:embed` eingebettet).

Die Anmeldung erfolgt **gegen die Betriebssystem-Benutzer** (PAM). Privilegierte
Aktionen erfordern die Mitgliedschaft in einer Admin-Gruppe (`sudo`/`admin`/`wheel`).

Das Design orientiert sich an der **KIS-Oberfläche** (shadcn/Tailwind, blaue
Primärfarbe, Hell-/Dunkel-Modus) und der **Windows-11-Einstellungen-App**
(Navigationsleiste mit Akzent-Pille, Breadcrumbs, SettingsCards mit Icon +
Titel + Beschreibung + rechtsbündiger Aktion, gruppierte Abschnitte).

---

## Funktionsumfang

| Modul | Funktionen |
|-------|-----------|
| **Übersicht** | Host-Eckdaten, Laufzeit, Last, RAM-/Swap-/Disk-Auslastung, Kernel/CPU |
| **Prozesse** | Laufende Prozesse nach Speicherverbrauch, Suche |
| **Dienste** | systemd-Units auflisten, starten/stoppen/neustarten, Enable-Status, Filter |
| **Benutzer & Gruppen** | Konten/Gruppen anzeigen, Benutzer anlegen/löschen, Passwort setzen |
| **Netzwerk** | Schnittstellen, IP-/MAC-Adressen, DNS, Gateway |
| **Speicher** | Dateisysteme (Auslastung), Blockgeräte (lsblk) |
| **Pakete & Updates** | Anzahl installierter Pakete, verfügbare Upgrades, `apt update` |
| **Firewall** | UFW-Status & -Regeln, Aktivieren/Deaktivieren |
| **Geplante Aufgaben** | systemd-Timer und Cron-Jobs |
| **Datum & Uhrzeit** | Zeitzone setzen, NTP-Synchronisation umschalten |
| **Systemprotokolle** | journald-Einträge mit Unit-/Prioritätsfilter |
| **Energie** | Neustart / Herunterfahren (nur Admins) |

Lesende Endpunkte stehen allen angemeldeten Benutzern offen; **schreibende
Aktionen sind auf Admin-Gruppen beschränkt** (serverseitig erzwungen).

## Architektur

```
setup/
├── cmd/debian-admin/        # main(): Konfiguration laden, HTTP-Server starten
├── internal/
│   ├── config/              # Konfiguration aus Umgebungsvariablen
│   ├── auth/                # PAM-Authentifizierung, JWT-Sessions, Middleware
│   ├── system/              # Linux-Collectoren (/proc, systemctl, apt, ip, …)
│   ├── api/                 # HTTP-Handler (JSON-REST)
│   ├── server/              # chi-Router, Middleware, Routen
│   └── web/                 # go:embed des gebauten Frontends (dist/)
├── web/                     # Vue-3-Frontend (Vite, Tailwind 4, Pinia, vue-router)
│   └── src/
│       ├── components/      # UI-Primitive + SettingsCard/PageHeader/…
│       ├── layouts/         # AppShell (Navigationsleiste, Topbar)
│       ├── views/           # Eine View je Modul
│       ├── composables/     # useAsyncData, useTheme, useToast
│       ├── stores/          # Pinia: Auth
│       └── lib/             # API-Client, Typen, Formatierung
└── deploy/                  # systemd-Unit + PAM-Stack
```

**Request-Fluss:** `Client → chi-Router → [securityHeaders] → /api → [Session-Middleware] → [RequireAdmin bei Schreibzugriff] → Handler → system-Collector → systemctl/apt/proc`.

### Authentifizierung

1. `POST /api/auth/login` prüft `username`/`password` über **PAM** (Service `login`,
   in Produktion eigener Stack `debian-admin`).
2. Bei Erfolg wird ein signiertes **JWT** als HTTP-only-Cookie gesetzt (HS256,
   8 h Gültigkeit, `Secure` + `SameSite=Lax`).
3. Gruppen werden aus dem System aufgelöst; Mitglieder einer Admin-Gruppe
   erhalten das `admin`-Flag.

## Schnellstart (Entwicklung)

Voraussetzungen: Go ≥ 1.24 (mit CGO/`libpam0g-dev`), Node ≥ 20, ein Linux-Host.

```bash
sudo apt-get install -y libpam0g-dev      # PAM-Header für CGO

# 1) Backend (Dev-Modus: CORS für Vite, unsichere Cookies erlaubt)
make run                                    # lauscht auf :8088

# 2) Frontend-Dev-Server mit Hot Reload (zweites Terminal)
make dev                                    # Vite auf :5173, proxyt /api → :8088
```

Im Dev-Modus läuft die UI unter <http://localhost:5173>.

## Produktions-Build

```bash
make build          # baut Frontend (→ internal/web/dist) und Go-Binary (→ bin/debian-admin)
```

Das Ergebnis ist ein **einzelnes Binary** mit eingebettetem Frontend.

### Deployment

```bash
sudo cp bin/debian-admin /usr/local/bin/
sudo cp deploy/debian-admin.service /etc/systemd/system/
sudo cp deploy/debian-admin.pam     /etc/pam.d/debian-admin
sudo systemctl daemon-reload
sudo systemctl enable --now debian-admin
```

Der Dienst läuft als `root` (für PAM und administrative Kommandos). Ein
Reverse-Proxy (nginx/caddy) sollte **TLS** terminieren.

## Konfiguration (Umgebungsvariablen)

| Variable | Default | Bedeutung |
|----------|---------|-----------|
| `DA_ADDR` | `:8088` | Listen-Adresse |
| `DA_PAM_SERVICE` | `login` | PAM-Service-Name |
| `DA_ADMIN_GROUPS` | `sudo,admin,wheel` | Privilegierte Gruppen |
| `DA_SESSION_SECRET` | *(zufällig)* | JWT-Signaturschlüssel (in Prod fest setzen!) |
| `DA_SESSION_TTL` | `8h` | Session-Gültigkeit |
| `DA_ALLOW_INSECURE_COOKIE` | `false` | Cookie ohne `Secure` (nur Dev/HTTP) |
| `DA_DEV` | `false` | CORS für den Vite-Dev-Server aktivieren |

## Tests

```bash
make test       # Go-Tests (Auth-Sessions, Eingabevalidierung)
make vet        # go vet
cd web && npm run typecheck
```

## Sicherheitshinweise

- Eingaben für `useradd`, `systemctl`, `timedatectl` etc. werden **streng
  validiert** (Allowlist-Zeichen, Flag-/Command-Injection ausgeschlossen).
- Schreibende Endpunkte erfordern Admin-Gruppen-Mitgliedschaft.
- Security-Header (`X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`,
  `Referrer-Policy`, `Permissions-Policy`) sind aktiv.
- Da der Dienst als root läuft, sollte der Zugang netzseitig (Firewall,
  Reverse-Proxy, ggf. mTLS/VPN) zusätzlich abgesichert werden.
