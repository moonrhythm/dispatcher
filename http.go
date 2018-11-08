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

type httpError struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	Error string      `json:"error"`
}

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
		w.Header().Set("X-Dispatch-Status", "1")
		return json.NewEncoder(w).Encode(v)
	}
}

// JSONHTTPErrorEncoder creates new json http error encoder
func JSONHTTPErrorEncoder() HTTPErrorEncoder {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Dispatch-Status", "0")

		typeRef := reflect.TypeOf(err)
		typeName := rtName(typeRef)
		if typeName == "" {
			typeName = typeRef.String()
		}

		json.NewEncoder(w).Encode(&httpError{
			typeName,
			err,
			err.Error(),
		})
	}
}

// HTTPHandler wraps dispatcher into http handler
type HTTPHandler struct {
	Dispatcher   Dispatcher
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
			h.Dispatcher = DefaultMux
		}
		if h.Decoder == nil {
			h.Decoder = JSONHTTPDecoder()
		}
		if h.Encoder == nil {
			h.Encoder = JSONHTTPEncoder()
		}
		if h.ErrorEncoder == nil {
			h.ErrorEncoder = JSONHTTPErrorEncoder()
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

	refM := reflect.New(t)
	m := refM.Interface()
	err := h.Decoder(r, m)
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}

	err = h.Dispatcher.Dispatch(r.Context(), m)
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}

	refResult := refM.Elem().FieldByName(h.ResultField)
	err = h.Encoder(w, r, refResult.Interface())
	if err != nil {
		h.ErrorEncoder(w, r, err)
		return
	}
}
