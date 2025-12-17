package lock

import (
	"context"
	"sync"

	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/response"
	"go.uber.org/zap"
)

// MemoryLock is in memory lock for single mort instance
type MemoryLock struct {
	lock     sync.RWMutex
	internal map[string]lockData
}

// NewMemoryLock create a new empty instance of MemoryLock
func NewMemoryLock() *MemoryLock {
	m := &MemoryLock{}
	m.internal = make(map[string]lockData)
	return m
}

func notifyListeners(lock lockData, respFactory func() (*response.Response, bool)) {
	for _, q := range lock.notifyQueue {
		select {
		case <-q.Cancel:
			// Observer revoked his interest in obtaining the response.
			close(q.ResponseChan)
			continue
		default:
		}
		resp, ok := respFactory()
		if ok {
			// As the response channel is under our package control
			// we are sure that it was initiated with a single place for a response
			// and there is no need to use select.
			q.ResponseChan <- resp
		}
		close(q.ResponseChan)
	}
}

// NotifyAndRelease tries notify all waiting goroutines about response.
// Uses SharedResponse to allow all waiting requests to share the same buffer
// without creating full copies, significantly reducing memory usage for
// duplicate requests (e.g., 10 waiters × 5MB = 50MB → 5MB).
func (m *MemoryLock) NotifyAndRelease(_ context.Context, key string, sharedResponse *response.SharedResponse) {
	m.lock.Lock()
	lock, ok := m.internal[key]
	if !ok {
		m.lock.Unlock()
		return
	}
	delete(m.internal, key)
	m.lock.Unlock()

	if len(lock.notifyQueue) == 0 {
		// No waiters, release the original reference
		if sharedResponse != nil {
			sharedResponse.Release()
		}
		return
	}

	monitoring.Log().Info("Notify lock queue", zap.String("key", key), zap.Int("len", len(lock.notifyQueue)))

	// Each waiter gets a shared reference (no copying!)
	// SharedResponse uses atomic reference counting to safely share the buffer
	// across all waiting goroutines. Each Acquire() increments the refcount,
	// and each consumer should Release() when done.
	notifyListeners(lock, func() (*response.Response, bool) {
		if sharedResponse == nil {
			return nil, false
		}
		// Acquire increments refcount and returns a lightweight view
		// that shares the underlying buffer
		return sharedResponse.Acquire(), true
	})

	// Release the original reference after distribution to all waiters
	sharedResponse.Release()
}

// Lock create unique entry in memory map
func (m *MemoryLock) Lock(ctx context.Context, key string) (LockResult, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.internal[key]
	result := LockResult{}
	if !ok {
		lock = lockData{}
		lock.notifyQueue = make([]LockResult, 0, 5)
	} else {
		result = lock.AddWatcher()
		// Monitor context cancellation for waiters
		if ctx != nil && result.Cancel != nil {
			go func(cancelChan chan bool, doneChan <-chan struct{}) {
				select {
				case <-doneChan:
					// Context was canceled, signal the Cancel channel
					select {
					case cancelChan <- true:
					default:
						// Cancel channel already closed or full
					}
				case <-cancelChan:
					// Cancel was already signaled by another source
					return
				}
			}(result.Cancel, ctx.Done())
		}
	}
	m.internal[key] = lock
	return result, !ok
}

func (m *MemoryLock) forceLockAndAddWatch(_ context.Context, key string) (LockResult, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.internal[key]
	result := LockResult{}
	if !ok {
		lock = lockData{}
		lock.notifyQueue = make([]LockResult, 0, 5)
		result = lock.AddWatcher()
	} else {
		result = lock.AddWatcher()
	}
	m.internal[key] = lock
	return result, !ok
}

// Release remove entry from memory map
func (m *MemoryLock) Release(_ context.Context, key string) {
	m.lock.RLock()
	_, ok := m.internal[key]
	m.lock.RUnlock()
	if ok {
		m.lock.Lock()
		defer m.lock.Unlock()
		res, exists := m.internal[key]
		if !exists {
			return
		}
		notifyListeners(res, func() (*response.Response, bool) {
			return nil, false
		})
		delete(m.internal, key)
		return
	}
}
