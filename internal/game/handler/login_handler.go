package handler

import (
	"cake/internal/game/logic/login"
	"cake/internal/gate/igate"
	"cake/internal/gate/router/irouter"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
	"time"
)

type LoginRoute struct {
}

func (r *LoginRoute) Register() {
	irouter.Reg().ConnCmd(&pb.HeartbeatC2S{}, r.HeartbeatC2S)
	irouter.Reg().ConnCmd(&pb.AccountAuthC2S{}, r.AccountAuthC2S)
	irouter.Reg().ConnCmd(&pb.SelectRolesC2S{}, r.SelectRolesC2S)
	irouter.Reg().ConnCmd(&pb.CreateRoleC2S{}, r.RoleCreateC2S)
	irouter.Reg().ConnCmd(&pb.LoginRoleC2S{}, r.LoginRoleC2S)
}

func (*LoginRoute) HeartbeatC2S(connSvc igate.ConnSvc, rawMsg proto.Message) {
	msg := rawMsg.(*pb.HeartbeatC2S)
	roleRpc := connSvc.GetRoleRpc()
	if roleRpc != nil && connSvc.GetRoleID() > 0 {
		roleRpc.Send5s(rpcid.RpcHeartbeat, msg.ClientTime)
	}
	connSvc.SendMsg(&pb.HeartbeatS2C{ServerTime: time.Now().Unix(), ClientTime: msg.ClientTime})
}

func (*LoginRoute) AccountAuthC2S(connSvc igate.ConnSvc, rawMsg proto.Message) {
	msg := rawMsg.(*pb.AccountAuthC2S)
	login.Logic().AccountAuth(connSvc, msg)
}

func (*LoginRoute) SelectRolesC2S(connSvc igate.ConnSvc, rawMsg proto.Message) {
	msg := rawMsg.(*pb.SelectRolesC2S)
	login.Logic().SelectRoles(connSvc, msg)
}

func (*LoginRoute) RoleCreateC2S(connSvc igate.ConnSvc, rawMsg proto.Message) {
	msg := rawMsg.(*pb.CreateRoleC2S)
	login.Logic().CreateRole(connSvc, msg)
}

func (*LoginRoute) LoginRoleC2S(connSvc igate.ConnSvc, rawMsg proto.Message) {
	roleLoginC2S := rawMsg.(*pb.LoginRoleC2S)
	login.Logic().LoginRole(connSvc, roleLoginC2S)
}
