package ahttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	r := newRouter()
	r.add(http.MethodGet, "/test/get", func(c *Context) error {
		return nil
	})
	r.add(http.MethodPost, "/test/post", func(c *Context) error {
		return nil
	})
	h, _ := r.find(http.MethodGet, "/test/get")
	assert.NotNil(t, h)
	h, _ = r.find(http.MethodPost, "/test/post")
	assert.NotNil(t, h)
	h, _ = r.find(http.MethodGet, "/notfound")
	assert.Nil(t, h)
}

func TestRouterWithParam(t *testing.T) {
	r := newRouter()
	r.add(http.MethodGet, "/test/name/:name", func(c *Context) error {
		return nil
	})
	r.add(http.MethodGet, "/test/name/:name/age/:age", func(c *Context) error {
		return nil
	})
	h, params := r.find(http.MethodGet, "/test/name/abc")
	assert.NotNil(t, h)
	assert.Equal(t, "abc", params["name"])
	h, params = r.find(http.MethodGet, "/test/name/abc/age/10")
	assert.NotNil(t, h)
	assert.Equal(t, "abc", params["name"])
	assert.Equal(t, "10", params["age"])
}

func TestRouterWithWild(t *testing.T) {
	r := newRouter()
	r.add(http.MethodGet, "/test/*", func(c *Context) error {
		return nil
	})
	h, _ := r.find(http.MethodGet, "/test/wild")
	assert.NotNil(t, h)
}
