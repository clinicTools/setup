package system

import (
	"strings"
)

// PackageSummary fasst den Zustand der Paketverwaltung zusammen.
type PackageSummary struct {
	Installed   int       `json:"installed"`
	Upgradable  int       `json:"upgradable"`
	Upgrades    []Package `json:"upgrades"`
	LastUpdated string    `json:"lastUpdated,omitempty"`
}

// Package beschreibt ein installiertes oder aktualisierbares Paket.
type Package struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	AvailableVersion string `json:"availableVersion,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
}

// Packages liefert eine Zusammenfassung über dpkg/apt: Anzahl installierter
// Pakete sowie die Liste verfügbarer Upgrades.
func Packages() (*PackageSummary, error) {
	sum := &PackageSummary{}

	if commandExists("dpkg-query") {
		out, err := run("dpkg-query", "-f", "${binary:Package}\n", "-W")
		if err == nil {
			sum.Installed = len(lines(out))
		}
	}

	sum.Upgrades = upgradablePackages()
	sum.Upgradable = len(sum.Upgrades)
	return sum, nil
}

// InstalledPackages liefert die vollständige Liste installierter Pakete.
func InstalledPackages() ([]Package, error) {
	if !commandExists("dpkg-query") {
		return nil, errMissingTool("dpkg-query")
	}
	out, err := run("dpkg-query", "-f", "${binary:Package}\t${Version}\t${Architecture}\n",
		"-W")
	if err != nil {
		return nil, err
	}
	var pkgs []Package
	for _, l := range lines(out) {
		f := strings.Split(l, "\t")
		if len(f) < 2 {
			continue
		}
		p := Package{Name: f[0], Version: f[1]}
		if len(f) >= 3 {
			p.Architecture = f[2]
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

// upgradablePackages parst `apt list --upgradable`.
func upgradablePackages() []Package {
	if !commandExists("apt") {
		return nil
	}
	out, err := run("apt", "list", "--upgradable")
	if err != nil {
		return nil
	}
	var pkgs []Package
	for _, l := range lines(out) {
		// Format: name/suite newversion arch [upgradable from: oldversion]
		if !strings.Contains(l, "[upgradable") {
			continue
		}
		nameField, rest, ok := strings.Cut(l, "/")
		if !ok {
			continue
		}
		// rest = "suite version arch [upgradable from: old]"
		fields := strings.Fields(rest)
		p := Package{Name: nameField}
		if len(fields) >= 2 {
			p.AvailableVersion = fields[1]
		}
		if len(fields) >= 3 {
			p.Architecture = fields[2]
		}
		if idx := strings.Index(l, "from:"); idx != -1 {
			tail := strings.TrimSpace(l[idx+len("from:"):])
			p.Version = strings.TrimRight(tail, "]")
			p.Version = strings.TrimSpace(p.Version)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// AptUpdate aktualisiert die Paketlisten (apt-get update).
func AptUpdate() (string, error) {
	if !commandExists("apt-get") {
		return "", errMissingTool("apt-get")
	}
	return run("apt-get", "update")
}

// AptUpgradeCount liefert nur die Anzahl aktualisierbarer Pakete (schnell).
func AptUpgradeCount() int {
	return len(upgradablePackages())
}
