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
			close(q.ResponseChan)
			continue
		default:
		}
		resp, ok := respFactory()
		if ok {
			q.ResponseChan <- resp
		}
		close(q.ResponseChan)
	}
}

// NotifyAndRelease tries notify all waiting goroutines about response
func (m *MemoryLock) NotifyAndRelease(key string, firstResponse *response.Response) {
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

	if firstResponse.IsBuffered() {
		mirroredResponse, err := firstResponse.Copy()
		if err != nil {
			notifyListeners(lock, func() (*response.Response, bool) {
				return nil, false
			})
			return
		}
		mirroredResponseBody, err := mirroredResponse.Body()
		if err != nil {
			notifyListeners(lock, func() (*response.Response, bool) {
				return nil, false
			})
			return
		}
		go func() {
			notifyListeners(lock, func() (*response.Response, bool) {
				res := response.NewBuf(mirroredResponse.StatusCode, mirroredResponseBody)
				res.CopyHeadersFrom(mirroredResponse)
				return res, true
			})
			mirroredResponse.Close()
		}()
	} else {
		mirroredResponse, err := firstResponse.Copy()
		if err != nil {
			notifyListeners(lock, func() (*response.Response, bool) {
				return nil, false
			})
			return
		}
		go func() {
			notifyListeners(lock, func() (*response.Response, bool) {
				res, err := mirroredResponse.CopyWithStream()
				return res, err == nil
			})
			mirroredResponse.Close()
		}()
	}

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
