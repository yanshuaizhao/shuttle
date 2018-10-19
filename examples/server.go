package main

import (
	"github.com/yanshuaizhao/shuttle"
	"./callback"
)

func main() {
	srv := shuttle.NewTCPServer("localhost:9001", &callback.Callback{}, &shuttle.Packet{})
	srv.SetReadDeadline(100)
	srv.SetWriteDeadline(100)
	srv.ListenAndServe()
}
