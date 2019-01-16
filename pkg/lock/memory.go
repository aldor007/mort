package lock

import (
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/response"
	"go.uber.org/zap"
	"sync"
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

// NotifyAndRelease tries notify all waiting goroutines about response
func (m *MemoryLock) NotifyAndRelease(key string, res *response.Response) {
	m.lock.Lock()
	result, ok := m.internal[key]
	if !ok {
		m.lock.Unlock()
		return
	}

	delete(m.internal, key)
	m.lock.Unlock()

	if len(result.notifyQueue) == 0 {
		return
	}

	monitoring.Log().Warn("Notify queue", zap.String("key", key), zap.Int("len", len(result.notifyQueue)))

	if res.IsBuffered() {
		resCopy, err := res.Copy()
		if err != nil {
			for _, q := range result.notifyQueue {
				close(q.ResponseChan)
			}

		} else {
			buf, err := resCopy.ReadBody()
			if err != nil {
				for _, q := range result.notifyQueue {
					close(q.ResponseChan)
				}
				return
			}

			go func() {
				for _, q := range result.notifyQueue {
					resCpy := response.NewBuf(resCopy.StatusCode, buf)
					resCpy.CopyHeadersFrom(resCopy)
					select {
					case <-q.Cancel:
						close(q.ResponseChan)
						continue
					default:

					}
					q.ResponseChan <- resCpy
					close(q.ResponseChan)
				}

				resCopy.Close()
			}()

		}
	} else {
		resCopy, _ := res.Copy()
		go func() {
			for _, q := range result.notifyQueue {
				resCpy, _ := resCopy.CopyWithStream()
				select {
				case <-q.Cancel:
					close(q.ResponseChan)
				case q.ResponseChan <- resCpy:
					close(q.ResponseChan)
				default:
				}
			}
			resCopy.Close()
		}()
	}

}

// Lock create unique entry in memory map
func (m *MemoryLock) Lock(key string) (LockResult, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result, ok := m.internal[key]
	if ok {
		r := result.AddWatcher()
		m.internal[key] = result
		return r, !ok
	}

	data := lockData{}
	data.notifyQueue = make([]LockResult, 0, 5)
	m.internal[key] = data
	return LockResult{}, !ok
}

// Release remove entry from memory map
func (m *MemoryLock) Release(key string) {
	m.lock.RLock()
	res, ok := m.internal[key]
	m.lock.RUnlock()
	if ok {
		m.lock.Lock()
		for _, q := range res.notifyQueue {
			close(q.ResponseChan)
		}
		defer m.lock.Unlock()
		delete(m.internal, key)
		return
	}
}
