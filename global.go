package dispatcher

import (
	"context"
)

// expose global vars
var (
	DefaultDispatcher = &Dispatcher{}
)

// Register registers a handler into default dispatcher
func Register(h Handler) {
	DefaultDispatcher.Register(h)
}

// Dispatch dispatchs default dispatcher
func Dispatch(ctx context.Context, msg Message) error {
	return DefaultDispatcher.Dispatch(ctx, msg)
}
