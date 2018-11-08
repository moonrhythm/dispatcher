package dispatcher

import (
	"fmt"
	"net/http"
	"reflect"
)

// HTTPDecoder is the function to decode value from http request
type HTTPDecoder func(r *http.Request, v interface{}) error

// HTTPEncoder is the function to encode value into http response writer
type HTTPEncoder func(w http.ResponseWriter, r *http.Request, v interface{}) error

// HTTPErrorEncoder is the function to encode error into http response writer
type HTTPErrorEncoder func(w http.ResponseWriter, r *http.Request, err error)

// HTTPRPC wraps dispatcher into http handler using rpc-style
type HTTPRPC struct {
	Dispatcher   *Dispatcher
	Decoder      HTTPDecoder
	Encoder      HTTPEncoder
	ErrorEncoder HTTPErrorEncoder
	m            map[string]reflect.Type // path => struct type
}

const httpResult = "Result"

// Register registers path with struct
func (h *HTTPRPC) Register(path string, msg Message) {
	if h.m == nil {
		h.m = make(map[string]reflect.Type)
	}

	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Ptr {
		panic("dispatcher/httprpc: msg is not a Message")
	}
	if _, ok := t.Elem().FieldByName(httpResult); !ok {
		panic(fmt.Sprintf("dispatcher/httprpc: msg don't have '%s' field", httpResult))
	}
	h.m[path] = t.Elem()
}

func (h *HTTPRPC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	t := h.m[path]
	if t == nil {
		h.ErrorEncoder(w, r, ErrNotFound)
		return
	}

	// create request body struct
	refReq := reflect.New(t)
	req := refReq.Interface()
	err := h.Decoder(r, req)
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}

	err = h.Dispatcher.Dispatch(r.Context(), req)
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}

	resp := refReq.Elem().FieldByName(httpResult)
	err = h.Encoder(w, r, resp.Interface())
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}
}
