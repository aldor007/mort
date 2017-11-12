package throttler

import (
	"time"
)

var defaultBacklogTimeout = time.Second * 60

type Throttler struct {
	tokens chan bool 
	backlogTokens chan bool 
	backlogTimeout time.Duration
}


func New(limit int) *Throttler {
	return NewBacklog(limit, 0, defaultBacklogTimeout)
}

func NewBacklog(limit int, backlog int, timeout time.Duration) *Throttler {
	max := limit + backlog
	t := &Throttler{
		tokens: make(chan bool, limit),
		backlogTokens: make(chan bool, max),
		backlogTimeout: timeout,
	}
	
	for i:= 0; i < max; i++ {
		if i < limit {
			t.tokens <- true 
		}
		t.backlogTokens <- true

	}
	return t
}

func (t *Throttler) Take()  bool {
	select {
	case btok := <-t.backlogTokens:
		timer := time.NewTimer(t.backlogTimeout)

		defer func() {
			t.backlogTokens <- btok
		}()

		select {
		case <-timer.C:
			return false
		case <-t.tokens:
			return true
		}
		return false
	default:
		return false
	}
}

func (t *Throttler) Release()  {
	t.tokens <- true
}
