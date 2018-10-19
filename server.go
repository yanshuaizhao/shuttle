package shuttle

import (
	"errors"
	"net"
	"fmt"
	"time"
)

type TCPServer struct {
	tcpAddr  string
	listener *net.TCPListener
	bucket   *TCPConnBucket
	callback CallBack
	protocol Protocol

	readDeadline  time.Duration
	writeDeadline time.Duration
	exitChan      chan struct{}
}

func NewTCPServer(tcpAddr string, callback CallBack, protocol Protocol) *TCPServer {
	return &TCPServer{
		tcpAddr:  tcpAddr,
		bucket:   NewTCPConnBucket(),
		callback: callback,
		protocol: protocol,

		readDeadline:  180 * time.Second,
		writeDeadline: 180 * time.Second,

		exitChan: make(chan struct{}),
	}
}

func (srv *TCPServer) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", srv.tcpAddr)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}
	fmt.Println("server listen start: ", srv.tcpAddr)
	srv.serve(ln)
	return nil
}

func (srv *TCPServer) serve(l *net.TCPListener) error {
	srv.listener = l
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("TCPServe error", r)
		}
		srv.listener.Close()
	}()

	for {
		select {
		case <-srv.exitChan:
			return errors.New("TCPServer closed")
		default:
		}
		conn, err := srv.listener.AcceptTCP()
		if err != nil {
			fmt.Println("ln error:", err.Error())
			return err
		}
		tcpConn := srv.newTCPConn(conn, srv.callback, srv.protocol)
		tcpConn.setReadDeadline(srv.readDeadline)
		tcpConn.setWriteDeadline(srv.writeDeadline)
		srv.bucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (srv *TCPServer) newTCPConn(conn *net.TCPConn, callback CallBack, protocol Protocol) *TCPConn {
	if callback == nil {
		callback = srv.callback
	}

	if protocol == nil {
		protocol = srv.protocol
	}

	c := srv.NewTCPConn(conn, callback, protocol)
	c.Connection()
	return c
}

func (srv *TCPServer) Close() {
	defer srv.listener.Close()
	for _, c := range srv.bucket.GetAll() {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

func (srv *TCPServer) SetReadDeadline(n time.Duration) {
	srv.readDeadline = n * time.Second
}

func (srv *TCPServer) SetWriteDeadline(n time.Duration) {
	srv.writeDeadline = n * time.Second
}
