package ws

type WsConn struct {
	ConnId   uint64
	sendChan chan *SendMessage
}

func (conn *WsConn) SendData(data []byte) {
	conn.sendChan <- &SendMessage{ConnId: conn.ConnId, Data: data}
}
