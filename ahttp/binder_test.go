package ahttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinderBindHeaders(t *testing.T) {
	c := newTestContextWithJson()
	b := &DefaultBinder{}
	bindData := map[string]string{}
	_ = b.BindHeaders(bindData, c)
	assert.Equal(t, MIMEApplicationJSON, bindData["Content-Type"])
}

func TestBinderBindQueryParams(t *testing.T) {
	c := newTestContextWithJson()
	b := &DefaultBinder{}
	bindData := map[string]string{}
	_ = b.BindQueryParams(bindData, c)
	assert.Equal(t, "test", bindData["name"])
	assert.Equal(t, "18", bindData["age"])
}

func TestBinderBindBodyJson(t *testing.T) {
	c := newTestContextWithJson()
	b := &DefaultBinder{}
	bindData := &testJsonData{}
	_ = b.BindBody(bindData, c)
	assert.Equal(t, "test", bindData.Name)
	assert.Equal(t, 18, bindData.Age)
}

func TestBinderBindBodyForm(t *testing.T) {
	c := newTestContextWithForm()
	b := &DefaultBinder{}
	bindData := map[string]string{}
	_ = b.BindBody(bindData, c)
	assert.Equal(t, "test", bindData["name"])
	assert.Equal(t, "18", bindData["age"])
}
