package index

import (
	"bytes"
	"fmt"
	"reflect"
)

func Serialize(v interface{}) ([]byte, error) {

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

func Deserialize(b []byte, vi interface{}) (interface{}, error) {

	// nil value or error?
	if len(b) < 2 {
		return nil, nil
	}

	t := reflect.TypeOf(vi)
	v := reflect.ValueOf(vi)

	switch string(b[0:2]) {
	case "0s":
		if vi == nil {
			return string(b[2:]), nil
		}
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
			if v.IsNil() {
				v.Set(reflect.New(t))
			}
			v = v.Elem()
		}
		if t.Kind() == reflect.String {
			v.SetString(string(b[2:]))
			return vi, nil
		} else {
			return nil, fmt.Errorf("cannot read value of type '%s' as %s", string(b[0:2]), t)
		}

	default:
		return nil, fmt.Errorf("cannot read value of type '%s' as %s", string(b[0:2]), t)
	}
}
