package tracing

// M implements Metadata.
type M struct {
	EID            string
	ECorrelationID string
	ECausationID   string
}

func (m M) New(id string) Metadata {
	return M{EID: id, ECorrelationID: m.ECorrelationID, ECausationID: m.EID}
}

func (m M) ID() string {
	return m.EID
}

func (m M) CorrelationID() string {
	return m.ECorrelationID
}

func (m M) CausationID() string {
	return m.ECausationID
}
