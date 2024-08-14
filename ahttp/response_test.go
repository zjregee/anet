package ahttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	r := newResponse(nil)
	assert.NotNil(t, r.Header())
	r.Header().Set("Content-Type", MIMEApplicationJSON)
	assert.Equal(t, MIMEApplicationJSON, r.Header().Get("Content-Type"))
	r.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, r.r.StatusCode)
}
