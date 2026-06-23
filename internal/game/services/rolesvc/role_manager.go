package rolesvc

import (
	"cake/internal/game/def/errcode"
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"cake/internal/util/errx"
	"context"
	"github.com/puzpuzpuz/xsync/v3"
	"sync"
)

var mgr = newManager()

type manager struct {
	wg       *sync.WaitGroup
	mu       sync.RWMutex
	cxt      context.Context
	cancel   context.CancelFunc
	nextSeq  int
	RoleSvcs *xsync.MapOf[uint64, *Service] //角色
}

func newManager() *manager {
	cxt, cancel := context.WithCancel(context.Background())
	return &manager{
		cxt:      cxt,
		cancel:   cancel,
		wg:       &sync.WaitGroup{},
		RoleSvcs: xsync.NewMapOf[uint64, *Service](),
	}
}

func (m *manager) StartRole(roleState *model.Role) (*rpc.Service, error) {
	roleID := roleState.RoleID
	var retSvc *rpc.Service
	var retErr error
	// Compute：key存在时，在当前Bucket锁内原子执行回调，杜绝竞态
	m.RoleSvcs.Compute(roleID, func(oldRoleSvc *Service, exists bool) (*Service, bool) {
		if !exists {
			return oldRoleSvc, false
		}

		if oldRoleSvc.IsClosed() {
			// 角色服务已关闭，复用，更新玩家状态
			oldRoleSvc.Service.UpdateState(roleState)
			retSvc = oldRoleSvc.Service
			retErr = nil
			// 保留key，不删除
			return oldRoleSvc, false
		}

		// 服务正常运行，拒绝重复登录
		retSvc = nil
		retErr = errx.New(errcode.LoginRepeat)
		return oldRoleSvc, false
	})
	if retErr != nil || retSvc != nil {
		return retSvc, retErr
	}

	cfg := rpc.NewCfg()
	cfg.Ctx = m.cxt
	cfg.Wg = m.wg
	cfg.InitArgs = roleState
	roleRpc, roleSvc, err := StartService(roleState, cfg)
	if err != nil {
		return nil, err
	}
	m.RoleSvcs.Store(roleState.RoleID, roleSvc)
	return roleRpc, err
}

func (m *manager) StopRole(roleID uint64) {
	m.RoleSvcs.Compute(roleID, func(s *Service, exists bool) (*Service, bool) {
		if !exists {
			return nil, false
		}
		s.Service.Stop()
		return nil, true
	})
}

func (m *manager) Stop() {
	m.cancel()
	m.wg.Wait()
}

func StartRole(roleState *model.Role) (*rpc.Service, error) {
	return mgr.StartRole(roleState)
}

func Stop() {
	mgr.cancel()
	mgr.wg.Wait()
}
