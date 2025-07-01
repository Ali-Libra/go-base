package http

import "net/http"

type HttpRequest struct {
	*http.Request // 组合原始请求
}
