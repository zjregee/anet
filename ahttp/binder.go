package ahttp

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type DefaultBinder struct{}

func (b *DefaultBinder) BindHeaders(c *Context, i interface{}) error {
	return b.bindData(i, c.Request().Header, "header")
}

func (b *DefaultBinder) BindPathParams(c *Context, i interface{}) error {
	names := c.ParamNames()
	values := c.ParamValues()
	params := map[string][]string{}
	for i, name := range names {
		params[name] = []string{values[i]}
	}
	return b.bindData(i, params, "param")
}

func (b *DefaultBinder) BindQueryParams(c *Context, i interface{}) error {
	return b.bindData(i, c.QueryParams(), "query")
}

func (b *DefaultBinder) BindBody(c *Context, i interface{}) error {
	request := c.Request()
	if request.ContentLength == 0 {
		return nil
	}
	ctype := request.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, "application/json") {
		return json.NewDecoder(c.Request().Body).Decode(i)
	}
	return errors.New("unsupported type")
}

func (b *DefaultBinder) bindData(dest interface{}, data map[string][]string, tag string) error {
	if dest == nil || len(data) == 0 {
		return nil
	}
	typ := reflect.TypeOf(dest).Elem()
	val := reflect.ValueOf(dest).Elem()

	if typ.Kind() == reflect.Map && typ.Key().Kind() == reflect.String {
		k := typ.Elem().Kind()
		isElemString := k == reflect.String
		isElemInterface := k == reflect.Interface
		ifElemSliceOfStrings := k == reflect.Slice && typ.Elem().Elem().Kind() == reflect.String
		if !(isElemString || isElemInterface || ifElemSliceOfStrings) {
			return nil
		}
		if val.IsNil() {
			val.Set(reflect.MakeMap(typ))
		}
		for k, v := range data {
			if isElemString {
				val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
			} else if isElemInterface {
				val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
			} else {
				val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
		}
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if typeField.Anonymous {
			if structField.Kind() == reflect.Ptr {
				structField = structField.Elem()
			}
		}
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)
		if typeField.Anonymous && structFieldKind == reflect.Struct && inputFieldName != "" {
			return nil
		}

		if inputFieldName == "" {
			continue
		}

		inputValue, exists := data[inputFieldName]

		if !exists {
			continue
		}

		if ok, err := unmarshalInputToField(typeField.Type.Kind(), inputValue[0], structField); ok {
			if err != nil {
				return err
			}
			continue
		}

		if structFieldKind == reflect.Pointer {
			structFieldKind = structField.Elem().Kind()
			structField = structField.Elem()
		}

		if structFieldKind == reflect.Slice {
			sliceOf := structField.Type().Elem().Kind()
			numElems := len(inputValue)
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			structField.Set(slice)
			continue
		}

		if err := setWithProperType(structFieldKind, inputValue[0], structField); err != nil {
			return err
		}
	}

	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	if ok, err := unmarshalInputToField(valueKind, val, structField); ok {
		return err
	}
	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
		return nil
	default:
		return errors.New("unsupported type")
	}
}

func unmarshalInputToField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	if valueKind == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}
	fieldIValue := field.Addr().Interface()
	switch unmarshaler := fieldIValue.(type) {
	case encoding.TextUnmarshaler:
		return true, unmarshaler.UnmarshalText([]byte(val))
	}
	return false, nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(intVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
