package dispatcher_test

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/moonrhythm/dispatcher"
)

type httpReq1 struct {
	Name string `json:"name"`

	Result struct {
		ID string `json:"id"`
	} `json:"-"`
}

func TestHTTPHandler(t *testing.T) {
	var errNotATester = errors.New("not a tester")

	d := NewMux()
	d.Register(func(_ context.Context, m *httpReq1) error {
		if m.Name != "tester" {
			return errNotATester
		}
		m.Result.ID = "1"
		return nil
	})

	hd := HTTPHandler{
		Dispatcher: d,
	}
	hd.Register("/req1", (*httpReq1)(nil))

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req1", bytes.NewBufferString(`{"name":"tester"}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 200 {
			t.Errorf("expected http response with 200; got %v", code)
		}
		if status := w.Result().Header.Get("X-Dispatch-Status"); status != "1" {
			t.Errorf("expected dispatch status 1; got %v", status)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `{"id":"1"}` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})

	t.Run("CustomError", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req1", bytes.NewBufferString(`{"name":"admin"}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 200 {
			t.Errorf("expected http response with 200; got %v", code)
		}
		if status := w.Result().Header.Get("X-Dispatch-Status"); status != "0" {
			t.Errorf("expected dispatch status 0; got %v", status)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `{"type":"*errors.errorString","value":{},"error":"not a tester"}` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req_not_found", bytes.NewBufferString(`{}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 200 {
			t.Errorf("expected http response with 200; got %v", code)
		}
		if status := w.Result().Header.Get("X-Dispatch-Status"); status != "0" {
			t.Errorf("expected dispatch status 0; got %v", status)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `{"type":"*errors.errorString","value":{},"error":"dispatcher: handler not found"}` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})
}
