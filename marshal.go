package jcrypt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type srcValue struct {
	Type        reflect.Type
	Value       reflect.Value
	StructField *reflect.StructField
}

func (src srcValue) elem() srcValue {
	return srcValue{src.Type.Elem(), src.Value.Elem(), src.StructField}
}

func (src srcValue) Kind() reflect.Kind {
	t := src.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind()
}

func (src srcValue) Interface() interface{} {
	v := src.Value
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Interface()
}

type marshalHandler func(src srcValue) (result interface{}, handled bool, err error)

func jsonMarshal(src interface{}, f marshalHandler) ([]byte, error) {
	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)
	dst, err := jsonMarshalValue(srcValue{t, v, nil}, f)
	if err != nil {
		return nil, err
	}

	return json.Marshal(dst)
}

func jsonMarshalValue(src srcValue, f marshalHandler) (interface{}, error) {
	switch src.Type.Kind() {
	case reflect.Ptr:
		return jsonMarshalValue(src.elem(), f)

	case reflect.Map:
		return nil, fmt.Errorf("maps not yet supported")
	case reflect.Struct:
		return jsonMarshalStruct(src, f)

	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return nil, fmt.Errorf("arrays and slices not yet supported")

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
		return src.Value.Interface(), nil

	default:
		return nil, fmt.Errorf("data type %s not yet supported", src.Type.Kind())
	}
}

func jsonMarshalStruct(src srcValue, f marshalHandler) (interface{}, error) {
	result := make(map[string]interface{})

	fieldCount := src.Type.NumField()
	for i := 0; i < fieldCount; i++ {
		field := src.Type.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")
		fieldName := field.Name
		if len(jsonTag[0]) > 0 {
			if jsonTag[0] == "-" {
				continue
			} else {
				fieldName = jsonTag[0]
			}
		}

		fieldValue := srcValue{field.Type, src.Value.Field(i), &field}

		//TODO omitempty and string options in jsonTag[1]

		value, err := func() (interface{}, error) {
			if f != nil {
				val, handled, err := f(fieldValue)
				if err != nil {
					return nil, err
				}

				if handled {
					return val, nil
				}
			}
			return jsonMarshalValue(fieldValue, f)
		}()
		if err != nil {
			return nil, err
		}

		result[fieldName] = value
	}

	return result, nil
}
