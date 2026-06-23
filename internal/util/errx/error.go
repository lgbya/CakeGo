package errx

import (
	"bytes"
	"cake/internal/game/def/errcode"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// BizError 统一业务错误结构体，实现 error、Unwrap 接口
type BizError struct {
	Code    uint32 `json:"code"`   // 业务错误码
	Message string `json:"msg"`    // 对外展示文案
	Detail  string `json:"detail"` // 内部详细错误（日志打印用）
	Stack   string `json:"-"`      // 堆栈信息，不返回前端
	Err     error  `json:"-"`      // 原始错误
}

func (e *BizError) Error() string {
	return fmt.Sprintf("[%d]%s | detail: %s", e.Code, e.Message, e.Detail)
}

func (e *BizError) Unwrap() error {
	return e.Err
}

// New 创建基础业务错误
func New(code uint32, args ...any) *BizError {
	var msg string
	var detail string
	switch len(args) {
	case 1:
		msg = args[0].(string)
	case 2:
		msg = args[0].(string)
		detail = args[1].(string)
	}

	return &BizError{
		Code:    code,
		Message: msg,
		Detail:  detail,
	}
}

// GetCode 获取错误码，非业务错误默认返回500
func GetCode(err error) uint32 {
	var be *BizError
	if errors.As(err, &be) {
		return be.Code
	}
	return errcode.System
}

// Wrap 包装原始错误，携带上下文+堆栈
func Wrap(code uint32, msg, detail string, err error) *BizError {
	be := New(code, msg, detail)
	be.Err = err
	be.Stack = GetStack(2) // 跳过当前Wrap函数，拿到调用方堆栈
	return be
}

// Is 判断错误是否为指定业务错误码
func Is(err error, code uint32) bool {
	var be *BizError
	if errors.As(err, &be) {
		return be.Code == code
	}
	return false
}

// From 从任意error转为BizError，非业务错误统一转为服务异常
func From(err error) *BizError {
	if err == nil {
		return nil
	}
	var be *BizError
	if errors.As(err, &be) {
		return be
	}
	// 原生系统错误统一封装为500
	return Wrap(errcode.System, "服务繁忙，请稍后重试", err.Error(), err)
}

// GetStack 获取调用堆栈信息，skip 跳过几层调用栈
func GetStack(skip int) string {
	const maxStackDepth = 32
	var buf bytes.Buffer
	stack := make([]uintptr, maxStackDepth)
	length := runtime.Callers(skip+1, stack[:])
	frames := runtime.CallersFrames(stack[:length])

	for {
		frame, more := frames.Next()
		buf.WriteString(frame.Function)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
		buf.WriteByte('\n')
		if !more {
			break
		}
	}
	return buf.String()
}
