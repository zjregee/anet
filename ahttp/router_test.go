package ahttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	r := newRouter()
	r.add("/test", func(c *Context) error {
		return nil
	})
	assert.NotNil(t, r.find("/test"))
	assert.Nil(t, r.find("/notfound"))
}
