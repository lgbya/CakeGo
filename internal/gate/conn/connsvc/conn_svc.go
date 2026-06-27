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

type Packet struct {
	Cmd  uint32
	Data []byte
}

const PacketBufMaxLen = 524288

var maxErrCount = 10

type Service struct {
	conn        IConn
	id          uint32
	account     string
	RoleID      uint64
	recvCount   uint32
	sendCount   uint32
	ErrCount    int
	msgQueue    chan proto.Message
	packetQueue chan *Packet
	closeChan   chan struct{}
	closed      atomic.Bool
	ctx         context.Context
	cancel      context.CancelFunc
	SceneRpc    **rpc.Service
	RoleRpc     *rpc.Service
	mu          sync.Mutex
	LeftBuf     []byte
}

func newConnSvc(conn IConn, connSvcId uint32, mgrCtx context.Context) *Service {
	cxt, cancel := context.WithCancel(mgrCtx)
	connSvc := &Service{
		id:          connSvcId,
		conn:        conn,
		msgQueue:    make(chan proto.Message, 256),
		packetQueue: make(chan *Packet, 256),
		closeChan:   make(chan struct{}),
		ctx:         cxt,
		cancel:      cancel,
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
		case <-s.ctx.Done():
			// 收到关闭信号，主循环退出
			return

		case data := <-s.msgQueue:
			s.SendMsg(data)

		case pkt := <-s.packetQueue:
			irouter.Reg().Dispatch(s, pkt)
		}
	}
}

// 读网络协程
func (s *Service) readLoop() {
	for {
		// 先非阻塞检查是否要关闭
		// 读网络（阻塞，会被 conn.stopConn() 打断）
		buff, err := s.conn.Read()
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

		packetList, ok := s.fullBuf(buff)
		if !ok {
			// 缓冲区超限/解析异常，直接断开连接
			connMgr.stopService(s.id)
			return
		}

		// 遍历本次所有解析成功的数据包，逐个投递业务队列
		for _, pkt := range packetList {
			select {
			case s.packetQueue <- pkt:
				// 投递成功，继续下一条
			case <-s.ctx.Done():
				// 业务主动取消，直接退出读协程
				return
			}
		}
	}
}

// fullBuf 从缓存+本次读到的数据中解析所有完整数据包
// 返回：解析成功的所有数据包、是否发生错误
func (s *Service) fullBuf(buf []byte) ([]*Packet, bool) {
	// 拼接历史半包 + 当前新读到的数据
	var fullBuf []byte
	if len(s.LeftBuf) > 0 {
		fullBuf = append(s.LeftBuf, buf...)
		s.LeftBuf = nil
	} else {
		fullBuf = buf
	}

	// 安全校验：总缓冲区不能超过最大限制，防止恶意攻击
	if len(fullBuf) > PacketBufMaxLen {
		logger.Errorf("total buffer too large, len:%d", len(fullBuf))
		return nil, false
	}

	var packetList []*Packet
	// 循环解析所有完整数据包
	for len(fullBuf) > 0 {
		cmd, data, left, ok := packet.DecodeMsg(fullBuf)
		if !ok {
			// 半包：剩余数据留存，等待下次收包继续解析
			s.LeftBuf = left
			break
		}

		// 解析出一条完整协议，加入列表
		packetList = append(packetList, &Packet{
			Cmd:  cmd,
			Data: data,
		})

		// 剩余数据继续循环解析
		fullBuf = left

		// 解析完所有数据，没有半包需要留存
		if len(fullBuf) == 0 {
			s.LeftBuf = nil
		}
	}

	return packetList, true
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
	logger.SendProto(s.account, s.RoleID, data)
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
