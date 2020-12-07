package query

import (
	"fmt"
)

type ErrIncorrectHandler struct {
	handler Handler
}

func (err *ErrIncorrectHandler) Error() string {
	return fmt.Sprintf("query handler %#v has incorrect format", err.handler)
}

type ErrQueryHandlerNotFound struct {
	queryName string
}

func (err *ErrQueryHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for query %s", err.queryName)
}

func NewErrQueryHandlerNotFound(queryName string) *ErrQueryHandlerNotFound {
	return &ErrQueryHandlerNotFound{queryName}
}
