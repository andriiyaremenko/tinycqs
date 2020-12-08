package command

import (
	"context"
	"errors"
	"sync"
)

var (
	WorkerStopped = errors.New("command worker is stopped")
)

func NewWorker(ctx context.Context, eventSink func(Event), commands Commands) CommandsWorker {
	w := &worker{
		ctx:       ctx,
		started:   false,
		eventSink: eventSink,
		commands:  commands}

	w.start()

	return w
}

type worker struct {
	ctx context.Context
	mu  sync.RWMutex

	started   bool
	commands  Commands
	eventPipe chan Event
	eventSink func(event Event)
}

func (w *worker) Handle(event Event) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.started {
		return WorkerStopped
	}

	w.eventPipe <- event

	return nil
}

func (w *worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.started
}

func (w *worker) start() {
	if w.started {
		return
	}

	w.mu.Lock()

	w.eventPipe = make(chan Event, w.commands.ConcurrencyLimit())
	w.started = true

	w.mu.Unlock()

	go func() {
		concurrencyLimit := w.commands.ConcurrencyLimit()
		var wg sync.WaitGroup

		for {
			select {
			case <-w.ctx.Done():
				w.mu.Lock()

				wg.Wait()

				w.started = false
				close(w.eventPipe)

				w.mu.Unlock()

				return
			case event := <-w.eventPipe:
				wg.Add(1)
				go func() {
					defer wg.Done()

					w.eventSink(w.commands.Handle(w.ctx, event))
				}()
			default:
				newLimit := w.commands.ConcurrencyLimit()
				if concurrencyLimit == newLimit {
					continue
				}

				w.mu.Lock()

				wg.Wait()

				close(w.eventPipe)

				w.eventPipe = make(chan Event, newLimit)
				concurrencyLimit = newLimit

				w.mu.Unlock()
			}
		}
	}()
}

func (w *worker) sealed() {}
