package dispatcher_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
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

	d := New()
	d.Register(func(_ context.Context, m *httpReq1) error {
		if m.Name != "tester" {
			return errNotATester
		}
		m.Result.ID = "1"
		return nil
	})

	hd := HTTPHandler{
		Dispatcher: d,
		ErrorEncoder: func(w http.ResponseWriter, r *http.Request, err error) {
			if err == errNotATester {
				http.Error(w, err.Error(), 400)
				return
			}
			if err == ErrNotFound {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, err.Error(), 500)
		},
	}
	hd.Register("/req1", (*httpReq1)(nil))

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req1", bytes.NewBufferString(`{"name":"tester"}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 200 {
			t.Errorf("expected http response with 200; got %v", code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `{"id":"1"}` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})

	t.Run("CustomError", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req1", bytes.NewBufferString(`{"name":"admin"}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 400 {
			t.Errorf("expected http response with 400; got %v", code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `not a tester` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/req_not_found", bytes.NewBufferString(`{}`))
		hd.ServeHTTP(w, r)

		if code := w.Result().StatusCode; code != 404 {
			t.Errorf("expected http response with 404; got %v", code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != `not found` {
			t.Errorf("invalid http response body; got %s", body)
		}
	})
}
