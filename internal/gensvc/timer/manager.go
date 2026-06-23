package timer

import (
	"cake/internal/pkg/logger"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StartTimerNum = 4
	TickInterval  = 10 * time.Millisecond
)

var (
	defaultMgr *manager
	once       sync.Once
)

type manager struct {
	timers map[int]*timer
	closed atomic.Bool
}

func Init() {
	once.Do(func() {
		defaultMgr = &manager{timers: make(map[int]*timer, StartTimerNum)}
		for i := 0; i < StartTimerNum; i++ {
			t := startTimer(i)
			defaultMgr.timers[i] = t
		}
	})
}

func getTimer(svc IGenService) *timer {
	if defaultMgr == nil || defaultMgr.closed.Load() {
		return nil
	}
	timerID := int(svc.PID()) % StartTimerNum
	t, ok := defaultMgr.timers[timerID]
	if !ok {
		logger.Errorf("[%d|%s]注册定时器失败", svc.PID(), svc.Name())
		return nil
	}
	return t
}

// Register 玩家GenServer启动时注册，开始接收全局节拍
func Register(svc IGenService) {
	t := getTimer(svc)
	if t == nil {
		return
	}
	t.services.Store(svc.PID(), svc)
}

// UnRegister 玩家下线注销，不再接收节拍，防止野消息
func UnRegister(svc IGenService) {
	t := getTimer(svc)
	if t == nil {
		return
	}
	t.services.Delete(svc.PID())
}

// Stop 进程关闭时停止全局节拍
func Stop() {
	if defaultMgr == nil || defaultMgr.closed.Swap(true) {
		return
	}
	for _, t := range defaultMgr.timers {
		close(t.closeCh)
	}
}
