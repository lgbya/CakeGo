package rpc

import (
	"context"
	"sync"
	"time"
)

type CbChan chan any

type Msg struct {
	Cmd     string // 命令字
	Args    any    // 参数: 绑定具体结构体
	Delay   time.Duration
	CbChat  CbChan
	CallCtx context.Context
	trace   string
}

// 重置函数必须把所有引用类型清空
func (m *Msg) Reset() {
	m.Cmd = ""
	m.Args = nil // any 指针必须置 nil，防止内存泄漏、旧数据残留
	m.Delay = 0
	m.CbChat = nil  // 回调chan置空
	m.CallCtx = nil // 上下文一定要释放，否则会一直持有父上下文
	m.trace = ""
}

type Cfg struct {
	InitArgs     any
	Ctx          context.Context
	isInit       bool
	StartTimeout time.Duration
	TimerTicker  time.Duration
	SendFn       func(*Msg) error
	Wg           *sync.WaitGroup
	SendMaxCap   int
	IsCopy       bool
}

func NewCfg() Cfg {
	return Cfg{
		InitArgs:     nil,
		StartTimeout: 30 * time.Second,
		//Ctx:          mgr.ctx,
		isInit:      false,
		TimerTicker: 100 * time.Millisecond,
		SendFn:      nil,
		SendMaxCap:  1024,
	}
}
