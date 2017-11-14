package lock

import "mort/response"

// Lock is responding for collapsing request for same object

type Lock interface {
	// Lock try  get a lock for given key
	Lock(key string) (resposeChan chan *response.Response, acquired bool)
	// Release remove lock for given key
	Release(key string)
	// NotifyAndRelease remove lock for given key and notify all clients waiting for result
	NotifyAndRelease(key string, res *response.Response)
}

type lockData struct {
	Key string
	responseChans []chan *response.Response
}

func (l *lockData) AddWatcher() chan*response.Response {
	w := make(chan *response.Response)
	l.responseChans = append(l.responseChans, w)
	return w
}

