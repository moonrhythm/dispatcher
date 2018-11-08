package dispatcher

import (
	"context"
	"errors"
	"reflect"
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

func rtName(r reflect.Type) string {
	pkg := r.PkgPath()
	name := r.Name()
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func msgNameFromHandler(h Handler) string {
	return rtName(reflect.TypeOf(h).In(1).Elem())
}

func msgName(msg Message) string {
	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Ptr {
		return ""
	}
	return rtName(t.Elem())
}

func isHandler(h Handler) bool {
	t := reflect.TypeOf(h)

	if t.Kind() != reflect.Func {
		return false
	}

	if t.NumIn() != 2 {
		return false
	}
	if t.In(0).Kind() != reflect.Interface && rtName(t.In(0)) != "context.Context" {
		return false
	}
	if t.In(1).Kind() != reflect.Ptr {
		return false
	}

	if t.NumOut() != 1 {
		return false
	}
	if rtName(t.Out(0)) != "error" {
		return false
	}

	return true
}
