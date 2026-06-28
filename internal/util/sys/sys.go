package sys

import (
	"cake/internal/pkg/logger"
	"runtime/debug"
	"time"
)

type SafeCfg struct {
	Context string //上下文
}

// Recover 通用panic恢复函数，传入上下文描述（玩家ID/链路ID/模块名）
// 使用方式：defer panicx.Recover("模块名-玩家10001")
func Recover(cxt string) {
	if err := recover(); err != nil {
		logger.Errorf("[SAFE RUN PANIC] 上下文: %s | 错误: %v", cxt, err)
		logger.Errorf("[STACK TRACE]\n%s", string(debug.Stack()))
	}
}

func SafeGo(fn func()) {
	SafeGoWithCfg(fn, SafeCfg{
		Context: "default",
	})
}

func SafeGoWithCfg(fn func(), safeCfg SafeCfg) {
	go func() {
		defer Recover(safeCfg.Context)
		fn()
	}()
}

func SafeRun(fn func()) {
	SafeRunWithCfgVoid(fn, SafeCfg{
		Context: "default",
	})
}

func SafeRunT[T any](fn func() T) T {
	return SafeRunWithCfg(fn, SafeCfg{
		Context: "default",
	})
}

func SafeRunWithCfgVoid(fn func(), safeCfg SafeCfg) {
	_ = SafeRunWithCfg(func() struct{} {
		fn()
		return struct{}{}
	}, safeCfg)
}

func SafeRunWithCfg[T any](fn func() T, safeCfg SafeCfg) T {
	defer Recover(safeCfg.Context)
	return fn()
}

func SafeCloseChan[T any](ch chan T) {
	SafeCloseChanWithCfg(ch, SafeCfg{
		Context: "default",
	})
}

func SafeCloseChanWithCfg[T any](ch chan T, safeCfg SafeCfg) {
	defer Recover(safeCfg.Context)
	close(ch)
}

// 无超时
func SafeSend[T any](ch chan T, msg T) bool {
	defer Recover("SafeSend")
	select {
	case ch <- msg:
		return true
	default:
		logger.Warnf("chan已经满了, 丢失msg:%v", msg)
		return false
	}
}

// 超时
func SafeSend5s[T any](ch chan T, msg T) bool {
	return SafeSendTimeout(ch, msg, 5*time.Second)
}

func SafeSendTimeout[T any](ch chan T, msg T, timeout time.Duration) bool {
	defer Recover("SafeSendTimeout")

	if timeout > 0 {
		select {
		case ch <- msg:
			return true
		case <-time.After(timeout):
			logger.Warnf("chan已经满了, 丢失msg:%v", msg)
			return false
		}
	} else {
		ch <- msg
		return true
	}
}

// 必定输入消息
func SafeSendWait[T any](ch chan T, msg T) {
	defer Recover("SafeSendWait")
	ch <- msg
}
