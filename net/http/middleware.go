package http

import (
	"fmt"
	"go-base/logger"
	"net/http"
	"time"
)

type Middleware func(HandlerFunc) HandlerFunc

func Chain(timeout time.Duration, f HandlerFunc, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		f = m(f)
	}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rsp := &HttpResponse{ResponseWriter: w}
		req := &HttpRequest{Request: r}

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("%v", err)))
			}
		}()
		f(rsp, req)

		{
			// select {
			// case <-ctx.Done():
			// 	w.WriteHeader(http.StatusGatewayTimeout)
			// 	w.Write([]byte("timeout"))
			// case <-done:
			// }

			// rsp := &HttpResponse{ResponseWriter: w}
			// ctx, cancel := context.WithTimeout(r.Context(), timeout) // 设置超时时间
			// defer cancel()

			// req := &HttpRequest{
			// 	Request: r.WithContext(ctx),
			// }

			// done := make(chan struct{})
			// go func() {
			// 	defer func() {
			// 		if err := recover(); err != nil {
			// 			rsp.RspError(fmt.Sprintf("%v", err))
			// 		}
			// 		close(done)
			// 	}()
			// 	f(rsp, req)
			// }()

			// select {
			// case <-ctx.Done():
			// 	w.WriteHeader(http.StatusGatewayTimeout)
			// 	w.Write([]byte("timeout"))
			// case <-done:
			// }
		}
	})

	return fn
}

func LoggingMiddleware(rsp *HttpResponse, req *HttpRequest) {
	logger.Debug("%s %s", req.Method, req.URL.Path)
}
