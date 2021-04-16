package command

import (
	"context"
	"sync"

	"github.com/andriiyaremenko/tinycqs/tracing"
	"github.com/google/uuid"
)

func newEventRW(ctx context.Context) EventReader {
	ctx, cancel := context.WithCancel(ctx)
	return &eventRW{
		ctx:           ctx,
		cancel:        cancel,
		closed:        false,
		eventsChannel: make(chan EventWithMetadata, 1)}
}

type eventRW struct {
	ctx    context.Context
	cancel context.CancelFunc

	eventsChannel chan EventWithMetadata
	closed        bool
	closedWRMutex sync.RWMutex
	once          sync.Once
}

func (r *eventRW) Read() <-chan EventWithMetadata {
	return r.eventsChannel
}

func (r *eventRW) Close() {
	r.once.Do(func() {
		r.closedWRMutex.Lock()

		r.closed = true

		r.closedWRMutex.Unlock()
		r.cancel()
		close(r.eventsChannel)
	})
}

func (r *eventRW) GetWriter(metadata tracing.Metadata) EventWriter {
	events := make(chan EventWithMetadata, 1)
	w := &eventW{eventRW: r, metadata: metadata, events: events}

	go func() {
		for e := range events {
			r.closedWRMutex.RLock()

			if r.closed {
				r.closedWRMutex.RUnlock()

				return
			}

			r.eventsChannel <- e
			r.closedWRMutex.RUnlock()
		}
	}()

	return w
}

type eventW struct {
	isDone bool
	once   sync.Once
	rwMu   sync.RWMutex

	writeWG      sync.WaitGroup
	writeWGMutex sync.Mutex

	eventRW  *eventRW
	metadata tracing.Metadata

	events chan EventWithMetadata
}

func (r *eventW) Write(e Event) {
	r.writeWG.Add(1)
	go func() {
		r.rwMu.RLock()
		defer r.rwMu.RUnlock()
		defer r.writeWG.Done()

		if !r.isDone {
			if withMetadata := AsEventWithMetadata(e); withMetadata != nil {
				r.events <- WithMetadata(e, withMetadata.Metadata())
				return
			}

			id := uuid.New().String()
			withMetadata := WithMetadata(e, r.metadata.New(id))

			r.events <- withMetadata
		}
	}()
}

func (r *eventW) Done() {
	go r.once.Do(func() {
		r.writeWG.Wait()
		r.rwMu.Lock()

		r.isDone = true
		r.events <- doneWriting

		close(r.events)
		r.rwMu.Unlock()
	})
}
