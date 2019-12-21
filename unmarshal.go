package jcrypt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type dstValue struct {
	Type        reflect.Type
	Value       reflect.Value
	StructField *reflect.StructField
}

func (dst dstValue) elem() dstValue {
	return dstValue{dst.Type.Elem(), dst.Value.Elem(), dst.StructField}
}

func (dst dstValue) Kind() reflect.Kind {
	t := dst.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind()
}

func (dst dstValue) Assign(val interface{}) error {
	v := dst.Value
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v.Set(reflect.ValueOf(val))
	return nil
}

type unmarshalHandler func(src interface{}, dst dstValue) (handled bool, err error)

func jsonUnmarshal(data []byte, dst interface{}, f unmarshalHandler) error {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return fmt.Errorf("malformed json: %s", err.Error())
	}

	t := reflect.TypeOf(dst)
	v := reflect.ValueOf(dst)
	return jsonUnmarshalValue(raw, dstValue{t, v, nil}, f)
}

func jsonUnmarshalValue(src interface{}, dst dstValue, f unmarshalHandler) error {
	switch dst.Type.Kind() {
	case reflect.Ptr:
		return jsonUnmarshalValue(src, dst.elem(), f)

	case reflect.Map:
		return fmt.Errorf("maps not yet supported")
	case reflect.Struct:
		return jsonUnmarshalStruct(src, dst, f)

	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return fmt.Errorf("arrays and slices not yet supported")

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Bool:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.String:
		if dst.Value.CanSet() {
			dst.Value.Set(reflect.ValueOf(src))
			return nil
		}
		return fmt.Errorf("cannot assign to %s", dst.Type.Name())

	default:
		return fmt.Errorf("data type %s not yet supported", dst.Type.Kind())
	}
}

func jsonUnmarshalStruct(src interface{}, dst dstValue, f unmarshalHandler) error {
	dstFields := make(map[string]dstValue)

	fieldCount := dst.Type.NumField()
	for i := 0; i < fieldCount; i++ {
		field := dst.Type.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")
		fieldName := field.Name
		if len(jsonTag[0]) > 0 {
			if jsonTag[0] == "-" {
				continue
			} else {
				fieldName = jsonTag[0]
			}
		}

		//TODO omitempty and string options in jsonTag[1]
		dstFields[fieldName] = dstValue{field.Type, dst.Value.Field(i), &field}
	}

	srcFields, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot assign %T to %s", src, dst.Type.Name())
	}

	for srcName, srcValue := range srcFields {
		if dstField, ok := dstFields[srcName]; ok {
			if f != nil {
				handled, err := f(srcValue, dstField)
				if err != nil {
					return err
				}

				if handled {
					continue
				}
			}

			if err := jsonUnmarshalValue(srcValue, dstField, f); err != nil {
				return err
			}
		}
	}

	return nil
}
