package conn

import (
	"bufio"
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
	"time"
)

var maxErrCount = 10

type Service struct {
	conn      net.Conn
	writeBuf  *bufio.Writer
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

func newConnSvc(conn net.Conn, connSvcId uint32, mgrCtx context.Context) *Service {
	cxt, cancel := context.WithCancel(mgrCtx)
	connSvc := &Service{
		id:        connSvcId,
		conn:      conn,
		writeBuf:  bufio.NewWriterSize(conn, 4096),
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

func (cs *Service) startService() {
	defer connMgr.stopConnSvc(cs.id)
	sys.SafeGo(cs.readLoop)
	for {
		select {
		case <-cs.cxt.Done():
			// 收到关闭信号，主循环退出
			return

		case data := <-cs.msgQueue:
			cs.SendMsg(data)

		case buf := <-cs.recvQueue:
			irouter.Reg().Dispatch(cs, buf)
		}
	}
}

// 读网络协程
func (cs *Service) readLoop() {
	buf := make([]byte, 1024)
	for {
		// 先非阻塞检查是否要关闭
		// 读网络（阻塞，会被 conn.stopConn() 打断）
		n, err := cs.conn.Read(buf)

		if err != nil {
			// 网络错误，触发关闭
			connMgr.stopConnSvc(cs.id)
			if errors.Is(err, io.EOF) {
				return
			}
			if !errors.Is(err, net.ErrClosed) {
				logger.Errorf("角色id[%d]关闭读网络协程，错误:%v", cs.RoleID, err)
			}
			return
		}

		// 已经关闭就不要再发
		if cs.closed.Load() {
			return
		}

		// 复制数据，防止 buf 被复用污染
		data := make([]byte, n)
		copy(data, buf[:n])

		select {
		case cs.recvQueue <- data:
		case <-cs.cxt.Done():
			//fmt.Println("关闭读网络协程（ctx取消）")
			return
		}
	}
}

func (cs *Service) SendSuccess(msg proto.Message) {
	data := packet.Success(msg)
	cs.SendMsg(data)
}

func (cs *Service) SendFail(msg proto.Message, errCode uint32) {
	data := packet.Fail(msg, errCode)
	cs.SendMsg(data)
}

func (cs *Service) SendMsg(msg proto.Message) {
	data, err := packet.EncodeMsg(msg)
	if err != nil {
		return
	}
	cs.SendPacket(data)

}

func (cs *Service) SendPacket(buf []byte) {
	defer sys.Recover("SendPacket")
	err := cs.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		logger.Errorf("set write dead line err %v", err)
		return
	}

	_, err = cs.writeBuf.Write(buf)
	if err != nil {
		logger.Errorf("write data err %v", err)
		return
	}
	// 强制刷到网络，小包立刻发送
	if err = cs.writeBuf.Flush(); err != nil {
		logger.Errorf("flush data err %v", err)
		return
	}
	cs.AddSendCount()
}

//func (cs *Service) SendPacket(buf []byte) {
//	defer sys.Recover("SendPacket")
//	logger.SendProto(cs.account, cs.RoleID, buf)
//	total := 0
//	for total < len(buf) {
//		n, err := cs.conn.Write(buf[total:])
//		total += n
//		if err != nil {
//			logger.Errorf("发包错误%v", buf)
//			return
//		}
//	}
//	return
//}

func (cs *Service) SetAuthData(authData map[string]any) {
	cs.account = authData["account"].(string)
}

func (cs *Service) GetAccount() string {
	return cs.account
}

func (cs *Service) SetRoleRpc(roleId uint64, roleRpc *rpc.Service) {
	cs.RoleID = roleId
	cs.RoleRpc = roleRpc
}

func (cs *Service) AddRecvCount() {
	cs.recvCount++
}

func (cs *Service) AddSendCount() {
	cs.sendCount++
}

func (cs *Service) GetRoleConn() *model.Conn {
	conn := &model.Conn{
		ID:       cs.id,
		MsgQueue: cs.msgQueue,
		Closed:   &cs.closed,
		StopFn: func() {
			connMgr.stopConnSvc(cs.id)
		},
	}
	cs.SceneRpc = &conn.SceneRpc
	return conn
}

func (cs *Service) AssertErrCount() {
	cs.ErrCount++
	if cs.ErrCount >= maxErrCount {
		connMgr.stopConnSvc(cs.id)
	}
}

func (cs *Service) GetRoleRpc() *rpc.Service {
	return cs.RoleRpc
}
func (cs *Service) GetRoleID() uint64 {
	return cs.RoleID
}

var _ igate.ConnSvc = new(Service)
