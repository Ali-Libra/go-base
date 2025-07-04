package ws

type WsConn struct {
	connId   uint64
	sendChan chan *SendMessage
}

func (conn *WsConn) SendData(data []byte) {
	conn.sendChan <- &SendMessage{ConnId: conn.connId, Data: data}
}
