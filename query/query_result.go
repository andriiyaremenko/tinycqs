package query

import "encoding/json"

// Q implements QueryResult
type Q struct {
	Name  string
	B     []byte
	Error error
}

func (q Q) QueryName() string {
	return q.Name
}

func (q Q) Body() []byte {
	return q.B
}

func (q Q) Err() error {
	return q.Error
}

func (q Q) UnmarshalJSONBody(v interface{}) error {
	if err := q.Err(); err != nil {
		return err
	}

	return json.Unmarshal(q.Body(), v)
}
