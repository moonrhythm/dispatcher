package dispatcher

import (
	"context"
	"errors"
)

// Dispatcher type
type Dispatcher interface {
	Dispatch(context.Context, ...Message) error
}

// Errors
var (
	ErrNotFound = errors.New("dispatcher: handler not found")
)

// Handler is the event handler
//
// func(context.Context, *Any) error
type Handler interface{}

// Message is the event message
type Message interface{}

// Messages is the message collection
type Messages []Message

// Push appends messages into collection
func (msgs *Messages) Push(ms ...Message) {
	*msgs = append(*msgs, ms...)
}

// Clear clears messages
func (msgs *Messages) Clear() {
	*msgs = nil
}
