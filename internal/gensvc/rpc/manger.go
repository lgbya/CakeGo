package rpc

import (
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"context"
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type rpcFn func(any, any, any) (any, error)

var mgr = newManager()

type startInfo struct {
	isSuccess bool
	err       error
}

type Manager struct {
	services *xsync.MapOf[string, *Service]
	ctx      context.Context    // 全局根上下文(进程退出用)
	cancel   context.CancelFunc // 全局取消函数
	Routes   *xsync.MapOf[reflect.Type, *map[string]rpcFn]
	wg       sync.WaitGroup
	closed   atomic.Bool
	nextSeq  uint64
}

func newManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:      ctx,
		cancel:   cancel,
		services: xsync.NewMapOf[string, *Service](),
		Routes:   xsync.NewMapOf[reflect.Type, *map[string]rpcFn](),
	}
}

func (m *Manager) StartSvc(name string, module GenService, cfg Cfg) (*Service, error) {
	if m.closed.Load() {
		return nil, errors.New("已经关闭服务器完成重新创建service")
	}

	defer sys.Recover(fmt.Sprintf("[%s]StartSvc失败", name))

	if cfg.isInit {
		return nil, errors.New("配置未初始化")
	}

	// 1. 空入参防护，直接返回空路由避免反射panic
	if module == nil {
		logger.Errorf("[gensvc autoReg] warn: input GenService is nil")
		return nil, errors.New("input GenService is nil")
	}

	if _, ok := m.getService(name); ok {
		return nil, errors.New("service exists")
	}

	svcID := m.genID()
	s, err := NewService(svcID, name, module, m.ctx, cfg)
	if err != nil {
		return nil, err
	}

	//如果查出来ok为true则不会
	if _, ok := m.services.LoadOrStore(name, s); ok {
		return nil, errors.New("service exists")
	}

	startChan := make(chan startInfo)
	sys.SafeGo(func() {
		if cfg.Wg != nil {
			defer (*cfg.Wg).Done()
			(*cfg.Wg).Add(1)
		}
		defer m.wg.Done()
		m.wg.Add(1)

		s.loop(startChan)
	})

	timer := time.NewTicker(cfg.StartTimeout)
	defer timer.Stop()
	for {
		select {
		case startInfo := <-startChan:
			if !startInfo.isSuccess {
				return nil, startInfo.err
			}
			return s, nil
		case <-timer.C:
			s.cancel()
			m.services.Delete(name)
			return nil, errors.New("timeout")
		}
	}

}

func (m *Manager) genID() uint64 {
	return atomic.AddUint64(&m.nextSeq, 1)
}

func (m *Manager) getService(dest string) (*Service, bool) {
	svc, ok := m.services.Load(dest)
	if !ok {
		return nil, false
	}
	return svc, true
}

func (m *Manager) stopService(dest string) {
	server, ok := m.services.LoadAndDelete(dest)
	if !ok {
		return
	}
	server.cancel()
}

func (m *Manager) stopAllService() {
	if m.closed.Swap(true) {
		return
	}
	logger.Errorf("等待关闭所有service！！！")
	m.cancel()
	m.wg.Wait()
}

// ======================
func Start(name string, module GenService) (*Service, error) {
	return StartWithCfg(name, module, NewCfg())
}

func StartWithCfg(name string, module GenService, cfg Cfg) (*Service, error) {
	return mgr.StartSvc(name, module, cfg)
}

func Stop(dest string) {
	mgr.stopService(dest)
}

func StopAll() {
	mgr.stopAllService()
}

func GetService(dest string) (*Service, bool) {
	return mgr.getService(dest)
}
