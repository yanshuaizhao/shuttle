package main

import (
	"github.com/yanshuaizhao/shuttle-go"
	"fmt"
	"net"
	"os"
	"time"
	"io"
)

func sendHeartBeat(conn net.Conn) {
	i := 0
	for {
		time.Sleep(1 * time.Second)
		conn.Write(tcp.NewPacket(0, 0, int32(i), 0, "").Packet())
		i++
	}
}

func main() {
	server := "127.0.0.1:9001"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("connect success")
	go sendHeartBeat(conn)

	//声明一个管道用于接收解包的数据
	readerChan := make(chan []byte, 10)
	go readChan(readerChan)

	p := tcp.Packet{}
	_, err2 := p.ReadPacket(conn, readerChan)
	if err2 != nil {
		fmt.Println("l: ", err2)
		if err2 == io.EOF {
			fmt.Println("连接已关闭")
		} else {
			fmt.Println("异常错误:", err2)
		}
	}
}

func readChan(readerChan chan []byte) {
	for {
		select {
		case d := <-readerChan:
			data := tcp.ReadPacket(d)
			fmt.Println("from to server:", data.Head, data.Data)
		}
	}
}
