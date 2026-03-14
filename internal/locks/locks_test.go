package locks

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestManager_BasicReadWrite(t *testing.T) {
	m := NewManager()
	path := "/some/path"

	m.Lock(path)
	m.Unlock(path)

	m.RLock(path)
	m.RUnlock(path)
}

func TestManager_MultipleReaders(t *testing.T) {
	m := NewManager()
	path := "/concurrent/read"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RLock(path)
			defer m.RUnlock(path)
		}()
	}
	wg.Wait()
}

func TestManager_WriteExcludesReaders(t *testing.T) {
	m := NewManager()
	path := "/exclusive/write"

	var active int32
	var wg sync.WaitGroup

	// Start a writer
	m.Lock(path)

	// Try readers concurrently – they should block until writer is done
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RLock(path)
			atomic.AddInt32(&active, 1)
			m.RUnlock(path)
		}()
	}

	// Readers should still be blocked
	if v := atomic.LoadInt32(&active); v != 0 {
		t.Errorf("expected 0 active readers while write lock held, got %d", v)
	}

	m.Unlock(path)
	wg.Wait()

	if v := atomic.LoadInt32(&active); v != 5 {
		t.Errorf("expected 5 completed readers, got %d", v)
	}
}

func TestManager_SeparatePaths(t *testing.T) {
	m := NewManager()

	// Locking different paths should not block each other
	m.Lock("/path/a")
	m.Lock("/path/b") // should not deadlock
	m.Unlock("/path/b")
	m.Unlock("/path/a")
}

func TestManager_GetCreatesLock(t *testing.T) {
	m := NewManager()
	path := "/new/path"

	// Two calls to get should return the same mutex
	l1 := m.get(path)
	l2 := m.get(path)
	if l1 != l2 {
		t.Error("expected same mutex for same path")
	}
}
