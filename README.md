# Shuttle
## 简单的 GoLang TCP Socket 框架

## 安装使用

```bash
go get github.com/yanshuaizhao/shuttle

```

## 启动一个TCPServer
### tcp_server.go
```golang
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

```

### callback/callback.go
```golang
package callback

import (
	"github.com/yanshuaizhao/shuttle"
	"fmt"
	"encoding/json"
)

type Callback struct{}

func (c *Callback) OnMessage(conn *shuttle.TCPConn, p []byte) {
	msg := shuttle.ReadPacket(p)

	switch msg.Head.Cmd {
	case 0:
		// 心跳包 ping <- pong
		conn.WritePacket(shuttle.NewPacket(0, 0, msg.Head.Index, 0, `{"code":0,"data":[],"msg":"pong"}`).Packet())
		break
	default:
		// 业务逻辑处理
		var obj interface{}
		err := json.Unmarshal([]byte(msg.Data), &obj)
		if err != nil {
			return
		}
		data := obj.(map[string]interface{})
		addr := data["addr"].(string)
		if data["code"].(float64) == float64(1) && addr != "" {
			cc := conn.Bucket.Get(addr)
			if cc != nil {
				cc.WritePacket(shuttle.NewPacket(1, 1, 0, 0, `{"code":0,"data":[],"msg":"服务器转发-有人给你发单聊信息"}`).Packet())
			}
		}

		fmt.Println("客户端data:", data)
		conn.WritePacket(shuttle.NewPacket(msg.Head.Cmd, msg.Head.Act, msg.Head.Index, 0, `{"code":0,"data":[],"msg":"消息发送成功"}`).Packet())

		break
	}
}

func (c *Callback) OnConnected(conn *shuttle.TCPConn) {
	fmt.Println("new conn:", conn.GetRemoteAddr().String())
	fmt.Println("在线连接数", len(conn.Bucket.GetAll())+1)
}

func (c *Callback) OnDisconnected(conn *shuttle.TCPConn) {
	fmt.Printf("%s disconnected\n", conn.GetRemoteIPAddress())
}

func (c *Callback) OnError(err error) {
	fmt.Println(err)
}

```

## 疑问讨论 QQ：1085951557
