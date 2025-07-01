package http

import (
	"go-base/logger"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Chain(f http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func LoggingMiddleware(next http.Handler) http.Handler {
	handler := WrapHandler(func(rsp *HttpResponse, req *HttpRequest) {
		logger.Debug("%s %s", req.Method, req.URL.Path)
		next.ServeHTTP(rsp, req.Request)
	})
	return http.Handler(handler)
}

func TimeoutHandler(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, "Request timed out\n")
	}
}
