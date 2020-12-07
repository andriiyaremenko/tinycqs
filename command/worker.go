package command

import "context"

func NewWorker(ctx context.Context, commands Commands) CommandsWorker {
	return &worker{ctx, commands}
}

type worker struct {
	ctx context.Context
	c   Commands
}

func (w *worker) HandleOnly(event Event, only ...string) Event {
	return w.c.HandleOnly(w.ctx, event, only...)
}

func (w *worker) Handle(event Event) Event {
	return w.c.Handle(w.ctx, event)
}

func (w *worker) sealed() {}
