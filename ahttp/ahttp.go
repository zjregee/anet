package ahttp

import (
	"bufio"
	"context"
	"net/http"
	"sync"

	"github.com/zjregee/anet"
)

type HandlerFunc func(c *Context) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

var NotFoundHandler HandlerFunc = func(c *Context) error {
	return ErrNotFound
}

const (
	MIMEApplicationJSON      = "application/json"
	MIMEApplicationForm      = "application/x-www-form-urlencoded"
	MIMETextPlain            = "text/plain"
	MIMETextPlainCharsetUTF8 = MIMETextPlain + "; charset=utf-8"
)

type Server struct {
	router        *router
	pool          sync.Pool
	premiddleware []MiddlewareFunc
	middleware    []MiddlewareFunc
}

func New() *Server {
	s := &Server{}
	s.router = newRouter()
	s.pool.New = func() interface{} {
		return newContext(nil, nil)
	}
	return s
}

func (s *Server) Start(address string) error {
	listener, err := anet.CreateListener("tcp", address)
	if err != nil {
		panic("shouldn't failed here")
	}
	eventLoop, err := anet.NewEventLoop(s.handleConnection)
	if err != nil {
		panic("shouldn't failed here")
	}
	return eventLoop.Serve(listener)
}

func (s *Server) Pre(middleware ...MiddlewareFunc) {
	s.premiddleware = append(s.premiddleware, middleware...)
}

func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.middleware = append(s.middleware, middleware...)
}

func (s *Server) Add(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	s.router.add(path, func(c *Context) error {
		h := applyMiddleware(handler, middleware...)
		return h(c)
	})
}

func (s *Server) handleConnection(_ context.Context, connection anet.Connection) error {
	reader := bufio.NewReader(connection)
	writer := connection.Writer()
	for {
		request, err := http.ReadRequest(reader)
		if err != nil {
			return err
		}
		response := newResponse(connection)
		s.serveHTTP(response, request)
		err = writer.Flush()
		if err != nil {
			return err
		}
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.pool.Get().(*Context)
	c.Reset(r, w)
	var h HandlerFunc
	if s.premiddleware == nil {
		h = s.router.find(getPath(r))
		c.SetHandler(h)
		h = applyMiddleware(h, s.middleware...)
	} else {
		h = func(c *Context) error {
			h = s.router.find(getPath(r))
			c.SetHandler(h)
			h = applyMiddleware(h, s.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, s.premiddleware...)
	}
	_ = h(c)
	s.pool.Put(c)
}

func getPath(r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	return path
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}