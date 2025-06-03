package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type TcpServer struct {
	port      string
	onconnect func(conn net.Conn)
	onmessage func(conn net.Conn, msg []byte)
	onclose   func(conn net.Conn)
}

func NewTcpServer(port string,
	onconnect func(conn net.Conn),
	onmessage func(conn net.Conn, msg []byte),
	onclose func(conn net.Conn)) *TcpServer {
	return &TcpServer{
		port:      port,
		onconnect: onconnect,
		onmessage: onmessage,
		onclose:   onclose,
	}
}

func (s *TcpServer) Run() {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *TcpServer) Close() {

}

func (s *TcpServer) handleConn(conn net.Conn) {
	defer func() {
		if s.onclose != nil {
			s.onclose(conn)
		}
		conn.Close()
	}()

	if s.onconnect != nil {
		s.onconnect(conn)
	}

	for {
		// 1. 读取 4 字节长度前缀
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
			} else {
				fmt.Println("Read length error:", err)
			}
			return
		}

		// 2. 解析长度
		msgLen := binary.BigEndian.Uint32(lenBuf)
		if msgLen == 0 || msgLen > 10*1024*1024 {
			fmt.Println("Invalid message length:", msgLen)
			return
		}

		// 3. 读取消息体
		msgBuf := make([]byte, msgLen)
		_, err = io.ReadFull(conn, msgBuf)
		if err != nil {
			fmt.Println("Read body error:", err)
			return
		}

		// 4. 处理消息
		if s.onmessage != nil {
			s.onmessage(conn, msgBuf)
		}
	}
}
