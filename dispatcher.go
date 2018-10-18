package dispatcher

import (
	"context"
	"errors"
	"reflect"
)

// New creates new dispatcher
func New() *Dispatcher {
	return &Dispatcher{}
}

// Errors
var (
	ErrHandlerNotFound = errors.New("dispatcher: handler not found")
)

// Handler is the event handler
//
// func(context.Context, *Any) error
type Handler interface{}

// Message is the event message
type Message interface{}

// Dispatcher is the event dispatcher
type Dispatcher struct {
	handler map[string]Handler
}

func rtName(r reflect.Type) string {
	pkg := r.PkgPath()
	name := r.Name()
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func nameFromHandler(h Handler) string {
	return reflect.TypeOf(h).In(1).Elem().Name()
}

func name(msg Message) string {
	return reflect.TypeOf(msg).Elem().Name()
}

func isHandler(h Handler) bool {
	t := reflect.TypeOf(h)

	if t.Kind() != reflect.Func {
		return false
	}

	if t.NumIn() != 2 {
		return false
	}
	if rtName(t.In(0)) != "context.Context" {
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

// Set sets handler, and override old handler if exists
func (d *Dispatcher) Set(h Handler) {
	if !isHandler(h) {
		panic("dispatcher: h is not a handler")
	}

	if d.handler == nil {
		d.handler = make(map[string]Handler)
	}

	d.handler[nameFromHandler(h)] = h
}

// Handler returns handler for given message
func (d *Dispatcher) Handler(msg Message) Handler {
	return d.handler[name(msg)]
}

// Dispatch calls handler for given event message
func (d *Dispatcher) Dispatch(ctx context.Context, msg Message) error {
	h := d.Handler(msg)

	if h == nil {
		return ErrHandlerNotFound
	}

	err := reflect.ValueOf(h).Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(msg),
	})[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}
