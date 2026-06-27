package connsvc

import (
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"context"
	"github.com/puzpuzpuz/xsync/v3"
	"sync"
	"sync/atomic"
)

var connMgr = newConnManager()

type manager struct {
	nexSeq   uint32
	wg       sync.WaitGroup
	connSvcs *xsync.MapOf[uint32, *Service]
	mu       sync.Mutex
	cxt      context.Context
	cancel   context.CancelFunc
}

func newConnManager() *manager {
	cxt, cancel := context.WithCancel(context.Background())
	return &manager{
		connSvcs: xsync.NewMapOf[uint32, *Service](),
		cxt:      cxt,
		cancel:   cancel,
	}
}

func (m *manager) startService(conn IConn) {
	defer func() {
		if err := recover(); err != nil {
			conn.Close(0)
		}
	}()
	svcID := m.GenID()
	connSvc := newConnSvc(conn, svcID, m.cxt)
	m.connSvcs.Store(svcID, connSvc)
	sys.SafeGo(func() {

		m.wg.Add(1)
		defer m.wg.Done()

		connSvc.startService()
	})
}

func (m *manager) stopService(svcID uint32) {
	connSvc, ok := m.connSvcs.LoadAndDelete(svcID)
	if !ok {
		return
	}
	if connSvc.closed.Swap(true) {
		return
	}
	err := connSvc.conn.Close(svcID)
	if err != nil {
		logger.Errorf("[%d]关闭网络错误：%v", svcID, err)
		return
	}

	// 发关闭信号：唤醒所有阻塞的 select
	connSvc.cancel()

	if connSvc.RoleRpc == nil || connSvc.RoleID == 0 {
		return
	}
	connSvc.RoleRpc.Send5s(rpcid.RpcConnClose, nil)
	logger.Debugf("[%d]角色关闭conn", connSvc.RoleID)

}

func (m *manager) GenID() uint32 {
	atomic.AddUint32(&m.nexSeq, 1)
	return m.nexSeq
}

func (m *manager) stop() {
	m.cancel()
	m.wg.Wait()
	logger.Errorf("所有conn service已全部退出")
}

// ========================
func StartService(conn IConn) {
	connMgr.startService(conn)
}

func StopService(id uint32) {
	connMgr.stopService(id)
}

func StopManager() {
	connMgr.stop()
}
