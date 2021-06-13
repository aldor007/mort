package lock

import (
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

// NotifyAndRelease tries notify all waiting goroutines about response
func (m *MemoryLock) NotifyAndRelease(key string, originalResponse *response.Response) {
	m.lock.Lock()
	lock, ok := m.internal[key]
	if !ok {
		m.lock.Unlock()
		return
	}
	delete(m.internal, key)
	m.lock.Unlock()

	if len(lock.notifyQueue) == 0 {
		return
	}

	monitoring.Log().Info("Notify lock queue", zap.String("key", key), zap.Int("len", len(lock.notifyQueue)))
	// Notify all listeners by sending them a copy of originalResponse.
	//
	// Current synchronous notification is simpler compared to asynchronous implementation.
	// The asynchronous implementation might be tricky since the response in not buffered mode must be
	// protected from being read before it is copied. Otherwise CopyWithStream in a worst case will deliver partial body
	// since it can read in parallel with HTTP handler. To prevent such behaviour extra temporary copy of response
	// must be created before returning from this method. Of course such creation must
	// also take into account whether the originalResponse is buffered or not.
	// The time spend on notifying listeners is negligible compared to the total time of image processing,
	// so making this process asynchronous makes almost no sense.
	notifyListeners(lock, func() (*response.Response, bool) {
		if originalResponse.IsBuffered() {
			res, err := originalResponse.Copy()
			return res, err == nil
		} else {
			res, err := originalResponse.CopyWithStream()
			return res, err == nil
		}
	})
}

// Lock create unique entry in memory map
func (m *MemoryLock) Lock(key string) (LockResult, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lock, ok := m.internal[key]
	result := LockResult{}
	if !ok {
		lock = lockData{}
		lock.notifyQueue = make([]LockResult, 0, 5)
	} else {
		result = lock.AddWatcher()
	}
	m.internal[key] = lock
	return result, !ok
}

// Release remove entry from memory map
func (m *MemoryLock) Release(key string) {
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
