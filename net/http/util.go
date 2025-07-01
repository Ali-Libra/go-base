package http

import "net/http"

type HandlerFunc func(*HttpResponse, *HttpRequest)

func WrapHandler(fn func(w *HttpResponse, r *HttpRequest)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rsp := &HttpResponse{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		req := &HttpRequest{
			Request: r,
		}
		fn(rsp, req)
	}
}
