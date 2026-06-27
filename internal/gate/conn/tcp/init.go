package tcp

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"net"
)

var isStartTcp bool

func Init() {
	addr := env.GetString("gate.tcpAddr")
	if addr == "" {
		logger.Infof("Success	TCP 服务未启动")
		return
	}
	sys.SafeGo(func() {
		isStartTcp = true
		// 监听 TCP 端口
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
	if isStartTcp {
		acceptInst.stop()
	}
}
