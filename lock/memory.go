package lock

import (
	"sync"
	"mort/response"
)

type MemoryLock struct {
    lock    sync.RWMutex
	internal map[string]LockData

}

func NewMemoryLock() *MemoryLock  {
	m := &MemoryLock{}
	m.internal = make(map[string]LockData)
	return m
}


func (m *MemoryLock) NotifyAndRelease(key string, res *response.Response) {
	buf, err := res.ReadBody()
	if err != nil {
		defer res.Close()
		buf = []byte{}
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	result, ok := m.internal[key]
	if !ok {
		return
	}

	for _, c := range  result.responseChans {
		resCpy := response.NewBuf(res.StatusCode, buf)
		resCpy.CopyHeadersFrom(res)
		c <- resCpy
		close(c)
	}

	delete(m.internal, key)
}

func (m *MemoryLock) Lock(key string) (chan *response.Response, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result, ok := m.internal[key]
	if ok {
		c := result.AddWatcher()
		m.internal[key] = result
		return c, !ok
	}

	data := LockData{}
	data.responseChans = make([]chan *response.Response, 0, 4)
	m.internal[key] = data
	return nil, !ok
}

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

