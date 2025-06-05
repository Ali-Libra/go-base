package tcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type ConnMessage struct {
	Conn net.Conn
	Data []byte
}

type TcpServer struct {
	port      string
	recvChan  chan ConnMessage         // 用于接收消息的通道
	sendChans map[net.Conn]chan []byte // 用于发送消息的通道
	closeChan chan net.Conn            // 用于关闭连接的通道

	onConnect func(conn net.Conn)
	onMessage func(conn net.Conn, msg []byte)
	onClose   func(conn net.Conn)
}

func NewTcpServer(port string) *TcpServer {
	return &TcpServer{
		port:      port,
		recvChan:  make(chan ConnMessage, 10240),  // 初始化消息通道
		sendChans: make(map[net.Conn]chan []byte), // 初始化发送通道映射
		closeChan: make(chan net.Conn),            // 初始化关闭通道
	}
}

func (c *TcpServer) SetOnConnect(callback func(conn net.Conn)) {
	c.onConnect = callback
}

func (c *TcpServer) SetOnMessage(callback func(conn net.Conn, msg []byte)) {
	c.onMessage = callback
}

func (c *TcpServer) SetOnClose(callback func(conn net.Conn)) {
	c.onClose = callback
}

func (s *TcpServer) Server() {
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

func (s *TcpServer) SendMessage(conn net.Conn, data []byte) {
	if _, ok := s.sendChans[conn]; !ok {
		return
	}
	s.sendChans[conn] <- data // 将数据发送到对应连接的发送通道
}

func (s *TcpServer) OnLoop() {
	for {
		select {
		case msg := <-s.recvChan: // 从接收通道读取数据
			if _, ok := s.sendChans[msg.Conn]; !ok {
				s.sendChans[msg.Conn] = make(chan []byte, 10240)
			}
			if s.onMessage != nil {
				s.onMessage(msg.Conn, msg.Data) // 调用接收消息的回调函数
			}
		messageLoop:
			for {
				select {
				case msg = <-s.recvChan: // 继续读取接收通道中的数据
					if s.onMessage != nil {
						s.onMessage(msg.Conn, msg.Data) // 调用接收消息的回调函数
					}
				default:
					break messageLoop // 如果没有更多数据，退出循环
				}
			}
		case conn := <-s.closeChan: // 监听关闭信号
			if s.onClose != nil {
				s.onClose(conn) // 调用连接关闭的回调函数
			}
			delete(s.sendChans, conn) // 从发送通道映射中删除连接
		default:
			return
		}
	}
}

func (s *TcpServer) Close() {

}

func (s *TcpServer) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.closeChan <- conn
	}()

	readLenBuf := make([]byte, 4)
	var readLenTotal uint32 = 0
	var readData []byte
	var readDataTotal uint32 = 0
	for {
	writeLoop:
		for {
			select {
			case data := <-s.sendChans[conn]: // 从发送通道读取数据
				length := uint32(len(data))
				lenBuf := make([]byte, 4)
				binary.BigEndian.PutUint32(lenBuf, length)

				_, err := conn.Write(lenBuf) // 发送长度前缀
				if err != nil {
					return
				}
				_, err = conn.Write(data)
				if err != nil {
					return
				}
			default:
				break writeLoop // 如果没有数据，退出循环
			}
		}

		if readLenTotal < 4 {
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			n, err := conn.Read(readLenBuf[readLenTotal:]) // 读取长度前缀
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // 如果是超时错误，继续循环
				}
				return // 读取失败，退出循环
			}
			readLenTotal += uint32(n)
		}
		if readLenTotal < 4 {
			continue
		}

		length := binary.BigEndian.Uint32(readLenBuf)
		if readDataTotal == 0 {
			if length == 0 || length > 10*1024*1024 {
				return // 长度不合法，退出循环
			}
			readData = make([]byte, length) // 分配足够的空间来存储数据
		}
		if readDataTotal < length {
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			n, err := conn.Read(readData[readDataTotal:]) // 读取实际数据
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // 如果是超时错误，继续循环
				}
				return // 读取失败，退出循环
			}
			readDataTotal += uint32(n)
		}
		if readDataTotal < length {
			continue
		}

		readLenTotal = 0
		readDataTotal = 0

		// 4. 处理消息
		s.recvChan <- ConnMessage{
			Conn: conn, // 将连接和消息数据封装到 ConnMessage 中
			Data: readData,
		}
	}
}
