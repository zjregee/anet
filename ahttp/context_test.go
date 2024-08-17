package ahttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testJson  = `{"name":"test","age":18}`
	testForm  = `name=test&age=18`
	testQuery = `/?name=test&age=18`
)

type testJsonData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func newTestContextWithJson() *Context {
	req := httptest.NewRequest(http.MethodPost, testQuery, strings.NewReader(testJson))
	req.Header.Set("Content-Type", MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return newContext(req, rec)
}

func newTestContextWithForm() *Context {
	req := httptest.NewRequest(http.MethodPost, testQuery, strings.NewReader(testForm))
	req.Header.Set("Content-Type", MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return newContext(req, rec)
}

func TestContextRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := newContext(req, rec)
	assert.NotNil(t, c.Request())
	assert.Equal(t, req, c.Request())
}

func TestContextResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := newContext(req, rec)
	assert.NotNil(t, c.Response())
}

func TestContextQueryParam(t *testing.T) {
	c := newTestContextWithJson()
	assert.Equal(t, "test", c.QueryParam("name"))
	assert.Equal(t, "18", c.QueryParam("age"))
}

func TestContextQueryParams(t *testing.T) {
	c := newTestContextWithJson()
	assert.Equal(t, "test", c.QueryParams().Get("name"))
	assert.Equal(t, "18", c.QueryParams().Get("age"))
}

func TestContextQueryString(t *testing.T) {
	c := newTestContextWithJson()
	assert.Equal(t, testQuery[2:], c.QueryString())
}

func TestContextFormValue(t *testing.T) {
	c := newTestContextWithForm()
	assert.Equal(t, "test", c.FormValue("name"))
	assert.Equal(t, "18", c.FormValue("age"))
}

func TestContextFormParams(t *testing.T) {
	c := newTestContextWithForm()
	v, err := c.FormParams()
	assert.Nil(t, err)
	assert.Equal(t, "test", v.Get("name"))
	assert.Equal(t, "18", v.Get("age"))
}

func TestContextSetHandler(t *testing.T) {
	c := newTestContextWithJson()
	c.SetHandler(NotFoundHandler)
	assert.NotNil(t, c.handler)
}

func TestContextString(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := newContext(req, rec)
	_ = c.String(http.StatusOK, testJson)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testJson, rec.Body.String())
}

func TestContextJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := newContext(req, rec)
	_ = c.JSON(http.StatusOK, testJsonData{"test", 18})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testJson, rec.Body.String())
}

func TestContextReset(t *testing.T) {
	c := newTestContextWithJson()
	c.Reset(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	assert.Equal(t, http.MethodGet, c.Request().Method)
	assert.Equal(t, "/", c.Request().URL.Path)
}
