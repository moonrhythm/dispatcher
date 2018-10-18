package dispatcher_test

import (
	"context"
	"testing"

	. "github.com/moonrhythm/dispatcher"
)

func TestGlobalDispatchSuccess(t *testing.T) {
	called := false
	Register(func(ctx context.Context, m *msg1) error {
		if ctx == nil {
			t.Errorf("expected ctx not nil")
		}

		called = true
		if m.Name != "test1" {
			t.Errorf("expected msg.Name to be 'test1'; got '%s'", m.Name)
		}
		return nil
	})
	Register(func(ctx context.Context, m *msg2) error {
		t.Error("expected handler for msg2 was not called")
		return nil
	})

	Dispatch(context.Background(), &msg1{Name: "test1"})

	if !called {
		t.Errorf("expected handler was called")
	}
}
