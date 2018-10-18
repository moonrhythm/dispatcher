package dispatcher_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/moonrhythm/dispatcher"
)

type msg1 struct {
	Name string
}

type msg2 struct {
	Data int
}

func TestDispatchSuccess(t *testing.T) {
	d := New()

	called := false
	d.Register(func(ctx context.Context, m *msg1) error {
		if ctx == nil {
			t.Errorf("expected ctx not nil")
		}

		called = true
		if m.Name != "test1" {
			t.Errorf("expected msg.Name to be 'test1'; got '%s'", m.Name)
		}
		return nil
	})
	d.Register(func(ctx context.Context, m *msg2) error {
		t.Error("expected handler for msg2 was not called")
		return nil
	})

	d.Dispatch(context.Background(), &msg1{Name: "test1"})

	if !called {
		t.Errorf("expected handler was called")
	}
}

func TestDispatchNotFound(t *testing.T) {
	d := New()

	if d.Dispatch(context.Background(), &msg1{}) != ErrNotFound {
		t.Error("expected returns handler not found error")
	}
}

func TestRegisterNotHandler(t *testing.T) {
	d := New()

	testCases := []struct {
		desc string
		h    interface{}
	}{
		{"string", "some string"},
		{"empty function", func() {}},
		{"invalid 1st input", func(int, int) {}},
		{"invalid 2rd input", func(context.Context, msg1) {}},
		{"empty output", func(context.Context, *msg1) {}},
		{"invalid output", func(context.Context, *msg1) *msg1 { return nil }},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic")
				}
			}()

			d.Register(tC.h)
		})
	}
}

func TestDispatchReturnError(t *testing.T) {
	var e = errors.New("err!")

	d := New()

	d.Register(func(ctx context.Context, m *msg1) error {
		return e
	})

	if d.Dispatch(context.Background(), &msg1{}) != e {
		t.Errorf("expected dispatch return error")
	}
}
