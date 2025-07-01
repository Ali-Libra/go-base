package http

import (
	"encoding/json"
	"go-base/logger"
	"net/http"
)

type HttpResponse struct {
	http.ResponseWriter
}

func (rsp *HttpResponse) RspError(txt string) {
	logger.Error("HttpResponse Error: %s", txt)
	http.Error(rsp, txt, 500)
}
func (rsp *HttpResponse) RspJson(code int, data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(code)

	json.NewEncoder(rsp).Encode(data)
}
