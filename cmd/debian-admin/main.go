// Command debian-admin startet die Web-Administrationsoberfläche für ein
// Debian-System. Die Anmeldung erfolgt gegen die Betriebssystem-Benutzer
// (PAM); privilegierte Aktionen erfordern die Mitgliedschaft in einer
// Admin-Gruppe (Standard: sudo/admin/wheel).
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clinictools/setup/internal/config"
	"github.com/clinictools/setup/internal/server"
)

func main() {
	cfg := config.Load()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.New(cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful Shutdown bei SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("debian-admin lauscht auf %s (PAM-Service=%q, Admin-Gruppen=%v)",
			cfg.Addr, cfg.PAMService, cfg.AdminGroups)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server-Fehler: %v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Println("Beende Dienst …")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown-Fehler: %v", err)
	}
}
