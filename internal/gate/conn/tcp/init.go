package tcp

import (
	"cake/env"
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

		logger.Infof("Success	TCP 服务已启动 %s", addr)
		// 循环接收客户端连接
		acceptInst.loop(listener)
	})
}

func Stop() {
	acceptInst.stop()
}
