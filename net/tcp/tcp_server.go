package tcp

import (
	"encoding/binary"
	"go-base/logger"
	"net"
	"sync"
	"time"
)

type RecvMessage struct {
	ConnId uint64
	Data   []byte
}

type SendMessage struct {
	ConnId uint64
	Data   []byte
}

type TcpServer struct {
	port    string
	IdCount uint64

	rwLock     sync.RWMutex
	conns      map[uint64]net.Conn
	recvChan   chan *RecvMessage // 用于接收消息的通道
	sendChan   chan *SendMessage // 用于发送消息的通道
	closeChan  chan uint64       // 用于关闭连接的通道
	closeRead  chan struct{}
	closeWrite chan struct{}

	onConnect func(conn uint64)
	onMessage func(conn uint64, msg []byte)
	onClose   func(conn uint64)
}

func NewTcpServer(port string) *TcpServer {
	return &TcpServer{
		port:       port,
		recvChan:   make(chan *RecvMessage, 10240), // 初始化消息通道
		sendChan:   make(chan *SendMessage, 10240), // 初始化发送通道映射
		closeChan:  make(chan uint64),              // 初始化关闭通道
		closeRead:  make(chan struct{}),
		closeWrite: make(chan struct{}),
		IdCount:    0,
	}
}

func (c *TcpServer) SetOnConnect(callback func(conn uint64)) {
	c.onConnect = callback
}

func (c *TcpServer) SetOnMessage(callback func(conn uint64, msg []byte)) {
	c.onMessage = callback
}

func (c *TcpServer) SetOnClose(callback func(conn uint64)) {
	c.onClose = callback
}

func (s *TcpServer) Server() {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		panic(err)
	}
	go s.handleWrite()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Accept connect error %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *TcpServer) OnLoop() {
	for {
		select {
		case msg := <-s.recvChan: // 从接收通道读取数据
			if s.onMessage != nil {
				s.onMessage(msg.ConnId, msg.Data) // 调用接收消息的回调函数
			}
			for {
				select {
				case msg = <-s.recvChan: // 继续读取接收通道中的数据
					if s.onMessage != nil {
						s.onMessage(msg.ConnId, msg.Data) // 调用接收消息的回调函数
					}
				default:
					return
				}
			}
		case conn := <-s.closeChan: // 监听关闭信号
			if s.onClose != nil {
				s.onClose(conn) // 调用连接关闭的回调函数
			}
		default:
			return
		}
	}
}

func (s *TcpServer) Close() {
	//先关闭接受通道，不再接收客户端消息
	s.closeRead <- struct{}{}

	for {
		if len(s.sendChan) == 0 {
			s.closeWrite <- struct{}{}
			for connId, conn := range s.conns {
				conn.Close()
				delete(s.conns, connId)
			}
		}
	}
}

func (s *TcpServer) SendMessage(connId uint64, data []byte) {
	s.sendChan <- &SendMessage{ConnId: connId, Data: data} // 将数据发送到对应连接的发送通道
}

func (s *TcpServer) handleConn(conn net.Conn) {
	var connId uint64
	defer func() {
		conn.Close()
		s.rwLock.Lock()
		delete(s.conns, connId)
		s.rwLock.Unlock()
		s.closeChan <- connId
	}()

	s.rwLock.Lock()
	s.IdCount++
	connId = s.IdCount
	s.conns[connId] = conn
	s.rwLock.Unlock()

	readLenBuf := make([]byte, 4)
	var readLenTotal uint32 = 0
	var readData []byte
	var readDataTotal uint32 = 0
	for {
		select {
		case <-s.closeRead:
			close(s.recvChan)
		default:
			if readLenTotal < 4 {
				deadline := time.Now().Add(10 * time.Millisecond)
				conn.SetReadDeadline(deadline)
				n, err := conn.Read(readLenBuf[readLenTotal:]) // 读取长度前缀
				if err != nil {
					if ne, ok := err.(net.Error); ok && !ne.Timeout() {
						logger.Error("connect %d read msg  error %v", connId, err)
						return // 读取失败，退出循环
					}
				}
				readLenTotal += uint32(n)
			}
			if readLenTotal < 4 {
				continue
			}

			length := binary.BigEndian.Uint32(readLenBuf)
			if readDataTotal == 0 {
				if length == 0 || length > 10*1024*1024 {
					logger.Error("connect %d read msg len error", connId)
					return // 长度不合法，退出循环
				}
				readData = make([]byte, length) // 分配足够的空间来存储数据
			}
			if readDataTotal < length {
				deadline := time.Now().Add(10 * time.Millisecond)
				conn.SetReadDeadline(deadline)
				n, err := conn.Read(readLenBuf[readDataTotal:]) // 读取长度前缀
				if err != nil {
					if ne, ok := err.(net.Error); ok && !ne.Timeout() {
						logger.Error("connect %d read msg  error %v", connId, err)
						return // 读取失败，退出循环
					}
				}
				readDataTotal += uint32(n)
			}
			if readDataTotal < length {
				continue
			}

			readLenTotal = 0
			readDataTotal = 0

			// 4. 处理消息
			s.recvChan <- &RecvMessage{
				ConnId: connId, // 将连接和消息数据封装到 ConnMessage 中
				Data:   readData,
			}
		}
	}
}

func (s *TcpServer) handleWrite() {
	for {
		select {
		case <-s.closeWrite:
			return
		default:
			for msg := range s.sendChan { // 从发送通道读取数据
				length := uint32(len(msg.Data))
				lenBuf := make([]byte, 4)
				binary.BigEndian.PutUint32(lenBuf, length)

				s.rwLock.RLock()
				conn, ok := s.conns[msg.ConnId]
				s.rwLock.RUnlock()
				if !ok {
					logger.Error("connect %d have  closed", msg.ConnId)
					continue
				}

				_, err := conn.Write(lenBuf) // 发送长度前缀
				if err != nil {
					logger.Error("connect %d have  write error", msg.ConnId)
					continue
				}
				_, err = conn.Write(msg.Data)
				if err != nil {
					logger.Error("connect %d have  write error", msg.ConnId)
					continue
				}
			}
		}
	}
}
