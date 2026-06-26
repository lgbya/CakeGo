package igate

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"google.golang.org/protobuf/proto"
)

type ConnMgr interface{}

type ConnSvc interface {
	SetAuthData(map[string]any) //设置账号
	GetAccount() string         //设置账号
	AddRecvCount()              //接收的次数+1
	AddSendCount()              //发送的次数+1
	SendMsg(proto.Message)
	SendSuccess(proto.Message)
	SendFail(proto.Message, uint32)
	SetRoleRpc(uint64, *rpc.Service) //登录成功后 设置角色id
	GetRoleConn() *model.Conn
	GetRoleRpc() *rpc.Service
	GetRoleID() uint64
}
