package ws

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Ali-Libra/go-base/logger"
	"github.com/gorilla/websocket"
)

type WsConn struct {
	conn      *websocket.Conn
	ConnId    uint64
	Addr      string
	Token     string
	close     bool
	writeChan chan *SendReq
}

func (ws *WsConn) writeLoop() {
	for msg := range ws.writeChan {
		ws.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		err := ws.conn.WriteMessage(msg.MsgType, msg.Data)
		if err != nil {
			logger.Error("connect %s write error: %v", ws.Addr, err)
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

	logger.Info("主动关闭连接: %s", ws.Addr)

	close(ws.writeChan)
	ws.conn.Close()
}

func (ws *WsConn) IsClosed() bool {
	return ws.close
}

type SendReq struct {
	ConnId  uint64
	MsgType int
	Data    []byte
}

type WsServer struct {
	server *http.Server
	mux    *http.ServeMux

	connId     uint64
	conns      map[uint64]*WsConn
	addConn    chan *WsConn
	removeConn chan *WsConn
	sendChan   chan *SendReq

	onConnect func(conn *WsConn)
	onMessage func(conn *WsConn, msg []byte)
	onClose   func(conn *WsConn)
}

func NewWsServer() *WsServer {
	s := &WsServer{
		mux:        http.NewServeMux(),
		conns:      make(map[uint64]*WsConn),
		addConn:    make(chan *WsConn, 1000),
		removeConn: make(chan *WsConn, 1000),
		sendChan:   make(chan *SendReq, 10000),
	}

	go s.loop()
	return s
}

func (s *WsServer) loop() {
	for {
		select {
		case conn := <-s.addConn:
			s.connId++
			conn.ConnId = s.connId
			s.conns[conn.ConnId] = conn
			if s.onConnect != nil {
				s.onConnect(conn)
			}

		case conn := <-s.removeConn:
			if _, ok := s.conns[conn.ConnId]; ok {
				delete(s.conns, conn.ConnId)
				if s.onClose != nil {
					s.onClose(conn)
				}
			}

		case req := <-s.sendChan:
			if _, ok := s.conns[req.ConnId]; ok && !s.conns[req.ConnId].IsClosed() {
				select {
				case s.conns[req.ConnId].writeChan <- req:
				default:
					logger.Error("connect %s writeChan full, kick", s.conns[req.ConnId].Addr)
					s.conns[req.ConnId].Close()
				}
			}
		}
	}
}

func (s *WsServer) Run(port string, path string) {
	s.server = &http.Server{
		Addr:    port,
		Handler: s.mux,
	}

	s.mux.HandleFunc("/"+path, s.wsHandler)

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			logger.Error("ListenAndServe error: %v", err)
		}
	}()
}

func (s *WsServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

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

func (s *WsServer) SetOnClose(callback func(conn *WsConn)) {
	s.onClose = callback
}

func (s *WsServer) SendData(connID uint64, msgType int, data []byte) {
	s.sendChan <- &SendReq{
		ConnId:  connID,
		MsgType: msgType,
		Data:    data,
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

	remoteAddr := conn.RemoteAddr().String()

	wsConn := &WsConn{
		conn:      conn,
		Addr:      remoteAddr,
		writeChan: make(chan *SendReq, 100),
	}

	logger.Info("client connected: %s", remoteAddr)

	s.addConn <- wsConn
	go wsConn.writeLoop()

	defer func() {
		logger.Info("client closed: %s", remoteAddr)

		wsConn.Close()

		// 删除连接（走事件循环）
		s.removeConn <- wsConn
	}()

	for {
		if wsConn.IsClosed() {
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway) ||
				strings.Contains(err.Error(), "use of closed network connection") {

				logger.Info("client closed: %s", remoteAddr)
			} else {
				logger.Error("client read error: %v", err)
			}
			return
		}

		if s.onMessage != nil {
			s.onMessage(wsConn, msg)
		}
	}
}
