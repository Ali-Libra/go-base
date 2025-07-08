package ws

import (
	"encoding/json"
	"go-base/logger"

	"github.com/gorilla/websocket"
)

type WsConn struct {
	ConnId   uint64
	sendChan chan *SendMessage
}

func (conn *WsConn) Close() {
	logger.Info("conn close: %d", conn.ConnId)
}

func (conn *WsConn) SendData(data []byte) {
	conn.sendChan <- &SendMessage{ConnId: conn.ConnId, Data: data}
}

func (conn *WsConn) SendJson(data interface{}) {
	sendData, err := json.Marshal(data)
	if err != nil {
		logger.Error("json.Marshal error")
		return
	}
	conn.sendChan <- &SendMessage{
		ConnId:  conn.ConnId,
		MsgType: websocket.TextMessage,
		Data:    sendData,
	}
}
