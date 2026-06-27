package ws

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"net/http"
)

func Init() {
	http.HandleFunc("/ws", wsHandler)
	addr := env.GetString("gate.websocketAddr")
	if addr == "" {
		logger.Infof("ws服务未启动")
		return
	}
	logger.Infof("ws服务启动 :%s", addr)
	_ = http.ListenAndServe(addr, nil)
}
