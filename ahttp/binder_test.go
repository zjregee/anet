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

func TestBinderBindStruct(t *testing.T) {
	c := newTestContextWithJson()
	b := &DefaultBinder{}

	type testQueryData struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}
	bindQueryData := &testQueryData{}
	_ = b.BindQueryParams(bindQueryData, c)
	assert.Equal(t, "test", bindQueryData.Name)
	assert.Equal(t, 18, bindQueryData.Age)

	type testFormData struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}
	bindFormData := &testFormData{}
	_ = b.BindBody(bindFormData, c)
	assert.Equal(t, "test", bindFormData.Name)
	assert.Equal(t, 18, bindFormData.Age)

	type testHeaderData struct {
		ContentType string `header:"Content-Type"`
	}
	bindHeaderData := &testHeaderData{}
	_ = b.BindHeaders(bindHeaderData, c)
	assert.Equal(t, MIMEApplicationJSON, bindHeaderData.ContentType)
}
