package command

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

var (
	WorkerStopped = errors.New("command worker is stopped")
)

// Returns CommandWorker based on Commands.
// eventSink is used to channel all unhandled errors in form of Event.
func NewWorker(ctx context.Context, eventSink func(CommandsWorker, Event), commands Commands, limit int) CommandsWorker {
	if limit < 1 {
		limit = 1
	}

	w := &worker{
		ctx:       ctx,
		started:   false,
		eventSink: eventSink,
		commands:  commands,
		cLimit:    limit}

	w.start()

	return w
}

type worker struct {
	ctx  context.Context
	rwMu sync.RWMutex

	started   bool
	commands  Commands
	eventPipe chan Event
	eventSink func(CommandsWorker, Event)
	cLimit    int
}

func (w *worker) Handle(event Event) error {
	w.rwMu.RLock()
	defer w.rwMu.RUnlock()

	if !w.started {
		return WorkerStopped
	}

	w.eventPipe <- event

	return nil
}

func (w *worker) IsRunning() bool {
	w.rwMu.RLock()
	defer w.rwMu.RUnlock()

	return w.started
}

func (w *worker) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.commands)
}

func (w *worker) start() {
	if w.started {
		return
	}

	w.rwMu.Lock()

	w.eventPipe = make(chan Event, w.cLimit)
	w.started = true

	w.rwMu.Unlock()

	go func() {
		var wg sync.WaitGroup

		for {
			select {
			case <-w.ctx.Done():
				w.rwMu.Lock()

				wg.Wait()

				w.started = false
				close(w.eventPipe)

				w.rwMu.Unlock()

				return
			case event := <-w.eventPipe:
				wg.Add(1)
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithCancel(w.ctx)

					defer cancel()
					w.eventSink(w, w.commands.Handle(ctx, event))
				}()
			}
		}
	}()
}

func (w *worker) sealed() {}
