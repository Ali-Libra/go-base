package http

import (
	"io"
	"net/http"
)

type HttpRequest struct {
	*http.Request // 组合原始请求
}

func (req *HttpRequest) ReadBody() ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
