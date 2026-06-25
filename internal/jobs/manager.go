package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
)

// Manager hält alle Jobs der Laufzeit im Speicher.
type Manager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewManager erzeugt einen leeren Job-Manager.
func NewManager() *Manager {
	return &Manager{jobs: map[string]*Job{}}
}

// Start legt einen Job an und startet das Kommando asynchron. Die Argumente
// werden vom Aufrufer (Server) zusammengestellt — niemals direkt aus
// Client-Eingaben, um Command-Injection auszuschließen.
func (m *Manager) Start(name, command string, args []string) *Job {
	return m.StartEnv(name, command, args, nil)
}

// StartEnv verhält sich wie Start, ergänzt aber die Prozessumgebung (z. B.
// DEBIAN_FRONTEND=noninteractive für apt).
func (m *Manager) StartEnv(name, command string, args, env []string) *Job {
	job := newJob(generateID(), name)

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	go job.run(context.Background(), command, args, env)
	return job
}

// Get liefert einen Job per ID.
func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	return j, ok
}

// List liefert alle Jobs, neueste zuerst.
func (m *Manager) List() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, j)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt > out[j].StartedAt })
	return out
}

func generateID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
