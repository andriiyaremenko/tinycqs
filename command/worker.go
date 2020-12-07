package command

import "context"

func NewWorker(ctx context.Context, eventSink func(Event), commands Commands) CommandsWorker {
	w := &worker{
		ctx:       ctx,
		eventSink: eventSink,
		commands:  commands}

	w.start()

	return w
}

type worker struct {
	ctx       context.Context
	commands  Commands
	eventPipe chan Event
	eventSink func(event Event)
}

func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				w.eventSink(&ErrDone{Cause: w.ctx.Err()})
				close(w.eventPipe)

				return
			case event := <-w.eventPipe:
				w.eventSink(w.commands.Handle(w.ctx, event))
			}
		}
	}()
}

func (w *worker) Handle(event Event) {
	w.eventPipe <- event
}

func (w *worker) sealed() {}
