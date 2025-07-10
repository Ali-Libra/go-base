package http

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// 全局 HTTP 客户端连接池
var client *http.Client

func init() {
	// 创建可复用连接池的 Transport
	transport := &http.Transport{
		MaxIdleConns:          100,              // 最大空闲连接数
		MaxIdleConnsPerHost:   10,               // 每个主机最大空闲连接数
		IdleConnTimeout:       60 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,  // TCP连接超时
			KeepAlive: 30 * time.Second, // 保持活跃连接
		}).DialContext,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // 整体请求超时
	}
}

// GetClient 获取 HTTP 客户端（连接池）
func GetClient() *http.Client {
	return client
}

type RequestOption struct {
	Method  string
	URL     string
	Headers map[string]string
	Query   map[string]string
	Body    []byte
	Timeout time.Duration
}

func NewRequestWithOption(opt *RequestOption) (*http.Response, error) {
	// 解析 URL 并附加 query 参数
	str := opt.URL
	if opt.Query != nil {
		u, err := url.Parse(opt.URL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for key, value := range opt.Query {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		str = u.String()
	}

	// 构造请求体
	var bodyReader io.Reader
	if opt.Body != nil {
		bodyReader = bytes.NewReader(opt.Body)
	}

	// 创建请求
	req, err := http.NewRequest(opt.Method, str, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置 header
	for key, value := range opt.Headers {
		req.Header.Set(key, value)
	}

	// 创建 Client，设置超时（如需要）
	client := GetClient()
	if opt.Timeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), opt.Timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	return client.Do(req)
}
