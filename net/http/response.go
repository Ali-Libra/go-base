package http

import (
	"encoding/json"
	stdhttp "net/http"

	"github.com/Ali-Libra/go-base/logger"
)

type HttpResponse struct {
	stdhttp.ResponseWriter
	success bool
}

func (rsp *HttpResponse) SendError(rspTxt string) {
	logger.Error("HttpResponse Error: %s", rspTxt)
	panic(rspTxt)
}

func (rsp *HttpResponse) SendOK() {
	rsp.WriteHeader(200)
	rsp.success = true
	panic("success")
}
func (rsp *HttpResponse) SendCode(code int) {
	rsp.WriteHeader(code)
	rsp.success = true
	panic("success")
}
func (rsp *HttpResponse) SendJson(data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(200)
	rsp.success = true

	if data != nil {
		json.NewEncoder(rsp).Encode(data)
	}
	panic("success")
}

func (rsp *HttpResponse) SendHtml(statusCode int, html string) {
	rsp.Header().Set("Content-Type", "text/html; charset=utf-8")
	rsp.WriteHeader(statusCode)
	rsp.success = true
	rsp.Write([]byte(html))
}
