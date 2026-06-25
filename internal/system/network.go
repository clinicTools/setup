package system

import (
	"net"
	"sort"
	"strings"
)

// NetworkInterface beschreibt eine Netzwerkschnittstelle inkl. Adressen.
type NetworkInterface struct {
	Name       string   `json:"name"`
	MAC        string   `json:"mac"`
	MTU        int      `json:"mtu"`
	Up         bool     `json:"up"`
	Loopback   bool     `json:"loopback"`
	Addresses  []string `json:"addresses"`
}

// NetworkInfo bündelt Schnittstellen, DNS-Server und Default-Gateway.
type NetworkInfo struct {
	Hostname   string             `json:"hostname"`
	Interfaces []NetworkInterface `json:"interfaces"`
	DNS        []string           `json:"dns"`
	Gateway    string             `json:"gateway,omitempty"`
}

// Network sammelt den Netzwerkzustand über das net-Paket und /etc/resolv.conf.
func Network() (*NetworkInfo, error) {
	info := &NetworkInfo{}
	if h, err := osHostname(); err == nil {
		info.Hostname = h
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, ifi := range ifaces {
		ni := NetworkInterface{
			Name:     ifi.Name,
			MAC:      ifi.HardwareAddr.String(),
			MTU:      ifi.MTU,
			Up:       ifi.Flags&net.FlagUp != 0,
			Loopback: ifi.Flags&net.FlagLoopback != 0,
		}
		addrs, _ := ifi.Addrs()
		for _, a := range addrs {
			ni.Addresses = append(ni.Addresses, a.String())
		}
		info.Interfaces = append(info.Interfaces, ni)
	}
	sort.Slice(info.Interfaces, func(i, j int) bool {
		return info.Interfaces[i].Name < info.Interfaces[j].Name
	})

	info.DNS = resolvConfNameservers()
	info.Gateway = defaultGateway()
	return info, nil
}

func resolvConfNameservers() []string {
	data, err := readFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	var out []string
	for _, l := range lines(data) {
		if strings.HasPrefix(l, "nameserver") {
			fields := strings.Fields(l)
			if len(fields) >= 2 {
				out = append(out, fields[1])
			}
		}
	}
	return out
}

// defaultGateway liest das Default-Gateway via `ip route` (falls vorhanden).
func defaultGateway() string {
	if !commandExists("ip") {
		return ""
	}
	out, err := run("ip", "route", "show", "default")
	if err != nil {
		return ""
	}
	for _, l := range lines(out) {
		fields := strings.Fields(l)
		for i, f := range fields {
			if f == "via" && i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}
	return ""
}
