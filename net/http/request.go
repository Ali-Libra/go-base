package http

import (
	"io"
	"net/http"
)

type HttpRequest struct {
	*http.Request
	contexts map[string]string
}

func (req *HttpRequest) ReadBody() ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (req *HttpRequest) SetContext(key string, value string) {
	if req.contexts == nil {
		req.contexts = make(map[string]string)
	}
	req.contexts[key] = value
}

func (req *HttpRequest) GetContext(key string) string {
	if req.contexts == nil {
		return ""
	}
	return req.contexts[key]
}
