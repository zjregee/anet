package ahttp

import (
	"bufio"
	"context"
	"net"
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
	router     *router
	pool       sync.Pool
	listener   net.Listener
	eventLoop  anet.EventLoop
	middleware []MiddlewareFunc
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
	s.listener = listener
	s.eventLoop = eventLoop
	return s.eventLoop.Serve(listener)
}

func (s *Server) Shutdown() error {
	_ = s.eventLoop.Shutdown(context.Background())
	_ = s.listener.Close()
	return nil
}

func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.middleware = append(s.middleware, middleware...)
}

func (s *Server) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	s.add(http.MethodGet, path, handler, middleware...)
}

func (s *Server) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	s.add(http.MethodPost, path, handler, middleware...)
}

func (s *Server) add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	s.router.add(method, path, func(c *Context) error {
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
	h, params := s.router.find(r.Method, getPath(r))
	if h != nil {
		for k, v := range params {
			c.Set(k, v)
		}
	} else {
		h = NotFoundHandler
	}
	c.SetHandler(h)
	h = applyMiddleware(h, s.middleware...)
	_ = h(c)
	c.Reset(nil, nil)
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
