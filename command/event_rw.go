package command

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

func newEventRW(ctx context.Context, limit int) EventReader {
	ctx, cancel := context.WithCancel(ctx)
	return &eventRW{ctx: ctx, cancel: cancel, ch: make(chan EventWithMetadata, limit)}
}

type eventRW struct {
	ctx    context.Context
	cancel context.CancelFunc

	ch             chan EventWithMetadata
	once           sync.Once
	writersWG      sync.WaitGroup
	writersWGMutex sync.Mutex
}

func (r *eventRW) Read() <-chan EventWithMetadata {
	return r.ch
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

func (r *eventRW) write(e EventWithMetadata) {
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

func (r *eventRW) done() {
	r.write(doneWriting)
}

func (r *eventRW) GetWriter(metadata Metadata) EventWriter {
	return &eventW{eventRW: r, metadata: metadata}
}

type eventW struct {
	isDone bool
	once   sync.Once
	mu     sync.Mutex

	eventRW  *eventRW
	metadata Metadata
}

func (r *eventW) Write(e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isDone {
		if withMetadata := AsEventWithMetadata(e); withMetadata != nil {
			r.eventRW.write(WithMetadata(e, withMetadata.Metadata()))
			return
		}

		id := uuid.New().String()
		withMetadata := WithMetadata(e, r.metadata.New(id))

		r.eventRW.write(withMetadata)
	}
}

func (r *eventW) Done() {
	r.once.Do(func() {
		r.mu.Lock()

		r.isDone = true

		r.eventRW.done()
		r.mu.Unlock()
	})
}
