package tcp

import (
	"cake/internal/gate/conn/connsvc"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"context"
	"errors"
	"net"
	"runtime"
	"sync"
	"time"
)

var acceptInst = newAccept()

type accept struct {
	wg     sync.WaitGroup
	cxt    context.Context
	cancel context.CancelFunc
	ln     net.Listener
}

func newAccept() *accept {
	ctx, cancel := context.WithCancel(context.Background())
	return &accept{cxt: ctx, cancel: cancel}
}

func (a *accept) loop(ln net.Listener) {
	acceptGoNum := runtime.NumCPU()
	a.ln = ln
	a.wg.Add(acceptGoNum)
	for i := 0; i < acceptGoNum; i++ {
		sys.SafeGo(func() {
			defer a.wg.Done()
			for {

				select {
				case <-a.cxt.Done():
					logger.Infof("收到停止信号，acceptInst 协程退出")
					return
				default:
				}

				conn, err := ln.Accept()
				if err != nil {
					// 监听被正常关闭，退出循环
					if errors.Is(err, net.ErrClosed) {
						logger.Errorf("监听已关闭，停止接收连接 : %v", err)
						return
					}
					// 其他临时错误，打印日志后重试
					logger.Errorf("acceptInst 错误（非致命，重试）: %v", err)
					time.Sleep(10 * time.Millisecond) // 避免错误风暴
					continue
				}
				//fmt.Println("新客户端接入:", conn.RemoteAddr())
				tpcConn := NewTcpConn(conn)
				connsvc.StartService(tpcConn)
			}
		})
	}
	a.wg.Wait()
}

func (a *accept) stop() {
	a.cancel()
	err := a.ln.Close()
	if err != nil {
		logger.Errorf("关闭监听失败: %v", err)
	}
	a.wg.Wait()
	logger.Errorf("所有Accept工作协程已全部退出")
}
