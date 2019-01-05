package dispatcher_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/moonrhythm/dispatcher"
)

type httpEcho struct {
	Text   string
	Result struct {
		Text string `json:"text"`
	} `json:"-"`
}

type httpNoResult struct {
	Text string
}

func TestHTTPHandlerWrapper(t *testing.T) {
	d := NewMux()

	d.Register(func(ctx context.Context, m *httpEcho) error {
		m.Result.Text = m.Text
		return nil
	})

	encoder := func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err, ok := v.(error); ok {
			v = struct {
				Error string `json:"error"`
			}{err.Error()}
			w.WriteHeader(500)
		}
		return json.NewEncoder(w).Encode(v)
	}

	decoder := func(r *http.Request, v interface{}) error {
		return json.NewDecoder(r.Body).Decode(v)
	}

	const result = "Result"

	t.Run("Nil Dispatcher", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Encoder: encoder,
			Decoder: decoder,
			Result:  result,
		}
		assert.Panics(t, func() { wrapper.Handler(new(httpEcho)) })
	})

	t.Run("Nil Encoder", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Decoder:    decoder,
			Result:     result,
		}
		assert.Panics(t, func() { wrapper.Handler(new(httpEcho)) })
	})

	t.Run("Nil Decoder", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Result:     result,
		}
		assert.Panics(t, func() { wrapper.Handler(new(httpEcho)) })
	})

	t.Run("Empty Result", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder:    decoder,
		}
		assert.Panics(t, func() { wrapper.Handler(new(httpEcho)) })
	})

	t.Run("Invalid Message", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder:    decoder,
			Result:     result,
		}
		assert.Panics(t, func() { wrapper.Handler(TestHTTPHandlerWrapper) })
	})

	t.Run("Result Not Found", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder:    decoder,
			Result:     result,
		}
		assert.Panics(t, func() { wrapper.Handler(new(httpNoResult)) })
	})

	t.Run("Success", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder:    decoder,
			Result:     result,
		}
		h := wrapper.Handler(new(httpEcho))
		if assert.NotNil(t, h) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"text":"hello"}`))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assert.JSONEq(t, `{"text":"hello"}`, w.Body.String())
		}
	})

	t.Run("Decode Error", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder: func(r *http.Request, v interface{}) error {
				return fmt.Errorf("decode error")
			},
			Result: result,
		}
		h := wrapper.Handler(new(httpEcho))
		if assert.NotNil(t, h) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"text":"hello"}`))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assert.EqualValues(t, 500, w.Code)
			assert.JSONEq(t, `{"error":"decode error"}`, w.Body.String())
		}
	})

	t.Run("Dispatch Error", func(t *testing.T) {
		d := NewMux()

		d.Register(func(ctx context.Context, m *httpEcho) error {
			return fmt.Errorf("dispatch error")
		})

		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder:    encoder,
			Decoder:    decoder,
			Result:     result,
		}
		h := wrapper.Handler(new(httpEcho))
		if assert.NotNil(t, h) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"text":"hello"}`))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assert.EqualValues(t, 500, w.Code)
			assert.JSONEq(t, `{"error":"dispatch error"}`, w.Body.String())
		}
	})

	t.Run("Encode Error", func(t *testing.T) {
		wrapper := HTTPHandlerWrapper{
			Dispatcher: d,
			Encoder: func(w http.ResponseWriter, r *http.Request, v interface{}) error {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				if err, ok := v.(error); ok {
					v = struct {
						Error string `json:"error"`
					}{err.Error()}
					w.WriteHeader(500)
					return json.NewEncoder(w).Encode(v)
				}
				return fmt.Errorf("encode error")
			},
			Decoder: decoder,
			Result:  result,
		}
		h := wrapper.Handler(new(httpEcho))
		if assert.NotNil(t, h) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"text":"hello"}`))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assert.EqualValues(t, 500, w.Code)
			assert.JSONEq(t, `{"error":"encode error"}`, w.Body.String())
		}
	})
}
