package game

import (
	"cake/env"
	"cake/internal/game/logic/sdk"
	"cake/internal/game/services/mapsvc"
	"cake/internal/game/services/rolesvc"
	"cake/internal/gate"
	"cake/internal/gate/router"
	"cake/internal/gensvc"
	"cake/internal/pkg/db"
	"cake/internal/pkg/logger"
	"cake/internal/pkg/pprof"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var initFn = []func(){
	env.Init,
	logger.Init,
	pprof.Init,
	db.Init,
	gensvc.Init,
	mapsvc.Init,
	sdk.Init,
	router.Init,
	gate.Init,
}

var stopFn = []func(){
	gate.Stop,
	mapsvc.Stop,
	rolesvc.Stop,
	gensvc.Stop,
	db.Stop,
}

func Init() {
	for _, fn := range initFn {
		fn()
	}
	//service, err := testsvc.StartService()
	//if err != nil {
	//	return
	//}
	//service.Send5s("RpcTest", 1)
	//service.Send5s("RpcTest", 2)
	//service.Send5s("RpcTest", 3)
	//service.Send5s("RpcTest", 4)
}

func Stop() {
	defer logger.Sync()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	sig := <-quit
	logger.Errorf("收到终止信号: %s, 服务器关闭中", sig.String())
	for _, fn := range stopFn {
		fn()
	}
	time.Sleep(time.Second)
	logger.Errorf("服务器关闭完成")

}

//func Stop() {
//	for _, fn := range stopFn {
//		fn()
//	}
//}
