package connsvc

import (
	"cake/internal/game/model"
	"cake/internal/gate/igate"
	"cake/internal/gate/packet"
	"cake/internal/gate/router/irouter"
	"cake/internal/gensvc/rpc"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"context"
	"errors"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var maxErrCount = 10

type Service struct {
	conn      IConn
	id        uint32
	account   string
	RoleID    uint64
	recvCount uint32
	sendCount uint32
	ErrCount  int
	msgQueue  chan proto.Message
	recvQueue chan []byte
	closeChan chan struct{}
	closed    atomic.Bool
	cxt       context.Context
	cancel    context.CancelFunc
	SceneRpc  **rpc.Service
	RoleRpc   *rpc.Service
	mu        sync.Mutex
	LeftBuf   []byte
}

func newConnSvc(conn IConn, connSvcId uint32, mgrCtx context.Context) *Service {
	cxt, cancel := context.WithCancel(mgrCtx)
	connSvc := &Service{
		id:        connSvcId,
		conn:      conn,
		msgQueue:  make(chan proto.Message, 256),
		recvQueue: make(chan []byte, 256),
		closeChan: make(chan struct{}),
		cxt:       cxt,
		cancel:    cancel,
	}
	//svcName := "conn_svc_" + strconv.Itoa(connSvcId)
	//gensvc.Start(svcName)
	return connSvc
}

func (s *Service) startService() {
	defer connMgr.stopService(s.id)
	sys.SafeGo(s.readLoop)
	for {
		select {
		case <-s.cxt.Done():
			// 收到关闭信号，主循环退出
			return

		case data := <-s.msgQueue:
			s.SendMsg(data)

		case buf := <-s.recvQueue:
			irouter.Reg().Dispatch(s, buf)
		}
	}
}

// 读网络协程
func (s *Service) readLoop() {
	buf := make([]byte, 1024)
	for {
		// 先非阻塞检查是否要关闭
		// 读网络（阻塞，会被 conn.stopConn() 打断）
		n, err := s.conn.Read(buf)

		if err != nil {
			// 网络错误，触发关闭
			connMgr.stopService(s.id)
			if errors.Is(err, io.EOF) {
				return
			}
			if !errors.Is(err, net.ErrClosed) {
				logger.Errorf("角色id[%d]关闭读网络协程，错误:%v", s.RoleID, err)
			}
			return
		}

		// 已经关闭就不要再发
		if s.closed.Load() {
			return
		}

		// 复制数据，防止 buf 被复用污染
		data := make([]byte, n)
		copy(data, buf[:n])

		select {
		case s.recvQueue <- data:
		case <-s.cxt.Done():
			//fmt.Println("关闭读网络协程（ctx取消）")
			return
		}
	}
}

func (s *Service) SendSuccess(msg proto.Message) {
	data := packet.Success(msg)
	s.SendMsg(data)
}

func (s *Service) SendFail(msg proto.Message, errCode uint32) {
	data := packet.Fail(msg, errCode)
	s.SendMsg(data)
}

func (s *Service) SendMsg(msg proto.Message) {
	data, err := packet.EncodeMsg(msg)
	if err != nil {
		return
	}
	s.conn.Send(data)
	s.AddSendCount()
}

func (s *Service) SetAuthData(authData map[string]any) {
	s.account = authData["account"].(string)
}

func (s *Service) GetAccount() string {
	return s.account
}

func (s *Service) SetRoleRpc(roleId uint64, roleRpc *rpc.Service) {
	s.RoleID = roleId
	s.RoleRpc = roleRpc
}

func (s *Service) AddRecvCount() {
	s.recvCount++
}

func (s *Service) AddSendCount() {
	s.sendCount++
}

func (s *Service) GetRoleConn() *model.Conn {
	conn := &model.Conn{
		ID:       s.id,
		MsgQueue: s.msgQueue,
		Closed:   &s.closed,
		StopFn: func() {
			connMgr.stopService(s.id)
		},
	}
	s.SceneRpc = &conn.SceneRpc
	return conn
}

func (s *Service) AssertErrCount() {
	s.ErrCount++
	if s.ErrCount >= maxErrCount {
		connMgr.stopService(s.id)
	}
}

func (s *Service) GetRoleRpc() *rpc.Service {
	return s.RoleRpc
}
func (s *Service) GetRoleID() uint64 {
	return s.RoleID
}

var _ igate.ConnSvc = new(Service)
