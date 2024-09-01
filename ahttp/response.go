package ahttp

import (
	"bytes"
	"io"
	"net/http"
)

func newResponse(w io.Writer) *Response {
	return &Response{
		r: &http.Response{
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		},
		w: w,
	}
}

type Response struct {
	r *http.Response
	w io.Writer
}

var _ http.ResponseWriter = &Response{}

func (r *Response) Header() http.Header {
	if r.r.Header == nil {
		r.r.Header = make(http.Header)
	}
	return r.r.Header
}

func (r *Response) WriteHeader(code int) {
	r.r.StatusCode = code
}

func (r *Response) Write(b []byte) (int, error) {
	r.r.Body = io.NopCloser(bytes.NewReader(b))
	err := r.r.Write(r.w)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
