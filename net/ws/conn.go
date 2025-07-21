package ws

import (
	"encoding/json"

	"github.com/Ali-Libra/go-base/logger"
	"github.com/gorilla/websocket"
)

type WsConn struct {
	ConnId   uint64
	conn     *websocket.Conn
	Addr     string
	Token    string
	close    bool
	sendChan chan *SendMessage
}

func (ws *WsConn) Close() {
	logger.Info("主动关闭连接: %d:%s", ws.ConnId, ws.Addr)
	if ws.conn != nil {
		ws.close = true
		ws.conn.Close()
	}
}

func (ws *WsConn) IsClosed() bool {
	return ws.close
}

func (ws *WsConn) SendData(data []byte) {
	if ws.IsClosed() {
		return
	}
	ws.sendChan <- &SendMessage{ConnId: ws.ConnId, Data: data}
}

func (ws *WsConn) SendJson(data interface{}) {
	if ws.IsClosed() {
		return
	}
	sendData, err := json.Marshal(data)
	if err != nil {
		logger.Error("json.Marshal error")
		return
	}
	ws.sendChan <- &SendMessage{
		ConnId:  ws.ConnId,
		MsgType: websocket.TextMessage,
		Data:    sendData,
	}
}
