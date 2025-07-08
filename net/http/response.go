package http

import (
	"encoding/json"
	"go-base/logger"
	"net/http"
)

type HttpResponse struct {
	http.ResponseWriter
}

func (rsp *HttpResponse) SendError(rspTxt string) {
	logger.Error("HttpResponse Error: %s", rspTxt)
	panic(rspTxt)
}
func (rsp *HttpResponse) SendJson(data interface{}) {
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(200)

	json.NewEncoder(rsp).Encode(data)
}
