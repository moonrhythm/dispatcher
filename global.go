package dispatcher

import (
	"context"
	"time"
)

// expose global vars
var (
	DefaultMux = NewMux()
)

// Register registers handlers into default dispatcher
func Register(hs ...Handler) {
	DefaultMux.Register(hs...)
}

// Dispatch dispatchs default dispatcher
func Dispatch(ctx context.Context, msg ...Message) error {
	return DefaultMux.Dispatch(ctx, msg...)
}

// DispatchAfter dispatchs default dispatcher
func DispatchAfter(ctx context.Context, duration time.Duration, resultFn func(err error), msg ...Message) {
	DefaultMux.DispatchAfter(ctx, duration, resultFn, msg...)
}

// DispatchAt dispatchs default dispatcher
func DispatchAt(ctx context.Context, t time.Time, resultFn func(err error), msg ...Message) {
	DefaultMux.DispatchAt(ctx, t, resultFn, msg...)
}
