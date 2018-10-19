package shuttle

type CallBack interface {
	//链接建立回调
	OnConnected(conn *TCPConn)

	//消息处理回调
	OnMessage(conn *TCPConn, p []byte)

	//链接断开回调
	OnDisconnected(conn *TCPConn)

	//错误回调
	OnError(err error)
}
