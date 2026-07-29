// Command debian-admin startet die Web-Administrationsoberfläche für ein
// Debian-System. Die Anmeldung erfolgt gegen die Betriebssystem-Benutzer
// (PAM); privilegierte Aktionen erfordern die Mitgliedschaft in einer
// Admin-Gruppe (Standard: sudo/admin/wheel).
//
// Wie Cockpit unterstützt der Dienst systemd-Socket-Activation und einen
// optionalen Idle-Shutdown: Wird er per Socket gestartet und ist DA_IDLE_TIMEOUT
// gesetzt, beendet er sich nach Ablauf der Inaktivität und wird bei der nächsten
// Verbindung automatisch neu gestartet.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clinictools/setup/internal/config"
	"github.com/clinictools/setup/internal/server"
	"github.com/clinictools/setup/internal/ws"
	"github.com/coreos/go-systemd/v22/activation"
)

func main() {
	cfg := config.Load()
	handler, activity := server.New(cfg)

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, viaSocket, err := listen(cfg.Addr)
	if err != nil {
		log.Fatalf("Listen fehlgeschlagen: %v", err)
	}

	tlsEnabled := cfg.TLSCert != "" && cfg.TLSKey != ""
	go func() {
		log.Printf("debian-admin lauscht auf %s (TLS=%v, Socket-Activation=%v, Idle-Timeout=%s, PAM=%q)",
			listener.Addr(), tlsEnabled, viaSocket, cfg.IdleTimeout, cfg.PAMService)
		var err error
		if tlsEnabled {
			err = srv.ServeTLS(listener, cfg.TLSCert, cfg.TLSKey)
		} else {
			err = srv.Serve(listener)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server-Fehler: %v", err)
			os.Exit(1)
		}
	}()

	// Cockpit-artiger Idle-Shutdown.
	if cfg.IdleTimeout > 0 {
		go watchIdle(ctx, stop, activity, cfg.IdleTimeout)
	}

	<-ctx.Done()
	log.Println("Beende Dienst …")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown-Fehler: %v", err)
	}
}

// listen liefert entweder den von systemd geerbten Socket (Socket-Activation)
// oder öffnet selbst einen Listener auf cfg.Addr.
func listen(addr string) (net.Listener, bool, error) {
	if listeners, err := activation.Listeners(); err == nil && len(listeners) > 0 {
		return listeners[0], true, nil
	}
	l, err := net.Listen("tcp", addr)
	return l, false, err
}

// watchIdle beendet den Dienst, sobald für idle keine Aktivität mehr vorlag.
func watchIdle(ctx context.Context, stop context.CancelFunc, activity *ws.Activity, idle time.Duration) {
	// Häufig genug prüfen, um zeitnah zu reagieren, ohne zu pollen.
	interval := idle / 3
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if activity.Idle(idle) {
				log.Printf("Keine Verbindung seit %s — beende Dienst (Socket-Activation startet bei Bedarf neu).", idle)
				stop()
				return
			}
		}
	}
}
