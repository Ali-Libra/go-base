package tcp

import (
	"encoding/binary"
	"net"
	"time"
)

type TcpClient struct {
	recvChan  chan []byte   // 用于接收消息的通道
	sendChan  chan []byte   // 用于发送消息的通道
	closeChan chan struct{} // 用于关闭连接的通道

	onConnect func()       // 连接成功的回调函数
	onMessage func([]byte) // 接收到消息的回调函数
	onClose   func()       // 连接关闭的回调函数
}

func NewTcpClient() *TcpClient {
	return &TcpClient{
		recvChan:  make(chan []byte, 10240), // 初始化消息通道
		sendChan:  make(chan []byte, 10240), // 初始化发送通道
		closeChan: make(chan struct{}),      // 初始化关闭通道
	}
}

func (c *TcpClient) Connect(ip string, port string) error {
	address := net.JoinHostPort(ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	if c.onConnect != nil {
		c.onConnect() // 调用连接成功的回调函数
	}
	go c.loop(conn) // 启动读取循环
	return nil
}

func (c *TcpClient) SendMessage(data []byte) {
	c.sendChan <- data
}

func (c *TcpClient) OnLoop() {
	for {
		select {
		case data := <-c.recvChan: // 从接收通道读取数据
			if c.onMessage != nil {
				c.onMessage(data) // 调用接收消息的回调函数
			}
		messageLoop:
			for {
				select {
				case data = <-c.recvChan: // 继续读取接收通道中的数据
					if c.onMessage != nil {
						c.onMessage(data) // 调用接收消息的回调函数
					}
				default:
					break messageLoop // 如果没有更多数据，退出循环
				}
			}
		case <-c.closeChan: // 监听关闭信号
			if c.onClose != nil {
				c.onClose() // 调用连接关闭的回调函数
			}
			c.recvChan = make(chan []byte, 10240) // 重置接收通道
			c.sendChan = make(chan []byte, 10240) // 重置发送通道
			return                                // 退出循环
		default:
			return
		}
	}
}

func (c *TcpClient) SetOnConnect(callback func()) {
	c.onConnect = callback
}

func (c *TcpClient) SetOnMessage(callback func([]byte)) {
	c.onMessage = callback
}

func (c *TcpClient) SetOnClose(callback func()) {
	c.onClose = callback
}

func (c *TcpClient) loop(conn net.Conn) {
	defer func() {
		conn.Close()
		c.closeChan <- struct{}{} // 关闭连接时发送信号
	}()

	readLenBuf := make([]byte, 4)
	var readLenTotal uint32 = 0
	var readData []byte
	var readDataTotal uint32 = 0
	for {
	writeLoop:
		for {
			select {
			case data := <-c.sendChan: // 从发送通道读取数据
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
		c.recvChan <- readData // 将接收到的数据发送到通道
	}
}
