package ahttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func newContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		request:  r,
		response: w,
	}
}

type Context struct {
	request  *http.Request
	response http.ResponseWriter
	query    url.Values
	handler  HandlerFunc
	store    map[string]interface{}
}

func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

func (c *Context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = val
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) Response() http.ResponseWriter {
	return c.response
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

func (c *Context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *Context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *Context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *Context) FormParams() (url.Values, error) {
	if err := c.request.ParseForm(); err != nil {
		return nil, err
	}
	return c.request.Form, nil
}

func (c *Context) Handler() HandlerFunc {
	return c.handler
}

func (c *Context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *Context) String(code int, s string) error {
	return c.blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *Context) JSON(code int, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.blob(code, MIMEApplicationJSON, b)
}

func (c *Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	_, err := c.response.Write(nil)
	return err
}

func (c *Context) Reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response = w
	c.query = nil
	c.handler = nil
	c.store = nil
}

func (c *Context) blob(code int, contentType string, b []byte) error {
	header := c.Response().Header()
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", fmt.Sprint(len(b)))
	c.response.WriteHeader(code)
	_, err := c.response.Write(b)
	return err
}
