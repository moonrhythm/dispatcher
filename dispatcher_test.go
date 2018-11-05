package dispatcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestDispatchMulti(t *testing.T) {
	d := New()

	called := 0
	var errStop = errors.New("some error")

	d.Register(func(ctx context.Context, m *msg1) error {
		called++
		if called == 1 && m.Name != "test1" {
			t.Errorf("expected msg.Name to be 'test1'; got '%s'", m.Name)
		}
		if called == 2 && m.Name != "test2" {
			t.Errorf("expected msg.Name to be 'test2'; got '%s'", m.Name)
		}
		if called == 3 {
			return errStop
		}
		return nil
	})

	err := d.Dispatch(context.Background(),
		&msg1{Name: "test1"},
		&msg1{Name: "test2"},
		&msg1{Name: "test3"},
		&msg1{Name: "test4"},
	)

	if called != 3 {
		t.Errorf("expected handler was called 3 times")
	}
	if err != errStop {
		t.Errorf("expected error return")
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

func TestDispatchInvalidMessage(t *testing.T) {
	d := New()
	err := d.Dispatch(context.Background(), msg1{})
	if err == nil {
		t.Errorf("expected return error when dispatch struct")
	}
}

func TestDispatchAfter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		d := New()
		called := false
		resultCalled := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		d.DispatchAfter(context.Background(), 10*time.Millisecond,
			func(err error) {
				resultCalled = true
				if err != nil {
					t.Errorf("expected no error")
				}
			},
			&msg1{},
		)

		time.Sleep(40 * time.Millisecond)
		if !called {
			t.Errorf("expected handler was called")
		}
		if !resultCalled {
			t.Errorf("expected resultFn was called")
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		d := New()
		retErr := errors.New("some error")
		d.Register(func(ctx context.Context, m *msg1) error {
			return retErr
		})

		d.DispatchAfter(context.Background(), 10*time.Millisecond,
			func(err error) {
				if err != retErr {
					t.Errorf("expected error")
				}
			},
			&msg1{},
		)

		time.Sleep(40 * time.Millisecond)
	})

	t.Run("Cancel", func(t *testing.T) {
		t.Parallel()

		d := New()
		d.Register(func(ctx context.Context, m *msg1) error {
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())
		d.DispatchAfter(ctx, 10*time.Millisecond,
			func(err error) {
				if err != context.Canceled {
					t.Errorf("expected context canceled error")
				}
			},
			&msg1{},
		)
		cancel()

		time.Sleep(40 * time.Millisecond)
	})

	t.Run("No Result", func(t *testing.T) {
		t.Parallel()

		d := New()
		d.Register(func(ctx context.Context, m *msg1) error {
			return nil
		})

		d.DispatchAfter(context.Background(), 10*time.Millisecond, nil, &msg1{})
		time.Sleep(40 * time.Millisecond)
	})
}
