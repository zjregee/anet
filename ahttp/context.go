package ahttp

import (
	"net/http"
	"net/url"
)

func newContext() *Context {
	return &Context{}
}

type Context struct {
	request  *http.Request
	response http.ResponseWriter
	query    url.Values
	handler  HandlerFunc
	path     string
	pnames   []string
	pvalues  []string
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *Context) Response() http.ResponseWriter {
	return c.response
}

func (c *Context) SetResponse(r http.ResponseWriter) {
	c.response = r
}

func (c *Context) Path() string {
	return c.path
}

func (c *Context) SetPath(p string) {
	c.path = p
}

func (c *Context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) && n == name {
			return c.pvalues[i]
		}
	}
	return ""
}

func (c *Context) ParamNames() []string {
	return c.pnames
}

func (c *Context) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *Context) ParamValues() []string {
	return c.pvalues[:len(c.pnames)]
}

func (c *Context) SetParamValues(values ...string) {
	c.pvalues = values
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

func (c *Context) Bind(i interface{}) error {
	return nil
}

func (c *Context) String(code int, s string) error {
	return c.Blob(code, "text/plain; charset=UTF-8", []byte(s))
}

func (c *Context) Blob(code int, contentType string, b []byte) error {
	header := c.Response().Header()
	header.Set("Content-Type", contentType)
	c.response.WriteHeader(code)
	_, err := c.response.Write(b)
	return err
}

func (c *Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *Context) Handler() HandlerFunc {
	return c.handler
}

func (c *Context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *Context) Reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response = w
	c.query = nil
	c.handler = nil
	c.path = ""
	c.pnames = nil
	c.pvalues = nil
}
