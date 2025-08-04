package http

import (
	"io"
	"net/http"
	"strconv"
)

type HttpRequest struct {
	*http.Request
	contexts map[string]interface{}
}

func (req *HttpRequest) ReadBody() ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (req *HttpRequest) SetContext(key string, value interface{}) {
	if req.contexts == nil {
		req.contexts = make(map[string]interface{})
	}
	req.contexts[key] = value
}

func (req *HttpRequest) GetContext(key string) interface{} {
	if req.contexts == nil {
		return nil
	}
	return req.contexts[key]
}

func (req *HttpRequest) GetContextString(key string) string {
	if val, ok := req.contexts[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (req *HttpRequest) GetContextUint64(key string) uint64 {
	if val, ok := req.contexts[key]; ok {
		switch v := val.(type) {
		case uint64:
			return v
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func (req *HttpRequest) GetContextUint32(key string) uint32 {
	if val, ok := req.contexts[key]; ok {
		switch v := val.(type) {
		case uint32:
			return v
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
				return uint32(parsed)
			}
		}
	}
	return 0
}
