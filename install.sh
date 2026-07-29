#!/usr/bin/env bash
#
# Debian Admin – Ein-Zeilen-Installer.
#
#   curl -sfL https://raw.githubusercontent.com/clinicTools/setup/main/install.sh | sudo bash
#
# Lädt das passende vorgefertigte Binary der neuesten Release, prüft die
# SHA256-Summe, installiert systemd-Socket-Aktivierung + PAM-Stack und startet
# den Dienst. Erfordert root (per sudo).
set -euo pipefail

REPO="${DA_REPO:-clinicTools/setup}"
VERSION="${DA_VERSION:-latest}"
BIN_PATH="/usr/local/bin/debian-admin"
ENV_FILE="/etc/debian-admin.env"
PORT="${DA_PORT:-8088}"

log() { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
err() { printf '\033[1;31mFehler:\033[0m %s\n' "$*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || err "Bitte als root ausführen (… | sudo bash)."
command -v curl >/dev/null 2>&1 || err "curl wird benötigt."
command -v systemctl >/dev/null 2>&1 || err "systemd (systemctl) wird benötigt."

# Architektur bestimmen.
case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) err "nicht unterstützte Architektur: $(uname -m)" ;;
esac

# Neueste Version auflösen.
if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' | cut -d'"' -f4)"
  [ -n "$VERSION" ] || err "konnte neueste Version nicht ermitteln (existiert ein Release?)."
fi
log "Installiere debian-admin ${VERSION} (${ARCH})"

ASSET="debian-admin-linux-${ARCH}"
BASE="https://github.com/${REPO}/releases/download/${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

log "Lade Binary …"
curl -fSL --proto '=https' --tlsv1.2 -o "${TMP}/${ASSET}" "${BASE}/${ASSET}" \
  || err "Download fehlgeschlagen: ${BASE}/${ASSET}"

# Integrität prüfen (SHA256SUMS aus derselben Release).
if curl -fsSL -o "${TMP}/SHA256SUMS" "${BASE}/SHA256SUMS" 2>/dev/null; then
  log "Prüfe SHA256 …"
  ( cd "$TMP" && grep " ${ASSET}\$" SHA256SUMS | sha256sum -c - ) \
    || err "Prüfsumme stimmt nicht – Abbruch."
else
  log "Warnung: keine SHA256SUMS gefunden, überspringe Prüfung."
fi

install -m 0755 "${TMP}/${ASSET}" "$BIN_PATH"
log "Binary installiert: ${BIN_PATH}"

# PAM-Stack (eigener Service-Name).
cat >/etc/pam.d/debian-admin <<'PAM'
auth     required pam_unix.so
account  required pam_unix.so
PAM

# Persistentes Session-Secret erzeugen (überdauert Neustarts), falls noch keins.
if [ ! -f "$ENV_FILE" ]; then
  if command -v openssl >/dev/null 2>&1; then
    SECRET="$(openssl rand -hex 32)"
  else
    SECRET="$(head -c32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  fi
  umask 077
  cat >"$ENV_FILE" <<ENV
DA_PAM_SERVICE=debian-admin
DA_ADMIN_GROUPS=sudo,admin,wheel
DA_SESSION_SECRET=${SECRET}
# Cockpit-artiger Idle-Shutdown (Prozess endet bei Inaktivität, Socket startet neu):
DA_IDLE_TIMEOUT=90s
# Hinter Reverse-Proxy mit X-Forwarded-For auf true setzen:
# DA_TRUST_PROXY=true
# TLS direkt im Dienst (sonst per Reverse-Proxy terminieren):
# DA_TLS_CERT=/etc/debian-admin/cert.pem
# DA_TLS_KEY=/etc/debian-admin/key.pem
ENV
  log "Konfiguration erstellt: ${ENV_FILE}"
fi

# systemd Socket-Aktivierung + Service.
cat >/etc/systemd/system/debian-admin.socket <<SOCK
[Unit]
Description=Debian Admin – Socket

[Socket]
ListenStream=${PORT}
Accept=no

[Install]
WantedBy=sockets.target
SOCK

cat >/etc/systemd/system/debian-admin.service <<'SERVICE'
[Unit]
Description=Debian Admin – Web-Administrationsoberfläche
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/debian-admin.env
ExecStart=/usr/local/bin/debian-admin
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
SERVICE

log "Aktiviere Dienst …"
systemctl daemon-reload
systemctl enable --now debian-admin.socket

IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
printf '\n\033[1;32mFertig.\033[0m Debian Admin läuft (Socket-Aktivierung auf Port %s).\n\n' "$PORT"
cat <<DONE
  Aufruf:    http://${IP:-<server-ip>}:${PORT}/
  Anmeldung: mit einem System-Benutzerkonto (PAM)

Hinweise:
  - Der Zugang sollte über VPN/Reverse-Proxy + TLS abgesichert werden
    (nicht ungeschützt ins Internet stellen).
  - Konfiguration: ${ENV_FILE}   Logs: journalctl -u debian-admin
DONE
