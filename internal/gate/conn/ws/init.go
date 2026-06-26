package ws

import (
	"cake/internal/pkg/logger"
	"net/http"
)

func Init() {
	http.HandleFunc("/ws", wsHandler)
	logger.Infof("ws服务启动 :8080")
	_ = http.ListenAndServe("0.0.0.0:8888", nil)
}
