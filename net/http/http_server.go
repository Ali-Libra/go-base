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

func (s *HttpServer) SetMiddleware(middlewares ...HandlerFunc) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		middle := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WrapHandler(middlewares[i])(w, r)
				next.ServeHTTP(w, r)
			})
		}

		s.middlewares = append(s.middlewares, middle)
	}
}

func (s *HttpServer) Handle(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	mws := append([]Middleware{LoggingMiddleware}, s.middlewares...)
	mws = append(mws, TimeoutHandler(s.timeout))
	for i := len(middlewares) - 1; i >= 0; i-- {
		middle := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WrapHandler(middlewares[i])(w, r)
				next.ServeHTTP(w, r)
			})
		}

		mws = append(mws, middle)
	}
	s.mux.Handle(pattern, Chain(WrapHandler(handler), mws...))
}

func (s *HttpServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}
}
