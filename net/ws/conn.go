package ws

import (
	"encoding/json"
	"go-base/logger"

	"github.com/gorilla/websocket"
)

type WsConn struct {
	ConnId   uint64
	conn     *websocket.Conn
	Addr     string
	Token    string
	sendChan chan *SendMessage
}

func (ws *WsConn) Close() {
	if ws.conn != nil {
		ws.conn.Close()
	}
}

func (ws *WsConn) SendData(data []byte) {
	ws.sendChan <- &SendMessage{ConnId: ws.ConnId, Data: data}
}

func (ws *WsConn) SendJson(data interface{}) {
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
