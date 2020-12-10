package command

const (
	DoneEventType string = "DONE"

	Done        doneEv = doneEv(DoneEventType)
	doneWriting doneEv = doneEv("WORKER_DONE_WRITING")
)

type doneEv string

func (done doneEv) EventType() string {
	return string(done)
}

func (done doneEv) Payload() []byte {
	return nil
}

func (done doneEv) Err() error {
	return nil
}

type E struct {
	EType    string
	EPayload []byte
}

func (e E) EventType() string {
	return e.EType
}

func (e E) Payload() []byte {
	return e.EPayload
}

func (e E) Err() error {
	return nil
}
