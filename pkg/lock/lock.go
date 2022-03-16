package lock

import (
	"context"

	"github.com/aldor007/mort/pkg/config"
	"github.com/aldor007/mort/pkg/monitoring"
	"github.com/aldor007/mort/pkg/response"
	"go.uber.org/zap"
)

// Lock is responding for collapsing request for same object
type Lock interface {
	// Lock try  get a lock for given key
	Lock(ctx context.Context, key string) (observer LockResult, acquired bool)
	// Release remove lock for given key
	Release(ctx context.Context, key string)
	// NotifyAndRelease remove lock for given key and notify all clients waiting for result
	NotifyAndRelease(ctx context.Context, key string, res *response.Response)
}

// LockResult contain struct
type LockResult struct {
	ResponseChan chan *response.Response // channel on which you get response
	Cancel       chan bool               // channel for notify about cancel of waiting
	Error        error                   // error when creating error
}

type lockData struct {
	Key         string
	notifyQueue []LockResult
}

// AddWatcher add next request waiting for lock to expire or return result
func (l *lockData) AddWatcher() LockResult {
	d := LockResult{}
	d.ResponseChan = make(chan *response.Response, 1)
	d.Cancel = make(chan bool, 1)
	l.notifyQueue = append(l.notifyQueue, d)
	return d
}

// NewNopLock create lock that do nothing
func NewNopLock() *NopLock {
	return &NopLock{}
}

// NopLock will never  collapse any request
type NopLock struct {
}

// Lock always return that lock was acquired
func (l *NopLock) Lock(_ context.Context, _ string) (LockResult, bool) {
	return LockResult{}, true
}

// Release do nothing
func (l *NopLock) Release(_ context.Context, _ string) {

}

// NotifyAndRelease do nothing
func (l *NopLock) NotifyAndRelease(_ context.Context, _ string, _ *response.Response) {

}

func Create(lockCfg *config.LockCfg, lockTimeout int) Lock {
	if lockCfg == nil {
		monitoring.Log().Info("Creating memory lock")
		return NewMemoryLock()
	}
	switch lockCfg.Type {
	case "redis":
		monitoring.Log().Info("Creating redis lock", zap.Strings("addr", lockCfg.Address), zap.Int("lockTimeout", lockTimeout))
		r := NewRedisLock(lockCfg.Address, lockCfg.ClientConfig)
		r.LockTimeout = lockTimeout
		return r
	case "redis-cluster":
		monitoring.Log().Info("Creating redis-cluster lock", zap.Strings("addr", lockCfg.Address), zap.Int("lockTimeout", lockTimeout))
		r := NewRedisCluster(lockCfg.Address, lockCfg.ClientConfig)
		r.LockTimeout = lockTimeout
		return r

	default:
		return NewMemoryLock()
	}
}
