package command

const Done doneEv = "DONE"

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

type Ev struct {
	EType string
	P     []byte
}

func (e Ev) EventType() string {
	return e.EType
}

func (e Ev) Payload() []byte {
	return e.P
}

func (e Ev) Err() error {
	return nil
}
