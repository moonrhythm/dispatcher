package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
)

// New creates new dispatcher
func New() *Dispatcher {
	return &Dispatcher{}
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

// Dispatcher is the event dispatcher
type Dispatcher struct {
	handler map[string]Handler

	Logger *log.Logger
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

func (d *Dispatcher) logf(format string, v ...interface{}) {
	if d.Logger != nil {
		d.Logger.Printf(format, v...)
	}
}

// Register registers handlers, and override old handler if exists
func (d *Dispatcher) Register(hs ...Handler) {
	if d.handler == nil {
		d.handler = make(map[string]Handler)
	}

	for _, h := range hs {
		if !isHandler(h) {
			panic("dispatcher: h is not a handler")
		}

		k := msgNameFromHandler(h)
		d.handler[k] = h
		d.logf("dispatcher: register %s", k)
	}
}

// Handler returns handler for given message
func (d *Dispatcher) Handler(msg Message) Handler {
	return d.handler[msgName(msg)]
}

func (d *Dispatcher) dispatch(ctx context.Context, msg Message) error {
	k := msgName(msg)
	if k == "" {
		return fmt.Errorf("dispatcher: invalid message type '%s'", reflect.TypeOf(msg))
	}

	d.logf("dispatcher: dispatching %s", k)

	h := d.handler[k]
	if h == nil {
		return ErrNotFound
	}

	err := reflect.ValueOf(h).Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(msg),
	})[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

// Dispatch calls handler for given messages in sequence order,
// when a handler returns error, dispatch will stop and return that error
func (d *Dispatcher) Dispatch(ctx context.Context, msg ...Message) error {
	for _, m := range msg {
		err := d.dispatch(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}
