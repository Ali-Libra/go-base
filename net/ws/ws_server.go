package ws

import (
	"context"
	"go-base/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type RecvMessage struct {
	conn    *WsConn
	MsgType int
	Data    []byte
}

type SendMessage struct {
	ConnId  uint64
	MsgType int
	Data    []byte
}

type WsServer struct {
	server *http.Server
	mux    *http.ServeMux
	limit  uint32

	rwLock    sync.RWMutex
	IdCount   uint64
	conns     map[uint64]*websocket.Conn
	recvChan  chan *RecvMessage // 用于接收客户端消息的通道
	sendChan  chan *SendMessage // 用于发送消息到客户端的通道
	closeChan chan uint64       // 用于关闭连接的通道

	onConnect func(conn uint64)
	onMessage func(conn *WsConn, msg []byte)
	onClose   func(conn uint64)
}

func NewWsServer() *WsServer {
	return &WsServer{
		mux:       http.NewServeMux(),
		conns:     make(map[uint64]*websocket.Conn),
		recvChan:  make(chan *RecvMessage, 10240), // 初始化消息通道
		sendChan:  make(chan *SendMessage, 10240), // 初始化发送通道映射
		closeChan: make(chan uint64, 1024),        // 初始化关闭通道
		IdCount:   0,
	}
}

func (s *WsServer) Run(port string, path string) {
	go s.handleWrite()

	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}
	s.mux.HandleFunc("/"+path, s.wsHandler)
	go s.server.ListenAndServe()
}

func (s *WsServer) OnLoop() {
	for {
		select {
		case msg := <-s.recvChan: // 从接收通道读取数据
			if s.onMessage != nil {
				s.onMessage(msg.conn, msg.Data) // 调用接收消息的回调函数
			}
			for {
				select {
				case msg = <-s.recvChan: // 继续读取接收通道中的数据
					if s.onMessage != nil {
						s.onMessage(msg.conn, msg.Data) // 调用接收消息的回调函数
					}
				default:
					return
				}
			}
		case conn := <-s.closeChan: // 监听关闭信号
			if s.onClose != nil {
				s.onClose(conn) // 调用连接关闭的回调函数
			}
		default:
			return
		}
	}
}

func (s *WsServer) Close() {
	close(s.sendChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}

	for _, conn := range s.conns {
		conn.Close()
	}
}

func (c *WsServer) SetOnConnect(callback func(conn uint64)) {
	c.onConnect = callback
}

func (c *WsServer) SetOnMessage(callback func(conn *WsConn, msg []byte)) {
	c.onMessage = callback
}

func (c *WsServer) SetOnClose(callback func(conn uint64)) {
	c.onClose = callback
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有跨域请求（生产建议限制）
		return true
	},
}

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("websocket启动失败: %v", err)
		return
	}

	var connId uint64
	defer func() {
		logger.Info("连接已断开: %v", conn.RemoteAddr())
		conn.Close()
		s.rwLock.Lock()
		delete(s.conns, connId)
		s.rwLock.Unlock()
		s.closeChan <- connId
	}()
	logger.Info("客户端已连接: %v", conn.RemoteAddr())

	s.rwLock.Lock()
	s.IdCount++
	connId = s.IdCount
	s.conns[connId] = conn
	s.rwLock.Unlock()

	wsConn := &WsConn{
		ConnId:   connId,
		sendChan: s.sendChan,
	}
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
			return
		}

		s.recvChan <- &RecvMessage{
			conn:    wsConn, // 将连接和消息数据封装到 ConnMessage 中
			MsgType: msgType,
			Data:    msg,
		}
	}
}

func (s *WsServer) handleWrite() {
	for msg := range s.sendChan { // 从发送通道读取数据
		s.rwLock.RLock()
		conn, ok := s.conns[msg.ConnId]
		s.rwLock.RUnlock()
		if !ok {
			logger.Error("connect %d have  closed", msg.ConnId)
			continue
		}

		err := conn.WriteMessage(msg.MsgType, msg.Data)
		if err != nil {
			logger.Error("connect %d have  write error", msg.ConnId)
			continue
		}
	}
}
