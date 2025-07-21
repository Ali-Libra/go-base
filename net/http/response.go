package http

import (
	"encoding/json"
	"go-base/logger"
	"net/http"
)

type HttpResponse struct {
	http.ResponseWriter
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
func (rsp *HttpResponse) SendJson(data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(200)
	rsp.success = true

	if data != nil {
		json.NewEncoder(rsp).Encode(data)
	}
	panic("success")
}
