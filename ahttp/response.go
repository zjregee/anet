package ahttp

import (
	"bytes"
	"io"
	"net/http"
)

func newResponse(w io.Writer) *response {
	return &response{
		r: &http.Response{},
		w: w,
	}
}

type response struct {
	r *http.Response
	w io.Writer
}

var _ http.ResponseWriter = &response{}

func (r *response) Header() http.Header {
	panic("unreachable code")
}

func (r *response) WriteHeader(code int) {
	r.r.StatusCode = code
}

func (r *response) Write(b []byte) (int, error) {
	r.r.Body = io.NopCloser(bytes.NewReader(b))
	err := r.r.Write(r.w)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
