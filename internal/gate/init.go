package gate

import (
	"cake/env"
	"cake/internal/gate/conn"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"net"
)

func Init() {
	sys.SafeGo(func() {
		// 监听 TCP 端口
		addr := env.GetString("gate.addr")
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		defer listener.Close()
		logger.Info("TCP 服务已启动：127.0.0.1:8888 ✅")
		// 循环接收客户端连接
		conn.Loop(listener)
	})
}

func Stop() {
	conn.StopAccept()
	conn.StopManager()
}
