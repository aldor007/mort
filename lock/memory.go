package lock

import (
	"sync"
	"mort/response"
)

// MemoryLock is in memory lock for single mort instance
type MemoryLock struct {
    lock    sync.RWMutex
	internal map[string]lockData

}

// NewMemoryLock create a new empty instance of MemoryLock
func NewMemoryLock() *MemoryLock  {
	m := &MemoryLock{}
	m.internal = make(map[string]lockData)
	return m
}


// NotifyAndRelease tries notify all waiting goroutines about response
func (m *MemoryLock) NotifyAndRelease(key string, res *response.Response) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result, ok := m.internal[key]
	if !ok {
		return
	}

	if res.IsBuffered() {
		resCopy, err := res.Copy()
		if err != nil {
			for _, c := range  result.responseChans {
				close(c)
			}

		} else {
			buf, err := resCopy.ReadBody()
			if err != nil {
				defer resCopy.Close()
				buf = []byte{}
			}

			for _, c := range result.responseChans {
				if c != nil {
					resCpy := response.NewBuf(res.StatusCode, buf)
					resCpy.CopyHeadersFrom(resCopy)
					c <- resCpy
					close(c)
				}
			}
		}
	} else {
		resCopy, _ := res.CopyWithStream()
		for _, c := range  result.responseChans {
			if c != nil {
				c <- resCopy
				resCopy, _ = resCopy.CopyWithStream()
				close(c)
			}
		}

	}

	delete(m.internal, key)
}

// Lock create unique entry in memory map
func (m *MemoryLock) Lock(key string) (chan *response.Response, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result, ok := m.internal[key]
	if ok {
		c := result.AddWatcher()
		m.internal[key] = result
		return c, !ok
	}

	data := lockData{}
	data.responseChans = make([]chan *response.Response, 0, 6)
	m.internal[key] = data
	return nil, !ok
}


// Release remove entry from memory map
func (m *MemoryLock) Release(key string) {
	m.lock.RLock()
	res, ok := m.internal[key]
	m.lock.RUnlock()
	if ok {
		m.lock.Lock()
		for _, c := range res.responseChans {
			close(c)
		}
		defer m.lock.Unlock()
		delete(m.internal, key)
		return
	}
}
