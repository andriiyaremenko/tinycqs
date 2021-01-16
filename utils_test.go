package tinycqs

import "sync"

type wasCalledCounter struct {
	mu    sync.Mutex
	count int
}

func (cc *wasCalledCounter) increase() {
	cc.mu.Lock()
	cc.count++
	cc.mu.Unlock()
}

func (cc *wasCalledCounter) getCount() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.count
}
