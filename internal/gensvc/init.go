package gensvc

import (
	"cake/internal/gensvc/rpc"
	"cake/internal/gensvc/rpcgen"
	"cake/internal/gensvc/timer"
	"cake/internal/pkg/metric"
)

func Init() {
	metric.Init()
	rpcgen.Init()
	timer.Init()
}

func Stop() {
	timer.Stop()
	rpc.StopAll()
}
