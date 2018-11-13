package dispatcher

import (
	"context"
	"time"
)

// expose global vars
var (
	DefaultDispatcher = &Dispatcher{}
)

// Register registers handlers into default dispatcher
func Register(hs ...Handler) {
	DefaultDispatcher.Register(hs...)
}

// Dispatch dispatchs default dispatcher
func Dispatch(ctx context.Context, msg ...Message) error {
	return DefaultDispatcher.Dispatch(ctx, msg...)
}

// DispatchAfter dispatchs default dispatcher
func DispatchAfter(ctx context.Context, duration time.Duration, resultFn func(err error), msg ...Message) {
	DefaultDispatcher.DispatchAfter(ctx, duration, resultFn, msg...)
}

// DispatchAt dispatchs default dispatcher
func DispatchAt(ctx context.Context, t time.Time, resultFn func(err error), msg ...Message) {
	DefaultDispatcher.DispatchAt(ctx, t, resultFn, msg...)
}
