package battle

//
//import (
//	"cake/internal/game/model"
//	"cake/internal/gate/router/irouter"
//	"cake/internal/pkg/logger"
//	"cake/internal/pkg/sys"
//	"context"
//	"time"
//)
//
//const interval = 50 * time.Millisecond
//
//type RpcFn func(*model.State, any) error
//
//type Msg struct {
//	Fn   RpcFn
//	Args any
//}
//
//type Service struct {
//	interval time.Duration
//	cmdQueue chan irouter.RoleCmd
//	msgCache []irouter.RoleCmd
//	rpcQueue chan Msg
//	rpcCache []Msg
//	State    *model.State
//	ctx      context.Context
//	cancel   context.CancelFunc
//}
//
//func NewService(State *model.State) *Service {
//	return &Service{
//		interval: interval,
//		cmdQueue: make(chan irouter.RoleCmd, 1024),
//		msgCache: make([]irouter.RoleCmd, 32),
//		rpcQueue: make(chan Msg, 1024),
//		rpcCache: make([]Msg, 32),
//		State:    State,
//	}
//}
//
//func (b *Service) StartService(tag string, sceneCtx context.Context) {
//
//	b.ctx, b.cancel = context.WithCancel(sceneCtx)
//	sys.SafeGo(func() {
//		// 退出时自动执行Close，做资源清理
//		defer b.Stop()
//
//		timer := time.NewTicker(b.interval)
//		defer timer.Stop()
//
//		logger.Debugf("开启战斗帧进程，帧间隔: %v", b.interval)
//
//		for {
//			select {
//			case <-b.ctx.Done():
//				// 收到退出信号，直接退出循环
//				return
//			case <-timer.C:
//				// 每帧触发，执行帧逻辑
//				sys.SafeRun(b.doWork)
//			}
//		}
//	})
//}
//
//func (b *Service) doWork() {
//	// 1 先处理玩家指令
//	sys.SafeRun(b.doRoleCmd)
//	sys.SafeRun(b.doRoleCmd)
//
//	//b.broadcastFrame()
//}
//
//func (b *Service) doRoleCmd() {
//	b.msgCache = b.msgCache[:0]
//	for {
//		select {
//		case msg := <-b.cmdQueue:
//			b.msgCache = append(b.msgCache, msg)
//		default:
//			for _, msg := range b.msgCache {
//				sys.SafeRun(func() {
//					errx := msg.SceneFn(b.State, msg.BcastRoleID, msg.Msg)
//					if errx != nil {
//						logger.Errorf("战斗进程处理协议错误 %v %v", msg, errx)
//					}
//				})
//			}
//			return
//		}
//	}
//}
//
//func (b *Service) SendMsg(msg Msg) {
//	sys.SafeSendWait(b.rpcQueue, msg)
//}
//
//func (b *Service) SendCmd(cmd irouter.RoleCmd) {
//	sys.SafeSendWait(b.cmdQueue, cmd)
//}
//
//func (b *Service) Stop() {
//
//}
