package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	server  *http.Server
	mux     *http.ServeMux
	timeout time.Duration
}

func DefaultHttpServer() *HttpServer {
	return &HttpServer{
		mux:     http.NewServeMux(),
		timeout: 5 * time.Second,
	}
}

func (s *HttpServer) Run(port string) {
	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}
	s.server.ListenAndServe()
}

func (s *HttpServer) Handle(pattern string, handler http.HandlerFunc) {
	s.mux.Handle(pattern, Chain(handler, LoggingMiddleware, TimeoutHandler(s.timeout)))
}

func (s *HttpServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
