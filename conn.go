package shuttle

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type TCPConn struct {
	Bucket *TCPConnBucket

	callback CallBack
	protocol Protocol

	conn      *net.TCPConn
	readChan  chan []byte
	writeChan chan []byte

	readDeadline  time.Duration
	writeDeadline time.Duration

	closeOnce sync.Once
	exitFlag  int32
}

func (sev *TCPServer) NewTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConn {
	c := &TCPConn{
		conn:     conn,
		callback: callback,
		protocol: protocol,

		readChan:  make(chan []byte, 100),
		writeChan: make(chan []byte, 100),

		Bucket:   sev.bucket,
		exitFlag: 0,
	}
	return c
}

func (c *TCPConn) Connection() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(fmt.Sprintf("tcp conn(%v) server error: %v ", c.GetRemoteIPAddress(), r))
		}
	}()
	if c.callback == nil || c.protocol == nil {
		err := fmt.Errorf("callback and protocol are not allowed to be nil")
		c.Close()
		return err
	}
	atomic.StoreInt32(&c.exitFlag, 1)
	c.callback.OnConnected(c)
	go c.handleConn()
	go c.readLoop()
	go c.writeLoop()
	return nil
}

func (c *TCPConn) handleConn() {
	if c.readDeadline > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readDeadline))
	}

	_, err := c.protocol.ReadPacket(c.conn, c.readChan)
	if err != nil {
		fmt.Println("l: ", err)
		/*if err == io.EOF {
		} else {
		}*/
		c.callback.OnError(err)
		c.Close()
		return
	}
}

func (c *TCPConn) readLoop() {
	for p := range c.readChan {
		if p == nil {
			continue
		}
		c.callback.OnMessage(c, p)
	}
}

func (c *TCPConn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for pkt := range c.writeChan {
		if pkt == nil {
			continue
		}
		if c.writeDeadline > 0 {
			c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline))
		}
		if err := c.protocol.WritePacket(c.conn, pkt); err != nil {
			c.callback.OnError(err)
			return
		}
	}
}

func (c *TCPConn) WritePacket(p []byte) error {
	if c.IsClosed() {
		return errors.New("该连接已关闭")
	}
	select {
	case c.writeChan <- p:
		return nil
	default:
		return errors.New("异步发送缓冲区已满")
	}
}

func (c *TCPConn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.exitFlag, 0)
		close(c.writeChan)
		close(c.readChan)
		c.Bucket.removeClosedTCPConn()
		if c.callback != nil {
			c.callback.OnDisconnected(c)
		}
		c.conn.Close()
	})
}

func (c *TCPConn) GetRawConn() *net.TCPConn {
	return c.conn
}

func (c *TCPConn) setReadDeadline(t time.Duration) {
	c.readDeadline = t
}

func (c *TCPConn) setWriteDeadline(t time.Duration) {
	c.writeDeadline = t
}

func (c *TCPConn) IsClosed() bool {
	return atomic.LoadInt32(&c.exitFlag) == 0
}

func (c *TCPConn) GetLocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TCPConn) GetLocalIPAddress() string {
	return strings.Split(c.GetLocalAddr().String(), ":")[0]
}

func (c *TCPConn) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConn) GetRemoteIPAddress() string {
	return strings.Split(c.GetRemoteAddr().String(), ":")[0]
}
