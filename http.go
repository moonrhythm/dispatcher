package dispatcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

// HTTPDecoder is the function to decode value from http request
type HTTPDecoder func(r *http.Request, v interface{}) error

// HTTPEncoder is the function to encode value into http response writer
type HTTPEncoder func(w http.ResponseWriter, r *http.Request, v interface{}) error

// HTTPErrorEncoder is the function to encode error into http response writer
type HTTPErrorEncoder func(w http.ResponseWriter, r *http.Request, err error)

// JSONHTTPDecoder creates new json http decoder
func JSONHTTPDecoder() HTTPDecoder {
	return func(r *http.Request, v interface{}) error {
		return json.NewDecoder(r.Body).Decode(v)
	}
}

// JSONHTTPEncoder creates new json http encoder
func JSONHTTPEncoder() HTTPEncoder {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(v)
	}
}

// HTTPHandler wraps dispatcher into http handler
type HTTPHandler struct {
	Dispatcher   *Dispatcher
	Decoder      HTTPDecoder
	Encoder      HTTPEncoder
	ErrorEncoder HTTPErrorEncoder
	ResultField  string

	once sync.Once
	m    map[string]reflect.Type // path => struct type
}

func (h *HTTPHandler) init() {
	h.once.Do(func() {
		if h.m == nil {
			h.m = make(map[string]reflect.Type)
		}
		if h.Dispatcher == nil {
			h.Dispatcher = DefaultDispatcher
		}
		if h.Decoder == nil {
			h.Decoder = JSONHTTPDecoder()
		}
		if h.Encoder == nil {
			h.Encoder = JSONHTTPEncoder()
		}
		if h.ErrorEncoder == nil {
			h.ErrorEncoder = func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		if h.ResultField == "" {
			h.ResultField = "Result"
		}
	})
}

// Register registers path with struct
func (h *HTTPHandler) Register(path string, msg Message) {
	h.init()

	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Ptr {
		panic("dispatcher/httprpc: msg is not a Message")
	}
	if _, ok := t.Elem().FieldByName(h.ResultField); !ok {
		panic(fmt.Sprintf("dispatcher/httprpc: msg don't have '%s' field", h.ResultField))
	}
	h.m[path] = t.Elem()
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.init()

	t := h.m[r.URL.Path]
	if t == nil {
		h.ErrorEncoder(w, r, ErrNotFound)
		return
	}

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

	resp := refReq.Elem().FieldByName(h.ResultField)
	err = h.Encoder(w, r, resp.Interface())
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}
}
