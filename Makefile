# Debian Admin — Build-Orchestrierung
#
# Das Frontend (Vite/Vue) wird nach internal/web/dist gebaut und per go:embed
# in das Go-Binary eingebettet. `make build` erzeugt ein einzelnes Binary.

BINARY    := debian-admin
CMD       := ./cmd/debian-admin
WEB_DIR   := web
GOFLAGS   := -trimpath
LDFLAGS   := -s -w

.PHONY: all build web go run dev test vet tidy clean

all: build

## build: Frontend + Backend zu einem einzelnen Binary bauen
build: web go

## web: Frontend mit Vite bauen (Output -> internal/web/dist)
web:
	cd $(WEB_DIR) && npm install && npm run build

## go: Go-Binary bauen (CGO für PAM erforderlich)
go:
	CGO_ENABLED=1 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(CMD)

## run: Dienst lokal starten (Dev-Modus, CORS für Vite aktiv)
run:
	DA_DEV=true DA_ALLOW_INSECURE_COOKIE=true CGO_ENABLED=1 go run $(CMD)

## dev: Frontend-Dev-Server (Hot Reload) — Backend separat via `make run`
dev:
	cd $(WEB_DIR) && npm run dev

## test: Go-Tests ausführen
test:
	CGO_ENABLED=1 go test ./...

## vet: statische Analyse
vet:
	go vet ./...

## tidy: Go-Module aufräumen
tidy:
	go mod tidy

## clean: Build-Artefakte entfernen
clean:
	rm -rf bin/ $(WEB_DIR)/node_modules
