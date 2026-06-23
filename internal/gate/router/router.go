package router

import (
	"cake/internal/gate/conn"
	"cake/internal/gate/packet"
	"cake/internal/gate/router/irouter"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
	"reflect"
)

const PacketBufMaxLen = 524288

type RoleProtoCmd struct {
	Msg proto.Message
	Fn  irouter.RoleRouteFn
}

type CmdHandler struct {
	cmd        uint32
	typ        reflect.Type
	connFn     irouter.ConnRouteFn
	roleFn     irouter.RoleRouteFn
	sceneFn    irouter.SceneRouteFn
	handleType int //请求类型
}

type Router struct {
	isDebug  bool
	handlers map[uint32]CmdHandler
}

func NewRouter() irouter.IRegister {
	return &Router{
		handlers: make(map[uint32]CmdHandler),
	}
}

func (r *Router) ConnCmd(c2sMsg proto.Message, fn irouter.ConnRouteFn) {
	cmd, c2sTyp, ok := pb.GetC2SCmdByMsg(c2sMsg)
	if !ok {
		return
	}
	r.handlers[cmd] = CmdHandler{cmd: cmd, typ: c2sTyp, connFn: fn, handleType: HandlerTypeConn}
}

func (r *Router) RoleCmd(c2sMsg proto.Message, fn irouter.RoleRouteFn) {
	cmd, c2sTyp, ok := pb.GetC2SCmdByMsg(c2sMsg)
	if !ok {
		return
	}
	r.handlers[cmd] = CmdHandler{cmd: cmd, typ: c2sTyp, roleFn: fn, handleType: HandlerTypeRole}
}

func (r *Router) SceneCmd(c2sMsg proto.Message, fn irouter.SceneRouteFn) {
	cmd, c2sTyp, ok := pb.GetC2SCmdByMsg(c2sMsg)
	if !ok {
		return
	}
	r.handlers[cmd] = CmdHandler{cmd: cmd, typ: c2sTyp, sceneFn: fn, handleType: HandlerTypeScene}
}

func (r *Router) Dispatch(rawConnSvc any, buf []byte) {
	defer sys.Recover("router-dispatch")

	if len(buf) >= PacketBufMaxLen {
		logger.Errorf("dispatch buffer too large")
		return
	}

	connSvc, ok := rawConnSvc.(*conn.Service)
	if !ok {
		return
	}

	connSvc.AddSendCount()
	var fullBuf []byte
	if len(connSvc.LeftBuf) > 0 {
		fullBuf = append(connSvc.LeftBuf, buf...)
		connSvc.LeftBuf = nil // 临时清空，解析后再赋值
	} else {
		fullBuf = buf
	}
	// 循环解析：处理缓冲区中所有完整数据包，解决粘包
	for len(fullBuf) > 0 {
		cmd, data, left, ok := packet.DecodeMsg(fullBuf)
		if !ok {
			// 半包：保存剩余未解析数据，跳出循环等待下次接收
			connSvc.LeftBuf = left
			break
		}

		// 解析成功，清空半包缓存
		connSvc.LeftBuf = nil
		// 业务处理当前包
		r.handlePacket(connSvc, cmd, data)

		// 把剩余未解析的数据作为下一轮待解析缓冲区
		fullBuf = left
	}
}

func (r *Router) handlePacket(connSvc *conn.Service, cmd uint32, data []byte) {
	route, ok := r.handlers[cmd]
	if !ok {
		logger.Errorf("没有找到对应的协议号:%d", cmd)
		connSvc.AssertErrCount()
		return
	}

	msg := reflect.New(route.typ.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(data, msg); err != nil {
		logger.Errorf("解包错误cmd:%d, msg:%v ", cmd, msg)
		connSvc.AssertErrCount()
		return
	}

	roleID := connSvc.RoleID
	if route.handleType == HandlerTypeScene && (connSvc.RoleRpc == nil || roleID == 0) {
		logger.Errorf("还没登陆无法转发: %d", cmd)
		connSvc.AssertErrCount()
		return
	}

	if route.handleType == HandlerTypeScene && connSvc.SceneRpc == nil {
		logger.Errorf("还没登陆无法转发: %d", cmd)
		connSvc.AssertErrCount()
		return
	}
	connSvc.ErrCount = 0
	logger.RecvProto(connSvc.GetAccount(), connSvc.RoleID, cmd, msg)
	//RecvProto(connSvc, cmd, msg)
	switch route.handleType {
	case HandlerTypeRole:
		//角色路由
		connSvc.RoleRpc.Send5s(rpcid.RpcRoleCmd, irouter.RoleCmd{RoleID: roleID, Msg: msg, RoleFn: route.roleFn})
	case HandlerTypeScene:
		//进程路由
		(*connSvc.SceneRpc).Send5s(rpcid.RpcRoleCmd, irouter.RoleCmd{RoleID: roleID, Msg: msg, SceneFn: route.sceneFn})
	default:
		//登录连接路由
		route.connFn(connSvc, msg)
	}
}

var _ irouter.IRegister = new(Router)
