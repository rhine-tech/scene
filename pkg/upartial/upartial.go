package upartial

import (
	"errors"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

func UpdateStruct(src interface{}, dest interface{}) error {
	srcValue := reflect.ValueOf(src)
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return errors.New("dst value must be a pointer")
	}
	destValue = destValue.Elem()
	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}

	for i := 0; i < srcValue.NumField(); i++ {
		srcFieldDef := srcValue.Type().Field(i)
		srcFieldInfo := parseSrcTag(srcFieldDef)
		srcField := srcValue.Field(i)
		destField := destValue.FieldByName(srcFieldInfo.fieldName)

		if destField.Kind() == reflect.Invalid {
			continue
		}
		// if source field is not pointer type of destination field
		if !(srcField.Kind() == reflect.Ptr) {
			return errors.New("src field must be a pointer type of dst field: " + srcFieldInfo.fieldName)
		}
		// if source field is not nil and destination field is struct
		if destField.Kind() == reflect.Struct {
			if srcField.IsNil() {
				continue
			}
			err := UpdateStruct(srcField.Interface(), destField.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}
		if !(srcField.Type().Elem() == destField.Type()) {
			return errors.New("src field must be a pointer type of dst field: " + srcFieldInfo.fieldName)
		}
		// if source field is not nil, set destination field with source field value
		if !srcField.IsNil() {
			destField.Set(srcField.Elem())
			continue
		}
		// if its nil, but source field has default value, set destination field with default value
		if srcFieldInfo.hasDefault {
			destField.Set(reflect.ValueOf(parseValue(destField.Kind(), srcFieldInfo.defaultValue)))
			continue
		}
	}
	return nil
}

type fieldInfo struct {
	fieldName    string // field name in struct
	tag          string // field tag
	defaultValue string // default value
	hasDefault   bool   // has default value
}

func parseSrcTag(field reflect.StructField) (info fieldInfo) {
	info.fieldName = field.Name
	info.defaultValue = ""
	info.hasDefault = false
	tagValue, ok := field.Tag.Lookup("upartial")
	if !ok || tagValue == "" {
		return info
	}
	tagParts := strings.Split(tagValue, ",")
	for idx, tagPart := range tagParts {
		if strings.HasPrefix(tagPart, "default=") {
			info.defaultValue = strings.TrimPrefix(tagPart, "default=")
			info.hasDefault = true
			continue
		}
		if idx == 0 {
			info.fieldName = tagPart
		}
	}
	return info
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
	case reflect.Float64:
		return cast.ToFloat64(val)
	case reflect.Float32:
		return cast.ToFloat32(val)
	case reflect.Bool:
		return cast.ToBool(val)
	default:
		return val
	}
}
