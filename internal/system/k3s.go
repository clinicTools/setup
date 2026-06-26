package system

import (
	"encoding/json"
	"os"
	"strings"
)

// K3sStatus beschreibt den Installations- und Laufzustand von k3s.
type K3sStatus struct {
	Installed   bool   `json:"installed"`
	Active      bool   `json:"active"`
	Version     string `json:"version,omitempty"`
	NodeName    string `json:"nodeName,omitempty"`
	HasUninstall bool  `json:"hasUninstall"`
}

// K3sNode ist eine vereinfachte Sicht auf einen Cluster-Knoten.
type K3sNode struct {
	Name    string   `json:"name"`
	Ready   bool     `json:"ready"`
	Roles   []string `json:"roles"`
	Version string   `json:"version"`
	IP      string   `json:"ip,omitempty"`
}

// K3sPod ist eine vereinfachte Sicht auf einen Pod.
type K3sPod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Ready     string `json:"ready"` // z. B. "1/1"
	Node      string `json:"node,omitempty"`
}

const (
	k3sKubeconfigPath = "/etc/rancher/k3s/k3s.yaml"
	k3sTokenPath      = "/var/lib/rancher/k3s/server/node-token"
	k3sUninstallPath  = "/usr/local/bin/k3s-uninstall.sh"
)

// K3sInstalled meldet, ob das k3s-Binary vorhanden ist.
func K3sInstalled() bool { return commandExists("k3s") }

// K3sGetStatus liefert den aktuellen k3s-Zustand.
func K3sGetStatus(r *Runner) *K3sStatus {
	st := &K3sStatus{Installed: K3sInstalled()}
	if !st.Installed {
		return st
	}
	if out, err := r.run("k3s", "--version"); err == nil {
		st.Version = firstLine(out)
	}
	if commandExists("systemctl") {
		if out, _ := r.run("systemctl", "is-active", "k3s"); out == "active" {
			st.Active = true
		}
	}
	if h, err := osHostname(); err == nil {
		st.NodeName = h
	}
	if _, err := os.Stat(k3sUninstallPath); err == nil {
		st.HasUninstall = true
	}
	return st
}

// K3sNodes liefert die Knoten des Clusters (k3s kubectl get nodes). Erfordert
// Zugriff auf die kubeconfig (root) und läuft daher privilegiert.
func K3sNodes(r *Runner) ([]K3sNode, error) {
	if !K3sInstalled() {
		return nil, errMissingTool("k3s")
	}
	out, err := r.sudo("k3s", "kubectl", "get", "nodes", "-o", "json")
	if err != nil {
		return nil, err
	}
	var list struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Status struct {
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
				NodeInfo struct {
					KubeletVersion string `json:"kubeletVersion"`
				} `json:"nodeInfo"`
				Addresses []struct {
					Type    string `json:"type"`
					Address string `json:"address"`
				} `json:"addresses"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, err
	}

	nodes := make([]K3sNode, 0, len(list.Items))
	for _, it := range list.Items {
		n := K3sNode{Name: it.Metadata.Name, Version: it.Status.NodeInfo.KubeletVersion}
		for _, c := range it.Status.Conditions {
			if c.Type == "Ready" {
				n.Ready = c.Status == "True"
			}
		}
		for label := range it.Metadata.Labels {
			if role, ok := strings.CutPrefix(label, "node-role.kubernetes.io/"); ok && role != "" {
				n.Roles = append(n.Roles, role)
			}
		}
		for _, addr := range it.Status.Addresses {
			if addr.Type == "InternalIP" {
				n.IP = addr.Address
			}
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// K3sPods liefert alle Pods aller Namespaces (privilegiert).
func K3sPods(r *Runner) ([]K3sPod, error) {
	if !K3sInstalled() {
		return nil, errMissingTool("k3s")
	}
	out, err := r.sudo("k3s", "kubectl", "get", "pods", "-A", "-o", "json")
	if err != nil {
		return nil, err
	}
	var list struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready bool `json:"ready"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, err
	}

	pods := make([]K3sPod, 0, len(list.Items))
	for _, it := range list.Items {
		ready := 0
		for _, c := range it.Status.ContainerStatuses {
			if c.Ready {
				ready++
			}
		}
		pods = append(pods, K3sPod{
			Namespace: it.Metadata.Namespace,
			Name:      it.Metadata.Name,
			Phase:     it.Status.Phase,
			Ready:     itoa(ready) + "/" + itoa(len(it.Status.ContainerStatuses)),
			Node:      it.Spec.NodeName,
		})
	}
	return pods, nil
}

// K3sKubeconfig liefert die kubeconfig (für externen Zugriff). Die Datei gehört
// root; gelesen wird daher privilegiert im Benutzerkontext.
func K3sKubeconfig(r *Runner) (string, error) {
	return r.sudo("cat", k3sKubeconfigPath)
}

// K3sNodeToken liefert den Join-Token für Agent-Knoten (privilegiert gelesen).
func K3sNodeToken(r *Runner) (string, error) {
	out, err := r.sudo("cat", k3sTokenPath)
	return strings.TrimSpace(out), err
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
