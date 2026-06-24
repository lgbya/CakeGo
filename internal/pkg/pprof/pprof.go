package pprof

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"errors"
	"log"
	"net/http"
	"net/http/pprof"
)

var pprofSrv *http.Server

func Init() {
	// 单独协程启动调试服务，建议只监听127.0.0.1
	addr := env.GetString("monitoring.pprofAddr")
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline) //获取当前进程启动命令
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile) //CPU 性能采样（最常用）
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)   //程序函数名、内存地址映射关系。
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)     //程序运行追踪（高级全链路采样）
	pprofSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	sys.SafeGo(func() {
		logger.Infof("Success	pprof")
		if err := pprofSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Fail	pprof listen err: %v", err)
		}
	})
}

//func Close() error {
//	if pprofSrv == nil {
//		return nil
//	}
//	// 设置5秒超时，强制关闭残留连接
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	return pprofSrv.Shutdown(ctx)
//}
