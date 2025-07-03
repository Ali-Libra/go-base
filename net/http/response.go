package http

import (
	"encoding/json"
	"go-base/logger"
	"net/http"
)

type HttpResponse struct {
	http.ResponseWriter
}

func (rsp *HttpResponse) RspError(rspTxt string) {
	logger.Error("HttpResponse Error: %s", rspTxt)
	http.Error(rsp, rspTxt, 500)
}
func (rsp *HttpResponse) RspJson(data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(200)

	json.NewEncoder(rsp).Encode(data)
}
