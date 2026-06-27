package gate

import (
	"cake/internal/gate/conn/connsvc"
	"cake/internal/gate/conn/tcp"
	"cake/internal/gate/conn/ws"
)

func Init() {
	tcp.Init()
	ws.Init()
}

func Stop() {
	tcp.Stop()
	connsvc.StopManager()
}
