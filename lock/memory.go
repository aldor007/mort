package lock

import (
	"mort/response"
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
		return
	}

	delete(m.internal, key)
	m.lock.Unlock()

	if len(result.notifyQueue) == 0 {
		return
	}

	if res.IsBuffered() {
		resCopy, err := res.Copy()
		if err != nil {
			for _, q := range result.notifyQueue {
				close(q.ResponseChan)
			}

		} else {
			buf, err := resCopy.ReadBody()
			if err != nil {
				defer resCopy.Close()
				buf = []byte{}
			}

			for _, q := range result.notifyQueue {
				select {
				case <-q.Cancel:
					close(q.ResponseChan)
				default:
					resCpy := response.NewBuf(resCopy.StatusCode, buf)
					resCpy.CopyHeadersFrom(resCopy)
					q.ResponseChan <- resCpy
					close(q.ResponseChan)

				}
			}
		}
	} else {
		resCopy, _ := res.CopyWithStream()
		for _, q := range result.notifyQueue {
			select {
			case <-q.Cancel:
				close(q.ResponseChan)
			default:
				resCpy, _ := resCopy.CopyWithStream()
				q.ResponseChan <- resCpy
				close(q.ResponseChan)
			}
		}

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
