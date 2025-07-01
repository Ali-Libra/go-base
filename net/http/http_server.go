package http

import (
	"context"
	"go-base/logger"
	"net/http"
	"time"
)

type HttpServer struct {
	server      *http.Server
	mux         *http.ServeMux
	timeout     time.Duration
	middlewares []Middleware
}

func DefaultHttpServer() *HttpServer {
	return &HttpServer{
		mux:         http.NewServeMux(),
		timeout:     5 * time.Second,
		middlewares: make([]Middleware, 0),
	}
}

func (s *HttpServer) Run(port string) {
	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}
	s.server.ListenAndServe()
}

func (s *HttpServer) SetMiddleware(middleware ...Middleware) {
	s.middlewares = append(s.middlewares, middleware...)
}

func (s *HttpServer) Handle(pattern string, handler HandlerFunc) {
	mws := append([]Middleware{LoggingMiddleware}, s.middlewares...)
	mws = append(mws, TimeoutHandler(s.timeout))
	s.mux.Handle(pattern, Chain(WrapHandler(handler), mws...))
}

func (s *HttpServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}
}
