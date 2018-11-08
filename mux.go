package dispatcher

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"
)

// NewMux creates new dispatch mux
func NewMux() *Mux {
	return &Mux{}
}

// Mux is the dispatch multiplexer
type Mux struct {
	handler map[string]Handler

	Logger *log.Logger
}

func (m *Mux) logf(format string, v ...interface{}) {
	if m.Logger != nil {
		m.Logger.Printf(format, v...)
	}
}

// Register registers handlers, and override old handler if exists
func (m *Mux) Register(hs ...Handler) {
	if m.handler == nil {
		m.handler = make(map[string]Handler)
	}

	for _, h := range hs {
		if !isHandler(h) {
			panic("dispatcher: h is not a handler")
		}

		k := msgNameFromHandler(h)
		m.handler[k] = h
		m.logf("dispatcher: register %s", k)
	}
}

// Handler returns handler for given message
func (m *Mux) Handler(msg Message) Handler {
	return m.handler[msgName(msg)]
}

func (m *Mux) dispatch(ctx context.Context, msg Message) error {
	k := msgName(msg)
	if k == "" {
		return fmt.Errorf("dispatcher: invalid message type '%s'", reflect.TypeOf(msg))
	}

	m.logf("dispatcher: dispatching %s", k)

	h := m.handler[k]
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
func (m *Mux) Dispatch(ctx context.Context, msgs ...Message) error {
	var err error
	for _, msg := range msgs {
		switch p := msg.(type) {
		case []Message:
			err = m.Dispatch(ctx, p...)
		case Messages:
			err = m.Dispatch(ctx, p...)
		default:
			err = m.dispatch(ctx, msg)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// DispatchAfter calls dispatch after given duration
// or run immediate if duration is negative,
// then call resultFn with return error
func (m *Mux) DispatchAfter(ctx context.Context, duration time.Duration, resultFn func(err error), msgs ...Message) {
	if resultFn == nil {
		resultFn = func(_ error) {}
	}

	go func() {
		select {
		case <-time.After(duration):
			resultFn(m.Dispatch(ctx, msgs...))
		case <-ctx.Done():
			resultFn(ctx.Err())
		}
	}()
}

// DispatchAt calls dispatch at given time,
// and will run immediate if time already passed
func (m *Mux) DispatchAt(ctx context.Context, t time.Time, resultFn func(err error), msgs ...Message) {
	m.DispatchAfter(ctx, time.Until(t), resultFn, msgs...)
}
