package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
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
}

func reflectTypeName(r reflect.Type) string {
	pkg := r.PkgPath()
	name := r.Name()
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

// MessageNameFromHandler gets message name from handler
func MessageNameFromHandler(h Handler) string {
	if !isHandler(h) {
		return ""
	}
	return reflectTypeName(reflect.TypeOf(h).In(1).Elem())
}

// MessageName gets message name
func MessageName(msg Message) string {
	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Ptr {
		return ""
	}
	return reflectTypeName(t.Elem())
}

func isHandler(h Handler) bool {
	t := reflect.TypeOf(h)

	if t.Kind() != reflect.Func {
		return false
	}

	if t.NumIn() != 2 {
		return false
	}
	if t.In(0).Kind() != reflect.Interface && reflectTypeName(t.In(0)) != "context.Context" {
		return false
	}
	if t.In(1).Kind() != reflect.Ptr {
		return false
	}

	if t.NumOut() != 1 {
		return false
	}
	if reflectTypeName(t.Out(0)) != "error" {
		return false
	}

	return true
}

// Register registers handlers, and override old handler if exists
func (d *Dispatcher) Register(hs ...Handler) {
	if d.handler == nil {
		d.handler = make(map[string]Handler)
	}

	for _, h := range hs {
		k := MessageNameFromHandler(h)
		if k == "" {
			panic("dispatcher: h is not a handler")
		}
		d.handler[k] = h
	}
}

// Handler returns handler for given message
func (d *Dispatcher) Handler(msg Message) Handler {
	return d.handler[MessageName(msg)]
}

func (d *Dispatcher) dispatch(ctx context.Context, msg Message) error {
	k := MessageName(msg)
	if k == "" {
		return fmt.Errorf("dispatcher: invalid message type '%s'", reflect.TypeOf(msg))
	}

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

// DispatchAfter calls dispatch after given duration
// or run immediate if duration is negative,
// then call resultFn with return error
func (d *Dispatcher) DispatchAfter(ctx context.Context, duration time.Duration, resultFn func(err error), msg ...Message) {
	if resultFn == nil {
		resultFn = func(_ error) {}
	}

	go func() {
		select {
		case <-time.After(duration):
			resultFn(d.Dispatch(ctx, msg...))
		case <-ctx.Done():
			resultFn(ctx.Err())
		}
	}()
}

// DispatchAt calls dispatch at given time,
// and will run immediate if time already passed
func (d *Dispatcher) DispatchAt(ctx context.Context, t time.Time, resultFn func(err error), msg ...Message) {
	d.DispatchAfter(ctx, time.Until(t), resultFn, msg...)
}
