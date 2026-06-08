package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

type fieldWithTag struct {
	field      reflect.Value
	tag        string
	defaultVal reflect.Value
	hasDefault bool
}

type providerMarshaller struct {
	ConfigProvider
}

func NewConfigUnmarshaler(provider ConfigProvider) IConfig {
	return &providerMarshaller{ConfigProvider: provider}
}

func (m *providerMarshaller) Unmarshal(val interface{}) error {
	return m.UnmarshalWithPrefix("", val)
}

func (m *providerMarshaller) UnmarshalWithPrefix(prefix string, val interface{}) error {
	return UnmarshalWithOptions(m.ConfigProvider, val, WithPrefix(prefix))
}

func UnmarshalWithOptions(provider ConfigProvider, val interface{}, opts ...UnmarshalOption) error {
	if provider == nil {
		return errors.New("config unmarshal failed: provider is nil")
	}

	options := DefaultUnmarshalOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if options.TagName == "" {
		options.TagName = TagName
	}

	fields, err := getFieldWithTag(val, options.TagName)
	if err != nil {
		return err
	}
	for _, field := range fields {
		key := joinKey(options.Prefix, field.tag)
		defaultVal := reflect.Zero(field.field.Type())
		if options.DefaultBehavior == DefaultBehaviorUseTag && field.hasDefault {
			defaultVal = field.defaultVal
		}
		switch field.field.Kind() {
		case reflect.String:
			if v, ok := provider.GetStringE(key); ok {
				field.field.SetString(v)
			} else {
				field.field.Set(defaultVal)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v, ok := provider.GetIntE(key); ok {
				if !fitsSignedInt(v, field.field.Type().Bits()) {
					return fmt.Errorf("config unmarshal failed: value for %s overflows %s", key, field.field.Type())
				}
				field.field.SetInt(v)
			} else {
				field.field.Set(defaultVal)
			}
		case reflect.Bool:
			if v, ok := provider.GetBoolE(key); ok {
				field.field.SetBool(v)
			} else {
				field.field.Set(defaultVal)
			}
		default:
			return fmt.Errorf("config unmarshal failed: unsupported config field type %s for %s", field.field.Type(), key)
		}
	}
	return nil
}

func parseDefaultValue(t reflect.Type, val string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(val).Convert(t), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := cast.ToInt64E(val)
		if err != nil {
			return reflect.Value{}, err
		}
		if !fitsSignedInt(parsed, t.Bits()) {
			return reflect.Value{}, fmt.Errorf("%d overflows %s", parsed, t)
		}
		v := reflect.New(t).Elem()
		v.SetInt(parsed)
		return v, nil
	case reflect.Bool:
		parsed, err := cast.ToBoolE(val)
		if err != nil {
			return reflect.Value{}, err
		}
		v := reflect.New(t).Elem()
		v.SetBool(parsed)
		return v, nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported config field type %s", t)
	}
}

func fitsSignedInt(val int64, bits int) bool {
	if bits == 0 || bits >= 64 {
		return true
	}
	min := -(int64(1) << (bits - 1))
	max := (int64(1) << (bits - 1)) - 1
	return val >= min && val <= max
}

func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "." + key
}

func getFieldWithTag(val interface{}, tagName string) ([]fieldWithTag, error) {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, errors.New("config unmarshal failed: expecting a non-nil pointer")
	}

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return nil, errors.New("config unmarshal failed: expecting a pointer to struct")
	}
	t := e.Type()

	fields := make([]fieldWithTag, 0)
	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		tagVal := t.Field(i).Tag.Get(tagName)
		if tagVal == "" {
			continue
		}
		tagParts := strings.Split(tagVal, ",")
		tagKey := tagParts[0]
		if tagKey == "-" {
			continue
		}
		if !f.CanSet() {
			return nil, fmt.Errorf("config unmarshal failed: field %s is not settable", t.Field(i).Name)
		}

		field := fieldWithTag{
			field:      f,
			tag:        tagKey,
			defaultVal: reflect.Zero(f.Type()),
		}
		for _, tagPart := range tagParts[1:] {
			if strings.HasPrefix(tagPart, "default=") {
				defaultVal, err := parseDefaultValue(f.Type(), strings.TrimPrefix(tagPart, "default="))
				if err != nil {
					return nil, fmt.Errorf("config unmarshal failed: invalid default for %s: %w", t.Field(i).Name, err)
				}
				field.defaultVal = defaultVal
				field.hasDefault = true
				break
			}
		}
		fields = append(fields, field)
	}
	return fields, nil
}
