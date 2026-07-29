// Package web bettet das gebaute Vue-Frontend (Vite-Output) in das Go-Binary
// ein, sodass der Dienst als einzelne, abhängigkeitsarme Datei ausgeliefert
// werden kann.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

// Dist liefert das eingebettete Frontend-Verzeichnis als fs.FS.
func Dist() (fs.FS, error) {
	return fs.Sub(embedded, "dist")
}

// SPAHandler liefert statische Assets aus dem eingebetteten dist-Verzeichnis.
// Unbekannte Pfade (Client-Routen) werden auf index.html zurückgeführt, damit
// das Vue-Router-History-Modell funktioniert.
func SPAHandler() http.Handler {
	dist, err := Dist()
	if err != nil {
		// Sollte nie passieren: dist ist eingebettet.
		return http.NotFoundHandler()
	}
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			clean = "index.html"
		}
		if _, err := fs.Stat(dist, clean); err != nil {
			if os.IsNotExist(err) {
				// Fallback auf die SPA-Einstiegsseite.
				r2 := new(http.Request)
				*r2 = *r
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
