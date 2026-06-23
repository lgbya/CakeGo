package rpc

import (
	"cake/internal/gensvc/router"
	"cake/internal/gensvc/timer"
	"cake/internal/pkg/logger"
	"cake/internal/pkg/metric"
	"cake/internal/util/sys"
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Timeout5s  = 5 * time.Second
	Timeout10s = 10 * time.Second

	PendingMaxLen = 1000
)

type GenService interface {
	Init(*Service, any) (any, error)
	Stop(any)
}

type Service struct {
	id         uint64
	name       string
	sendQueue  chan *Msg
	callChan   chan *Msg
	routes     map[string]router.RpcFn
	genService GenService
	state      any
	*timer.GbTree
	ctx         context.Context
	cancel      context.CancelFunc
	closed      atomic.Bool
	cfg         Cfg
	pending     *pending
	trace       string
	msgPool     sync.Pool
	tickerQueue chan struct{}
	isInit      bool
}

func NewService(id uint64, name string, service GenService, rootCtx context.Context, cfg Cfg) (*Service, error) {

	routes, ok := router.Gen.GetRoutes(reflect.TypeOf(service))
	if !ok {
		return nil, errors.New("rpcgen routes error")
	}
	if routes == nil || len(routes) == 0 {
		return nil, errors.New(fmt.Sprintf("[%s] 静态路由表不能为空，请检查go generate是否执行生成路由代码", name))
	}

	ctx, cancel := context.WithCancel(rootCtx)
	return &Service{
		id:          id,
		name:        name,
		genService:  service,
		sendQueue:   make(chan *Msg, cfg.SendMaxCap),
		callChan:    make(chan *Msg),
		tickerQueue: make(chan struct{}, 100),
		GbTree:      timer.NewGbTree(),
		ctx:         ctx,
		cancel:      cancel,
		routes:      routes,
		cfg:         cfg,
		pending:     newPending(),
		msgPool: sync.Pool{New: func() interface{} {
			return &Msg{}
		}},
	}, nil

}

func (s *Service) TickerQueue() chan struct{} {
	return s.tickerQueue
}
func (s *Service) PID() uint64 {
	return s.id
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) State() any {
	return s.state
}

func (s *Service) GetCtx() context.Context {
	return s.ctx
}

func (s *Service) Send5s(cmd string, args any) bool {
	return sendMsgTimeout(s, 0, cmd, args, Timeout5s)
}

func (s *Service) SendTimeout(cmd string, args any, timeout time.Duration) bool {
	return sendMsgTimeout(s, 0, cmd, args, timeout)
}

func (s *Service) SendAfter5s(delay time.Duration, cmd string, args any) bool {
	return sendMsgTimeout(s, delay, cmd, args, Timeout5s)
}

func (s *Service) SendAfterTimeout(delay time.Duration, cmd string, args any, timeout time.Duration) bool {
	return sendMsgTimeout(s, delay, cmd, args, timeout)
}

func (s *Service) Call5s(cmd string, args any) (any, error) {
	return callMsgTimeout(s, cmd, args, Timeout5s)
}

func (s *Service) Call10s(cmd string, args any) (any, error) {
	return callMsgTimeout(s, cmd, args, Timeout10s)
}

func (s *Service) CallTimeout(cmd string, args any, timeout time.Duration) (any, error) {
	return callMsgTimeout(s, cmd, args, timeout)
}

func (s *Service) DoMsgFn(msg *Msg) (any, error) {
	cmd := msg.Cmd
	funArgs := msg.Args
	fn, ok := s.routes[cmd]
	if !ok {
		logger.Errorf("[%s]不存在路由:%s", s.name, cmd)
		return nil, errors.New("路由不存在")
	}
	return fn(s.genService, s.state, funArgs)
}
func (s *Service) GetSendQueueStat() (int, int, bool) {
	return len(s.sendQueue), cap(s.sendQueue), s.closed.Load()
}

func (s *Service) loop(startChan chan startInfo) {
	defer s.stop()
	initState, err := s.genService.Init(s, s.cfg.InitArgs)
	if err != nil {
		startChan <- startInfo{isSuccess: false, err: err}
		return
	}
	s.state = initState
	s.isInit = true
	startChan <- startInfo{isSuccess: true, err: nil}

	var cfgCtxChan <-chan struct{}
	if s.cfg.Ctx != nil {
		cfgCtxChan = s.cfg.Ctx.Done()
	}

	sendFn := s.cfg.SendFn
	isInternal := false
	if sendFn == nil {
		sendFn = s.handleInfo
		isInternal = true // 内部路由，run内自动归还Msg
	}
	timer.Register(s)
	for {
		select {
		case msg := <-s.sendQueue:
			s.run(sendFn, msg, isInternal)
		case msg := <-s.callChan:
			s.run(s.handleCall, msg, true)
		case <-s.tickerQueue:
			used, _, _ := s.GetSendQueueStat()
			metric.SendQueueLen.WithLabelValues(s.name).Set(float64(used))
			s.Tick(s)
		case <-s.ctx.Done():
			return
		case <-cfgCtxChan:
			return
		}
	}
}
func (s *Service) run(fn func(*Msg) error, msg *Msg, isInternal bool) {
	defer sys.Recover(s.name)
	if isInternal {
		defer s.PutMsg(msg)
	}
	start := time.Now()
	cmd := msg.Cmd

	var err error

	s.before(msg)
	err = fn(msg)
	s.after(msg)

	costMs := float64(time.Since(start).Milliseconds())
	metric.RpcDuration.WithLabelValues(s.name, cmd).Observe(costMs)

	// todo：慢调用判断,后面加上令牌熔断再开出来
	//if time.Since(start) >= meta.SlowThresh {
	//	metric.RpcSlowTotal.WithLabelValues(s.name, cmd).Inc()
	//}

	// 业务错误埋点
	if err != nil {
		metric.RpcErrTotal.WithLabelValues(s.name, cmd, "biz_error").Inc()
	}
}

func (s *Service) before(msg *Msg) {

}

func (s *Service) after(msg *Msg) {
}

func (s *Service) handleInfo(msg *Msg) error {
	cmd := msg.Cmd
	funArgs := msg.Args
	delay := msg.Delay
	fn, ok := s.routes[cmd]
	if !ok {
		return errors.New(fmt.Sprintf("[%s]不存在路由:%s", s.name, cmd))
	}

	if delay <= 0 {
		_, err := fn(s.genService, s.state, funArgs)
		if err == nil {
			s.retryPendingMsg()
		}
		return err
	}

	// 延迟消息：注册到定时器，到期后切回角色协程执行
	timerName := fmt.Sprintf("delay_%s_%d", cmd, time.Now().UnixNano())
	// 这里的回调不直接执行 fn，而是把执行指令发回 sendQueue
	s.AddTimer(timerName, delay, 1, func(state any, _ any) error {
		// 到期时，发一个 delay=0 的消息，让角色自己处理
		s.SendTimeout(cmd, funArgs, 1*time.Second)
		return nil
	}, nil)

	return nil
}

func (s *Service) handleCall(msg *Msg) error {
	defer sys.Recover(s.name)

	cmd := msg.Cmd
	funArgs := msg.Args
	cb := msg.CbChat
	fn, ok := s.routes[cmd]
	if !ok {
		// 安全写入错误，防止阻塞
		select {
		case cb <- errors.New("cmd not found: " + cmd):
		default:
			// 调用方已退出，无需写入
		}
		return errors.New(fmt.Sprintf("[%s]不存在路由:%s", s.name, cmd))
	}

	// 2. 执行业务函数
	result, err := fn(s.genService, s.state, funArgs)

	// 回写前再次校验：防止执行业务期间调用超时
	select {
	case <-msg.CallCtx.Done():
		return errors.New(fmt.Sprintf("[%s]:%s业务超时", s.name, cmd))
	default:
	}

	// 正常回写：通道永远不会处于已关闭状态，裸写完全安全
	if err != nil {
		msg.CbChat <- err
	} else {
		msg.CbChat <- result
	}
	return err

}

func (s *Service) isClosed() bool {
	return s.closed.Load()
}

func (s *Service) UpdateState(state any) {
	s.state = state
}

func (s *Service) retryPendingMsg() {
	sendLen, _, _ := s.GetSendQueueStat()

	batchMsgs, batchNum := s.pending.getBatchMsgs(sendLen, s.cfg.SendMaxCap)
	if batchNum <= 0 {
		return
	}

	failIdx := -1
	for idx, msg := range batchMsgs {
		select {
		case s.sendQueue <- msg:
		default:
			failIdx = idx
			break
		}
	}

	if failIdx != -1 {
		s.pending.mu.Lock()
		defer s.pending.mu.Unlock()
		for _, failMsg := range batchMsgs[failIdx:] {
			s.pending.pushLocked(s.name, failMsg)
		}
	}
}

func (s *Service) stop() {

	if s.closed.Swap(true) {
		return
	}
	//先停定时器
	s.GbTree.Close(s)
	//初始化成功才，执行结束方法
	if s.isInit {
		sys.SafeRun(func() { s.genService.Stop(s.state) })
	}

	//关闭消息
	s.pending.Clear(s)
	s.closeMsgChan(s.sendQueue)
	s.closeMsgChan(s.callChan)

	//清空一切
	s.routes = nil
	s.state = nil
	s.genService = nil
	s.ctx = nil
	s.cancel = nil
	s.GbTree = nil
	s.pending = nil
	s.callChan = nil
	s.sendQueue = nil

}

func (s *Service) closeMsgChan(msgChan chan *Msg) {
	//排空未处理消息，防止内存泄漏
	for {
		select {
		case msg := <-msgChan:
			s.PutMsg(msg)
		default:
			goto close
		}
	}
close:
	sys.SafeCloseChan(msgChan)
}

func (s *Service) Stop() {
	Stop(s.name)
}

// 获取+自动重置
func (s *Service) GetMsg() *Msg {
	msg := s.msgPool.Get().(*Msg)
	msg.Reset()
	return msg
}

// 使用完归还
func (s *Service) PutMsg(msg *Msg) {
	s.msgPool.Put(msg)
}

// =====================================
func Send(dest string, cmd string, args any, timeout time.Duration) bool {
	return SendAfter(0, dest, cmd, args, timeout)
}

func SendAfter(delay time.Duration, dest string, cmd string, args any, timeout time.Duration) bool {
	svc, ok := mgr.getService(dest)
	if !ok {
		return false
	}
	return sendMsgTimeout(svc, delay, cmd, args, timeout)
}

func sendMsgTimeout(s *Service, delay time.Duration, cmd string, args any, timeout time.Duration) bool {
	if s.isClosed() {
		return false
	}
	msg := s.GetMsg()
	msg.Cmd = cmd
	msg.Args = args
	msg.Delay = delay
	if len(s.sendQueue) >= s.cfg.SendMaxCap {
		s.pending.mu.Lock()
		defer s.pending.mu.Unlock()
		s.pending.pushLocked(s.name, msg)
		return true
	}
	//trace := fmt.Sprintf("%s->%s.%s", s.trace, s.name, cmd)
	return sys.SafeSendTimeout(s.sendQueue, msg, timeout)
}

func Call5s(dest string, cmd string, args any) (any, error) {
	return CallWithTimeout(dest, cmd, args, Timeout5s)
}

func CallWithTimeout(dest string, cmd string, args any, timeout time.Duration) (any, error) {
	svc, ok := mgr.getService(dest)
	if !ok {
		return nil, errors.New("不存在Service！！！")
	}
	return callMsgTimeout(svc, cmd, args, timeout)
}

func callMsgTimeout(svc *Service, cmd string, args any, timeout time.Duration) (any, error) {
	if svc.closed.Load() {
		return nil, errors.New("service closed")
	}

	// 调用总超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cb := make(CbChan, 1) // 加缓冲，避免极端阻塞
	msg := svc.GetMsg()
	msg.Cmd = cmd
	msg.Args = args
	msg.CbChat = cb
	msg.CallCtx = ctx
	//  写入 callChan：防止 callChan 满 / 服务退出导致永久阻塞
	select {
	case svc.callChan <- msg:
	// 超时 / 服务退出
	case <-ctx.Done():
		return nil, errors.New("call send timeout")
	case <-svc.ctx.Done():
		return nil, errors.New("service already shutdown")
	}

	//  等待返回结果，同样受超时控制
	select {
	case res := <-cb:
		close(cb)
		if err, ok := res.(error); ok && err != nil {
			return nil, err
		}
		return res, nil
	case <-ctx.Done():
		return nil, errors.New("call send timeout")
	case <-svc.ctx.Done():
		return nil, errors.New("service already shutdown")
	}
}
