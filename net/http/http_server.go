package http

import (
	"context"
	"html/template"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/Ali-Libra/go-base/logger"
)

type HandlerFunc func(*HttpResponse, *HttpRequest)

type HttpServer struct {
	server      *http.Server
	mux         *http.ServeMux
	timeout     time.Duration
	idleTimeout time.Duration
	middlewares []Middleware
	templates   *template.Template
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		mux:         http.NewServeMux(),
		timeout:     5 * time.Second,
		idleTimeout: 120 * time.Second,
		middlewares: make([]Middleware, 0),
	}
}

func (s *HttpServer) Run(port string) error {
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
	return s.server.ListenAndServe()
}

func (s *HttpServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}
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

// LoadHTMLGlob 加载HTML模板
func (s *HttpServer) LoadHTMLGlob(pattern string) error {
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return err
	}
	s.templates = tmpl
	return nil
}

// RenderHTML 渲染HTML模板并返回给客户端
func (s *HttpServer) RenderHTML(w http.ResponseWriter, statusCode int, templateName string, data interface{}) error {
	if s.templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	return s.templates.ExecuteTemplate(w, templateName, data)
}
