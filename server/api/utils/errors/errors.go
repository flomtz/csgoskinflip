package errors

import "runtime/debug"

type Error struct {
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace"`
}

func NewError(message string) *Error {
	newError := Error{
		Message:    message,
		StackTrace: string(debug.Stack()),
	}

	return &newError
}
