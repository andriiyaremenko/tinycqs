package command

import (
	"context"
	"sync"
)

func newEventRW(ctx context.Context, limit int) *eventRW {
	ctx, cancel := context.WithCancel(ctx)
	return &eventRW{ctx: ctx, cancel: cancel, ch: make(chan Event, limit)}
}

type eventRW struct {
	ctx    context.Context
	cancel context.CancelFunc

	ch             chan Event
	once           sync.Once
	writersWG      sync.WaitGroup
	writersWGMutex sync.Mutex
}

func (r *eventRW) Write(e Event) {
	select {
	case <-r.ctx.Done():
		r.Close()

		return
	default:
	}

	r.writersWGMutex.Lock()
	r.writersWG.Add(1)
	defer r.writersWG.Done()
	defer r.writersWGMutex.Unlock()

	select {
	case <-r.ctx.Done():
		r.Close()

		return
	case r.ch <- e:
		return
	}
}

func (r *eventRW) Read() <-chan Event {
	return r.ch
}

func (r *eventRW) Done() {
	r.Write(doneWriting)
}

func (r *eventRW) Close() {
	r.once.Do(func() {
		r.cancel()

		r.writersWGMutex.Lock()
		r.writersWG.Wait()
		close(r.ch)
		r.writersWGMutex.Unlock()
	})
}
