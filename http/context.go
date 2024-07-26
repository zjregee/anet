package http

import (
	"net/http"
	"net/url"
)

type HandlerFunc func(c Context) error

type Context interface {
	Request() *http.Request
	SetRequest(r *http.Request)
	Response() http.ResponseWriter
	SetResponse(r http.ResponseWriter)

	Path() string
	SetPath(p string)

	Param(name string) string
	ParamNames() []string
	SetParamNames(names ...string)
	ParamValues() []string
	SetParamValues(values ...string)

	QueryParam(name string) string
	QueryParams() url.Values
	QueryString() string

	Bind(i interface{}) error

	String(code int, s string) error
	Blob(code int, contentType string, b []byte) error
	NoContent(code int) error

	Handler() HandlerFunc
	SetHandler(h HandlerFunc)
	Reset(r *http.Request, w http.ResponseWriter)
}

var _ Context = &context{}

type context struct {
	request  *http.Request
	response http.ResponseWriter
	query    url.Values
	handler  HandlerFunc
	path     string
	pnames   []string
	pvalues  []string
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) Response() http.ResponseWriter {
	return c.response
}

func (c *context) SetResponse(r http.ResponseWriter) {
	c.response = r
}

func (c *context) Path() string {
	return c.path
}

func (c *context) SetPath(p string) {
	c.path = p
}

func (c *context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) && n == name {
			return c.pvalues[i]
		}
	}
	return ""
}

func (c *context) ParamNames() []string {
	return c.pnames
}

func (c *context) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *context) ParamValues() []string {
	return c.pvalues[:len(c.pnames)]
}

func (c *context) SetParamValues(values ...string) {
	c.pvalues = values
}

func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *context) Bind(i interface{}) error {
	return nil
}

func (c *context) String(code int, s string) error {
	return c.Blob(code, "text/plain; charset=UTF-8", []byte(s))
}

func (c *context) Blob(code int, contentType string, b []byte) error {
	header := c.Response().Header()
	header.Set("Content-Type", contentType)
	c.response.WriteHeader(code)
	_, err := c.response.Write(b)
	return err
}

func (c *context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

func (c *context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *context) Reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response = w
	c.query = nil
	c.handler = nil
	c.path = ""
	c.pnames = nil
	c.pvalues = nil
}
