package tcp

import (
	"encoding/binary"
	"net"
)

type TcpClient struct {
	conn net.Conn
}

func NewTcpClient() *TcpClient {
	return &TcpClient{}
}

func (c *TcpClient) Connect(ip string, port string) error {
	address := net.JoinHostPort(ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil

}

func (c *TcpClient) SendMessage(msg string) {
	data := []byte(msg)
	length := uint32(len(data))
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, length)

	c.conn.Write(lenBuf) // 发送长度前缀
	c.conn.Write(data)
}
