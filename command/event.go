package command

type Ev struct {
	Type string
	P    []byte
}

func (e Ev) EventType() string {
	return e.Type
}

func (e Ev) Payload() []byte {
	return e.P
}
