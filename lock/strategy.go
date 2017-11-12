package lock

import (
	"sync"
	"mort/response"
)

type Lock interface {
	Lock(key string) (LockData, bool)
	Release(key string)
	Counter(key string) int
}

type LockData struct {
	Key string
	ResponseChan chan *response.Response
}

type MemoryLock struct {
    lock    sync.RWMutex
    counterLock sync.RWMutex
	internal map[string]LockData
	counter  map[string]int

}

func NewMemoryLock() *MemoryLock  {
	m := &MemoryLock{}
	m.internal = make(map[string]LockData)
	m.counter = make(map[string]int)
	return m
}

func (m *MemoryLock) Lock(key string) (LockData, bool) {
	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	c, cOk := m.counter[key]
	if cOk {
		m.counter[key] = c + 1
	} else {
		m.counter[key] = 0
	}

	m.lock.RLock()
	result, ok := m.internal[key]
	m.lock.RUnlock()
	if ok {
		return result, !ok
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	data := LockData{}
	data.ResponseChan = make(chan *response.Response, 3)
	data.Key = key
	m.internal[key] = data
	return data, !ok
}

func (m *MemoryLock) Release(key string) {
	m.lock.RLock()
	res, ok := m.internal[key]
	m.lock.RUnlock()
	if ok {
		m.lock.Lock()
		close(res.ResponseChan)
		defer m.lock.Unlock()
		delete(m.internal, key)
		return
	}
}

func (m *MemoryLock) Counter(key string) int {
	m.counterLock.RLock()
	defer m.counterLock.RUnlock()
	return m.counter[key]
}