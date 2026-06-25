package timer

import (
	"cake/internal/util/sys"
	"github.com/puzpuzpuz/xsync/v3"
	"sync"
	"sync/atomic"
	"time"
)

type IGenService interface {
	TickerQueue() chan struct{}
	PID() uint64
	Name() string
	State() any
	Copy(string) (any, bool)
	Restore(any)
}

type timer struct {
	id           int
	tickInterval time.Duration
	services     *xsync.MapOf[uint64, IGenService] // 所有需要接收tick的玩家服务
	mu           sync.RWMutex
	ticker       *time.Ticker
	closeCh      chan struct{}
	closed       atomic.Bool
}

func startTimer(id int) *timer {
	t := &timer{
		id:           id,
		tickInterval: TickInterval,
		services:     xsync.NewMapOf[uint64, IGenService](),
		closeCh:      make(chan struct{}),
	}
	t.ticker = time.NewTicker(TickInterval)
	sys.SafeGo(t.loop)
	return t
}

func (t *timer) loop() {
	for {
		select {
		case <-t.ticker.C:
			if !t.closed.Load() {
				t.bcastTick()
			}
		case <-t.closeCh:
			t.ticker.Stop()
			return
		}
	}
}

// bcastTick 向所有注册的GenServer广播Tick消息
func (t *timer) bcastTick() {
	// 遍历所有在线玩家，逐条发送节拍消息
	t.services.Range(func(key uint64, svc IGenService) bool {
		ch := svc.TickerQueue()
		// 队列满，丢弃防止阻塞全局tick
		sys.SafeSend(ch, struct{}{})
		return true // 返回 true 继续遍历
	})

}
