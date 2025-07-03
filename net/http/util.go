package http

import "net/http"

type HandlerFunc func(*HttpResponse, *HttpRequest)

func WrapHandler(fn func(rsp *HttpResponse, req *HttpRequest)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rsp := &HttpResponse{
			ResponseWriter: w,
		}
		req := &HttpRequest{
			Request: r,
		}
		fn(rsp, req)
	}
}
