package lock

import "mort/response"

type Lock interface {
	Lock(key string) (chan *response.Response, bool)
	Release(key string)
	NotifyAndRelease(key string, res *response.Response)
}

type LockData struct {
	Key string
	responseChans []chan *response.Response
}

func (l *LockData) AddWatcher() chan*response.Response {
	w := make(chan *response.Response)
	l.responseChans = append(l.responseChans, w)
	return w
}

