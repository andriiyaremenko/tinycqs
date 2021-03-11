package tinycqs

import (
	"sync"

	"github.com/stretchr/testify/assert"
)

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

type activeHandlersCounter struct {
	mu                   sync.Mutex
	count                int
	concurrentCallsLimit int
}

func (ahc *activeHandlersCounter) increase(assert *assert.Assertions) {
	ahc.mu.Lock()
	defer ahc.mu.Unlock()

	ahc.count++
	if ahc.concurrentCallsLimit <= ahc.count {
		assert.Fail("concurrency limit violation")
	}
}

func (ahc *activeHandlersCounter) decrease() {
	ahc.mu.Lock()
	defer ahc.mu.Unlock()

	ahc.count--
}
