package http

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type HttpRequest struct {
	*http.Request
	contexts    map[string]interface{}
	queryValues url.Values
	queryParsed bool
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

func (req *HttpRequest) getQueryValues() url.Values {
	if req.queryParsed {
		return req.queryValues
	}
	req.queryParsed = true
	if req.URL == nil {
		req.queryValues = url.Values{}
		return req.queryValues
	}
	req.queryValues = req.URL.Query()
	return req.queryValues
}

func (req *HttpRequest) GetQuery(key string) string {
	return req.getQueryValues().Get(key)
}

func (req *HttpRequest) GetQueryUint64(key string) uint64 {
	val := req.GetQuery(key)
	if val == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func (req *HttpRequest) GetQueryUint32(key string) uint32 {
	val := req.GetQuery(key)
	if val == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(parsed)
}
