package http

import (
	"context"
	"go-base/logger"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

type HandlerFunc func(*HttpResponse, *HttpRequest)

type HttpServer struct {
	server      *http.Server
	mux         *http.ServeMux
	timeout     time.Duration
	idleTimeout time.Duration
	middlewares []Middleware
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		mux:         http.NewServeMux(),
		timeout:     5 * time.Second,
		idleTimeout: 120 * time.Second,
		middlewares: make([]Middleware, 0),
	}
}

func (s *HttpServer) Run(port string) {
	s.server = &http.Server{
		Addr:         port,
		Handler:      s.mux,
		ReadTimeout:  s.timeout,
		WriteTimeout: s.timeout,
		IdleTimeout:  s.idleTimeout,
		ConnState: func(conn net.Conn, state http.ConnState) {
			logger.Info("conn %v state: %v", conn.RemoteAddr(), state)
		},
	}
	s.server.ListenAndServe()
}

func (s *HttpServer) SetMiddleware(middleHandler HandlerFunc) {
	middleware := func(next HandlerFunc) HandlerFunc {
		return func(rsp *HttpResponse, req *HttpRequest) {
			middleHandler(rsp, req)
			next(rsp, req)
		}
	}
	fnName := runtime.FuncForPC(reflect.ValueOf(middleHandler).Pointer()).Name()
	if fnName == "go-base/net/http.WatchMiddleware" {
		PrintWatchMiddleware()
	}
	s.middlewares = append(s.middlewares, middleware)
}

func (s *HttpServer) Handle(pattern string, handler HandlerFunc, middleHandlers ...HandlerFunc) {
	mws := make([]Middleware, 0)
	mws = append(mws, s.middlewares...)
	for i := len(middleHandlers) - 1; i >= 0; i-- {
		middle := func(next HandlerFunc) HandlerFunc {
			return func(rsp *HttpResponse, req *HttpRequest) {
				middleHandlers[i](rsp, req)
				next(rsp, req)
			}
		}

		mws = append(mws, middle)
	}

	s.mux.Handle(pattern, Chain(handler, mws...))
}

func (s *HttpServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}
}
