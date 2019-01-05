package dispatcher

import (
	"fmt"
	"net/http"
	"reflect"
)

// HTTPHandlerWrapper wraps dispatcher message to http handler
type HTTPHandlerWrapper struct {
	Dispatcher Dispatcher
	Decoder    func(r *http.Request, v interface{}) error
	Encoder    func(w http.ResponseWriter, r *http.Request, v interface{}) error
	Result     string
}

// Handler wraps message to http handler
func (wrapper HTTPHandlerWrapper) Handler(msg Message) http.Handler {
	if wrapper.Dispatcher == nil {
		panic("dispatcher: nil dispatcher")
	}
	if wrapper.Decoder == nil {
		panic("dispatcher: nil decoder")
	}
	if wrapper.Encoder == nil {
		panic("dispatcher: nil encoder")
	}
	if wrapper.Result == "" {
		panic("dispatcher: empty result")
	}

	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Ptr {
		panic("dispatcher: msg is not a Message")
	}
	if _, ok := t.Elem().FieldByName(wrapper.Result); !ok {
		panic(fmt.Sprintf("dispatcher: msg don't have '%s' field", wrapper.Result))
	}

	t = t.Elem()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refM := reflect.New(t)
		m := refM.Interface()
		err := wrapper.Decoder(r, m)
		if err != nil {
			wrapper.Encoder(w, r, err)
			return
		}

		err = wrapper.Dispatcher.Dispatch(r.Context(), m)
		if err != nil {
			wrapper.Encoder(w, r, err)
			return
		}

		refResult := refM.Elem().FieldByName(wrapper.Result)
		err = wrapper.Encoder(w, r, refResult.Interface())
		if err != nil {
			wrapper.Encoder(w, r, err)
			return
		}
	})
}
