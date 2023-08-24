package cfgur

import (
	"errors"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

type fieldWithTag struct {
	field    reflect.Value
	tag      string
	defValue interface{}
}

func parseValue(kind reflect.Kind, val string) interface{} {
	switch kind {
	case reflect.String:
		return val
	case reflect.Int:
		return cast.ToInt(val)
	case reflect.Int8:
		return cast.ToInt8(val)
	case reflect.Int16:
		return cast.ToInt16(val)
	case reflect.Int32:
		return cast.ToInt32(val)
	case reflect.Int64:
		return cast.ToInt64(val)
	case reflect.Bool:
		return cast.ToBool(val)
	default:
		return val
	}
}

func getFieldWithTag(val interface{}) ([]fieldWithTag, error) {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, errors.New("cfgur unmarshal failed: expecting a non-nil pointer")
	}

	e := v.Elem()
	t := e.Type()

	fields := make([]fieldWithTag, 0)

	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		tagVal := t.Field(i).Tag.Get("cfgur")
		if tagVal != "" {
			tagParts := strings.Split(tagVal, ",")
			tagKey := tagParts[0]

			var defValue interface{}
			if len(tagParts) > 1 && strings.HasPrefix(tagParts[1], "default=") {
				defValue = parseValue(f.Kind(), strings.TrimPrefix(tagParts[1], "default="))
			} else {
				defValue = reflect.Zero(f.Type()).Interface()
			}

			fields = append(fields, fieldWithTag{
				field:    f,
				tag:      tagKey,
				defValue: defValue,
			})
		}
	}
	return fields, nil
}

type commonMarshaller struct {
	ConfigProvider
}

func (m *commonMarshaller) Unmarshal(val interface{}) error {
	fields, err := getFieldWithTag(val)
	if err != nil {
		return err
	}
	for _, field := range fields {
		switch field.field.Kind() {
		case reflect.String:
			if v, ok := m.ConfigProvider.GetStringE(field.tag); ok {
				field.field.SetString(v)
			} else {
				field.field.SetString(field.defValue.(string))
			}
		case reflect.Int:
			if v, ok := m.ConfigProvider.GetIntE(field.tag); ok {
				field.field.SetInt(v)
			} else {
				field.field.SetInt(field.defValue.(int64))
			}
		case reflect.Bool:
			if v, ok := m.ConfigProvider.GetBoolE(field.tag); ok {
				field.field.SetBool(v)
			} else {
				field.field.SetBool(field.defValue.(bool))
			}
		}
	}
	return nil
}
