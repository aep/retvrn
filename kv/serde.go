package kv

import (
	"bytes"
	"fmt"
	"reflect"
)

func serialize(v interface{}) ([]byte, error) {

	var b bytes.Buffer

	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.String:
		b.Write([]byte("0s"))
		b.WriteString(v.(string))
	default:
		return nil, fmt.Errorf("unsupported type %s", t)
	}

	return b.Bytes(), nil
}

func deserialize(b []byte, v interface{}) error {

	// nil value or error?
	if len(b) < 2 {
		return nil
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("unsupported type %s . should be pointer", t)
	}

	t = t.Elem()
	switch t.Kind() {
	case reflect.String:
		switch string(b[0:2]) {
		case "0s":
			reflect.ValueOf(v).Elem().SetString(string(b[2:]))
		default:
			return fmt.Errorf("cannot read value of type '%s' as %s", string(b[0:2]), t)
		}
	default:
		return fmt.Errorf("unsupported type %s", t)
	}

	return nil
}
