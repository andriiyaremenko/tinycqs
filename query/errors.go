package query

import (
	"fmt"
)

type ErrQueryHandlerNotFound struct {
	queryName string
}

func (err *ErrQueryHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for query %s", err.queryName)
}

func NewErrQueryHandlerNotFound(queryName string) *ErrQueryHandlerNotFound {
	return &ErrQueryHandlerNotFound{queryName}
}
