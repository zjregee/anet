package ahttp

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type DefaultBinder struct{}

func (b *DefaultBinder) BindHeaders(i interface{}, c *Context) error {
	if err := b.bindData(i, c.Request().Header, "header"); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBinder) BindQueryParams(i interface{}, c *Context) error {
	if err := b.bindData(i, c.QueryParams(), "query"); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBinder) BindBody(i interface{}, c *Context) error {
	request := c.Request()
	if request.ContentLength == 0 {
		return nil
	}
	ctype := request.Header.Get("Content-Type")
	if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		if err := json.NewDecoder(request.Body).Decode(i); err != nil {
			switch err.(type) {
			case *HTTPError:
				return err
			default:
				return ErrBadRequest.SetInternal(err)
			}
		}
		return nil
	}
	if strings.HasPrefix(ctype, MIMEApplicationForm) {
		params, err := c.FormParams()
		if err != nil {
			return ErrBadRequest.SetInternal(err)
		}
		if err := b.bindData(i, params, "param"); err != nil {
			return ErrBadRequest.SetInternal(err)
		}
		return nil
	}
	return ErrUnsupportedMediaType
}

func (b *DefaultBinder) Bind(i interface{}, c *Context) error {
	method := c.Request().Method
	if method == http.MethodGet {
		return b.BindQueryParams(i, c)
	} else {
		return b.BindBody(i, c)
	}
}

func (b *DefaultBinder) bindData(dest interface{}, data map[string][]string, tag string) error {
	if dest == nil || len(data) == 0 {
		return nil
	}
	typ := reflect.TypeOf(dest)
	val := reflect.ValueOf(dest)
	if typ.Kind() == reflect.Map && typ.Key().Kind() == reflect.String {
		k := typ.Elem().Kind()
		isElemString := k == reflect.String
		isElemInterface := k == reflect.Interface
		ifElemSliceOfStrings := k == reflect.Slice && typ.Elem().Elem().Kind() == reflect.String
		if !(isElemString || isElemInterface || ifElemSliceOfStrings) {
			return nil
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
		return nil
	}
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return ErrUnsupportedMediaType
	}
	typ = typ.Elem()
	val = val.Elem()
	for i := 0; i < typ.NumField(); i++ {
		typField := typ.Field(i)
		inputFieldName := typField.Tag.Get(tag)
		if inputFieldName == "" {
			return ErrUnsupportedMediaType
		}
		inputFiledValue, ok := data[inputFieldName]
		if !ok {
			return ErrUnsupportedMediaType
		}
		valField := val.Field(i)
		valFieldKind := valField.Kind()
		if err := setWithProperType(valFieldKind, inputFiledValue[0], valField); err != nil {
			return err
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structFiled reflect.Value) error {
	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structFiled.Elem().Kind(), val, structFiled.Elem())
	case reflect.Int:
		return setIntFiled(val, 0, structFiled)
	case reflect.Int8:
		return setIntFiled(val, 8, structFiled)
	case reflect.Int16:
		return setIntFiled(val, 16, structFiled)
	case reflect.Int32:
		return setIntFiled(val, 32, structFiled)
	case reflect.Int64:
		return setIntFiled(val, 64, structFiled)
	case reflect.Uint:
		return setUintFiled(val, 0, structFiled)
	case reflect.Uint8:
		return setUintFiled(val, 8, structFiled)
	case reflect.Uint16:
		return setUintFiled(val, 16, structFiled)
	case reflect.Uint32:
		return setUintFiled(val, 32, structFiled)
	case reflect.Uint64:
		return setUintFiled(val, 64, structFiled)
	case reflect.Bool:
		return setBoolFiled(val, structFiled)
	case reflect.Float32:
		return setFloatFiled(val, 32, structFiled)
	case reflect.Float64:
		return setFloatFiled(val, 64, structFiled)
	case reflect.String:
		structFiled.SetString(val)
	default:
		return ErrUnsupportedMediaType
	}
	return nil
}

func setIntFiled(value string, bitSize int, filed reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		filed.SetInt(intVal)
	}
	return err
}

func setUintFiled(value string, bitSize int, filed reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		filed.SetUint(uintVal)
	}
	return err
}

func setBoolFiled(value string, filed reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		filed.SetBool(boolVal)
	}
	return err
}

func setFloatFiled(value string, bitSize int, filed reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		filed.SetFloat(floatVal)
	}
	return err
}
