package dispatcher

import (
	"context"
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
