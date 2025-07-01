package http

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HttpResponse struct {
	http.ResponseWriter
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func (rsp *HttpResponse) Header() http.Header {
	return rsp.header
}

func (rsp *HttpResponse) Write(data []byte) (int, error) {
	return rsp.ResponseWriter.Write(data)
}
func (rsp *HttpResponse) WriteHeader(statusCode int) {
	rsp.statusCode = statusCode
	rsp.ResponseWriter.WriteHeader(statusCode)
}
func (rsp *HttpResponse) WriteJson(code int, data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(code)

	json.NewEncoder(rsp).Encode(data)
}
