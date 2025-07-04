package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WsConn struct {
	*websocket.Conn
}

func (conn *WsConn) RspJson(reply interface{}) {
	replyBytes, _ := json.Marshal(reply)
	conn.WriteMessage(websocket.TextMessage, replyBytes)
}
