package ws

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Ali-Libra/go-base/logger"
	"github.com/gorilla/websocket"
)

type RecvMessage struct {
	conn    *WsConn
	MsgType int
	Data    []byte
}

type SendMessage struct {
	MsgType int
	Data    []byte
}

type WsServer struct {
	server *http.Server
	mux    *http.ServeMux

	rwLock  sync.RWMutex
	IdCount uint64
	conns   map[uint64]*WsConn

	connChan  chan *WsConn
	recvChan  chan *RecvMessage
	closeChan chan uint64

	onConnect func(conn *WsConn)
	onMessage func(conn *WsConn, msg []byte)
	onClose   func(conn uint64)
}

func NewWsServer() *WsServer {
	return &WsServer{
		mux:       http.NewServeMux(),
		conns:     make(map[uint64]*WsConn),
		connChan:  make(chan *WsConn, 1024),
		recvChan:  make(chan *RecvMessage, 10240),
		closeChan: make(chan uint64, 1024),
	}
}

func (s *WsServer) Run(port string, path string) {
	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}
	s.mux.HandleFunc("/"+path, s.wsHandler)

	go s.server.ListenAndServe()
	go s.loop()
}

func (s *WsServer) loop() {
	for {
		select {
		case conn := <-s.connChan:
			if s.onConnect != nil {
				s.onConnect(conn)
			}

		case msg := <-s.recvChan:
			if s.onMessage != nil && !msg.conn.IsClosed() {
				s.onMessage(msg.conn, msg.Data)
			}

		case connID := <-s.closeChan:
			if s.onClose != nil {
				s.onClose(connID)
			}
		}
	}
}

func (s *WsServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	for _, conn := range s.conns {
		conn.Close()
	}
}

func (s *WsServer) SetOnConnect(callback func(conn *WsConn)) {
	s.onConnect = callback
}

func (s *WsServer) SetOnMessage(callback func(conn *WsConn, msg []byte)) {
	s.onMessage = callback
}

func (s *WsServer) SetOnClose(callback func(conn uint64)) {
	s.onClose = callback
}

func (s *WsServer) SendData(connID uint64, msgType int, data []byte) {
	s.rwLock.RLock()
	conn, ok := s.conns[connID]
	s.rwLock.RUnlock()

	if !ok || conn.IsClosed() {
		return
	}

	msg := &SendMessage{
		MsgType: msgType,
		Data:    data,
	}

	// ⭐ 非阻塞写（核心）
	select {
	case conn.writeChan <- msg:
	default:
		logger.Error("connect %d writeChan full, kick", connID)
		conn.Close()
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
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
	remoteAddr := conn.RemoteAddr().String()

	s.rwLock.Lock()
	s.IdCount++
	connId = s.IdCount

	wsConn := &WsConn{
		ConnId:    connId,
		conn:      conn,
		Addr:      remoteAddr,
		writeChan: make(chan *SendMessage, 100),
	}

	s.conns[connId] = wsConn
	s.rwLock.Unlock()

	logger.Info("client connected: %d:%s", connId, remoteAddr)

	// ⭐ 启动写协程（关键）
	go wsConn.writeLoop()

	s.connChan <- wsConn

	defer func() {
		logger.Info("client closed: %d:%s", connId, remoteAddr)

		wsConn.Close()

		s.rwLock.Lock()
		delete(s.conns, connId)
		s.rwLock.Unlock()

		s.closeChan <- connId
	}()

	for {
		if wsConn.IsClosed() {
			return
		}

		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway) {
				logger.Info("client closed: %s", remoteAddr)
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				logger.Info("client closed: %s", remoteAddr)
			} else {
				logger.Error("client read error: %v", err)
			}
			return
		}

		s.recvChan <- &RecvMessage{
			conn:    wsConn,
			MsgType: msgType,
			Data:    msg,
		}
	}
}

type WsConn struct {
	ConnId uint64
	conn   *websocket.Conn
	Addr   string
	Token  string

	close     bool
	writeChan chan *SendMessage
}

func (ws *WsConn) writeLoop() {
	for msg := range ws.writeChan {
		ws.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		err := ws.conn.WriteMessage(msg.MsgType, msg.Data)
		if err != nil {
			logger.Error("connect %d write error: %v", ws.ConnId, err)
			ws.Close()
			return
		}
	}
}

func (ws *WsConn) Close() {
	if ws.close {
		return
	}

	ws.close = true
	logger.Info("主动关闭连接: %d:%s", ws.ConnId, ws.Addr)

	close(ws.writeChan) // ⭐ 关闭写队列
	ws.conn.Close()
}

func (ws *WsConn) IsClosed() bool {
	return ws.close
}
