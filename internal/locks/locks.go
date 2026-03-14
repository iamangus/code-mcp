package locks

import (
	"sync"
)

// Manager provides per-path read/write locking.
type Manager struct {
	mu    sync.Mutex
	locks map[string]*sync.RWMutex
}

// NewManager creates a new lock Manager.
func NewManager() *Manager {
	return &Manager{locks: make(map[string]*sync.RWMutex)}
}

func (m *Manager) get(path string) *sync.RWMutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.locks[path]; ok {
		return l
	}
	l := &sync.RWMutex{}
	m.locks[path] = l
	return l
}

// RLock acquires a read lock for the given path.
func (m *Manager) RLock(path string) { m.get(path).RLock() }

// RUnlock releases a read lock for the given path.
func (m *Manager) RUnlock(path string) { m.get(path).RUnlock() }

// Lock acquires a write lock for the given path.
func (m *Manager) Lock(path string) { m.get(path).Lock() }

// Unlock releases a write lock for the given path.
func (m *Manager) Unlock(path string) { m.get(path).Unlock() }
