package ws

import (
	"context"
	"go-base/logger"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RecvMessage struct {
	ConnId uint64
	Data   []byte
}

type SendMessage struct {
	ConnId uint64
	Data   []byte
}

type WsServer struct {
	server *http.Server
	mux    *http.ServeMux

	IdCount    uint64
	closeChan  chan uint64 // 用于关闭连接的通道
	closeRead  chan struct{}
	closeWrite chan struct{}

	onConnect func(conn *WsConn)
	onMessage func(conn *WsConn, msg []byte)
	onClose   func(conn *WsConn)
}

func NewWsServer() *WsServer {
	return &WsServer{
		mux:        http.NewServeMux(),
		closeChan:  make(chan uint64), // 初始化关闭通道
		closeRead:  make(chan struct{}),
		closeWrite: make(chan struct{}),
		IdCount:    0,
	}
}

func (s *WsServer) Server(port string) {
	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}
	s.mux.HandleFunc("/", s.wsHandler)
	s.server.ListenAndServe()
}

func (s *WsServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}
}

func (c *WsServer) SetOnConnect(callback func(conn *WsConn)) {
	c.onConnect = callback
}

func (c *WsServer) SetOnMessage(callback func(conn *WsConn, msg []byte)) {
	c.onMessage = callback
}

func (c *WsServer) SetOnClose(callback func(conn *WsConn)) {
	c.onClose = callback
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有跨域请求（生产建议限制）
		return true
	},
}

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	// 升级连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("websocket启动失败: %v", err)
		return
	}
	defer conn.Close()

	logger.Info("客户端已连接: %v", conn.RemoteAddr())

	// 循环读取消息
	wsConn := &WsConn{conn}
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway) {
				logger.Info("客户端断开: %v", conn.RemoteAddr())
			} else {
				logger.Error("读取消息失败: %v", err)
			}
			break
		}

		if msgType != websocket.TextMessage {
			logger.Error("不支持的消息类型:", msgType)
			continue
		}

		logger.Info("收到消息: %s", msg)
		if s.onMessage != nil {
			s.onMessage(wsConn, msg)
		}
	}
}
