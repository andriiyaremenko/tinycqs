package query

import (
	"fmt"
)

// error type returned if incorrect Handler was passed to Queries.
type ErrIncorrectHandler struct {
	handler Handler
}

// Implementation of error.
func (err *ErrIncorrectHandler) Error() string {
	return fmt.Sprintf("query handler %#v has incorrect format", err.handler)
}

// error type returned if no Handlers was found for particular query.
type ErrQueryHandlerNotFound struct {
	queryName string
}

// Implementation of error.
func (err *ErrQueryHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for query %s", err.queryName)
}

// returns *ErrQueryHandlerNotFound of queryName.
func NewErrQueryHandlerNotFound(queryName string) *ErrQueryHandlerNotFound {
	return &ErrQueryHandlerNotFound{queryName}
}
