package http

import (
	"fmt"
	"go-base/logger"
	"net/http"
	"sync/atomic"
	"time"
)

type Middleware func(HandlerFunc) HandlerFunc

func Chain(f HandlerFunc, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		f = m(f)
	}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rsp := &HttpResponse{ResponseWriter: w}
		req := &HttpRequest{Request: r}

		defer func() {
			if err := recover(); err != nil && !rsp.success {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("%v", err)))
			}
		}()
		f(rsp, req)
	})

	return fn
}

func LoggingMiddleware(rsp *HttpResponse, req *HttpRequest) {
	logger.Debug("%s %s", req.Method, req.URL.Path)
}

var counter atomic.Int64
var counterAll atomic.Int64

func WatchMiddleware(rsp *HttpResponse, req *HttpRequest) {
	counter.Add(1)
	counterAll.Add(1)
}

func PrintWatchMiddleware() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			count := counter.Swap(0) // 原子读取并重置为0
			if count == 0 {
				continue
			}
			logger.Info("每秒请求次数:%d 总请求次数:%d", count, counterAll.Load())
		}
	}()
}
