package gate

import (
	"cake/internal/gate/conn/connsvc"
	"cake/internal/gate/conn/tcp"
)

func Init() {
	tcp.Init()
}

func Stop() {
	tcp.Stop()
	connsvc.StopManager()
}
