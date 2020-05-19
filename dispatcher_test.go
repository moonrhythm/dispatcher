// +build !race

package dispatcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/moonrhythm/dispatcher"
)

type msg1 struct {
	Name string
}

type msg2 struct {
	Data int
}

func TestDispatchSuccess(t *testing.T) {
	d := NewMux()

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

	if d.Handler(new(msg1)) == nil {
		t.Errorf("expected get registered handler not nil")
	}

	d.Dispatch(context.Background(), &msg1{Name: "test1"})

	if !called {
		t.Errorf("expected handler was called")
	}
}

func TestDispatchMulti(t *testing.T) {
	d := NewMux()

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

	err := Dispatch(context.Background(), d,
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
	d := NewMux()

	err := d.Dispatch(context.Background(), &msg1{})
	if !errors.Is(err, ErrNotFound) {
		t.Error("expected returns handler not found error")
	}
	assert.Contains(t, err.Error(), "msg1")
}

func TestRegisterNotHandler(t *testing.T) {
	d := NewMux()

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

	d := NewMux()

	d.Register(func(ctx context.Context, m *msg1) error {
		return e
	})

	if d.Dispatch(context.Background(), &msg1{}) != e {
		t.Errorf("expected dispatch return error")
	}
}

func TestDispatchInvalidMessage(t *testing.T) {
	d := NewMux()
	err := d.Dispatch(context.Background(), msg1{})
	if err == nil {
		t.Errorf("expected return error when dispatch struct")
	}
}

func TestDispatchAfter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		d := NewMux()
		called := false
		resultCalled := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAfter(context.Background(), d, 10*time.Millisecond,
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

		d := NewMux()
		retErr := errors.New("some error")
		d.Register(func(ctx context.Context, m *msg1) error {
			return retErr
		})

		DispatchAfter(context.Background(), d, 10*time.Millisecond,
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

		d := NewMux()
		d.Register(func(ctx context.Context, m *msg1) error {
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())
		DispatchAfter(ctx, d, 10*time.Millisecond,
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

		d := NewMux()
		called := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAfter(context.Background(), d, 10*time.Millisecond, nil, &msg1{})
		time.Sleep(40 * time.Millisecond)
		if !called {
			t.Errorf("expected handler was called")
		}
	})

	t.Run("Zero duration", func(t *testing.T) {
		t.Parallel()

		d := NewMux()
		called := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAfter(context.Background(), d, 0, nil, &msg1{})
		time.Sleep(10 * time.Millisecond)
		if !called {
			t.Errorf("expected handler was called")
		}
	})

	t.Run("Negative duration", func(t *testing.T) {
		t.Parallel()

		d := NewMux()
		called := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAfter(context.Background(), d, -time.Hour, nil, &msg1{})
		time.Sleep(10 * time.Millisecond)
		if !called {
			t.Errorf("expected handler was called")
		}
	})
}

func TestDispatchAt(t *testing.T) {
	t.Run("Future", func(t *testing.T) {
		t.Parallel()

		d := NewMux()
		called := false
		resultCalled := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAt(context.Background(), d, time.Now().Add(10*time.Millisecond),
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

	t.Run("Past", func(t *testing.T) {
		t.Parallel()

		d := NewMux()
		called := false
		d.Register(func(ctx context.Context, m *msg1) error {
			called = true
			return nil
		})

		DispatchAt(context.Background(), d, time.Now().Add(-time.Hour), nil, &msg1{})
		time.Sleep(10 * time.Millisecond)
		if !called {
			t.Errorf("expected handler was called")
		}
	})
}

type benchMsg struct {
	A, B   int
	Result int
}

func BenchmarkDispatch(b *testing.B) {
	b.Run("Direct call function", func(b *testing.B) {
		f := func(_ context.Context, m *benchMsg) error {
			m.Result = m.A + m.B
			return nil
		}

		for i := 0; i < b.N; i++ {
			f(context.Background(), &benchMsg{A: 1, B: 2})
		}
	})

	b.Run("Direct call function from map", func(b *testing.B) {
		k := MessageName(new(benchMsg))

		m := map[string]func(context.Context, *benchMsg) error{
			k: func(_ context.Context, m *benchMsg) error {
				m.Result = m.A + m.B
				return nil
			},
		}

		for i := 0; i < b.N; i++ {
			m[k](context.Background(), &benchMsg{A: 1, B: 2})
		}
	})

	b.Run("Dispatch", func(b *testing.B) {
		d := NewMux()
		d.Register(func(_ context.Context, m *benchMsg) error {
			m.Result = m.A + m.B
			return nil
		})

		for i := 0; i < b.N; i++ {
			d.Dispatch(context.Background(), &benchMsg{A: 1, B: 2})
		}
	})
}
