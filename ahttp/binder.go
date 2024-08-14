package ahttp

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

type DefaultBinder struct{}

func (b *DefaultBinder) BindHeaders(i interface{}, c *Context) error {
	if err := b.bindData(i, c.Request().Header); err != nil {
		return err
	}
	return nil
}

func (b *DefaultBinder) BindQueryParams(i interface{}, c *Context) error {
	if err := b.bindData(i, c.QueryParams()); err != nil {
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
		if err := b.bindData(i, params); err != nil {
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

func (b *DefaultBinder) bindData(dest interface{}, data map[string][]string) error {
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
	}
	return nil
}
