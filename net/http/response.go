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
	panic(rspTxt)
}
func (rsp *HttpResponse) RspJson(data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(200)

	json.NewEncoder(rsp).Encode(data)
}
