package client

import (
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
)

type CbFn struct {
	msg proto.Message
	fn  func(uint32, proto.Message)
}

func (c *Client) regCbFn() {
	cbFnList := []CbFn{
		{&pb.AccountAuthS2C{}, c.AccountAuthS2C},   //1验证账号sdk
		{&pb.SelectRolesS2C{}, c.SelectRolesS2C},   //2查询角色信息
		{&pb.CreateRoleS2C{}, c.CreateRoleS2C},     //3没有角色就创建角色
		{&pb.LoginRoleS2C{}, c.LoginRoleS2C},       //4有角色就登陆
		{&pb.HeartbeatS2C{}, c.HeartbeatS2C},       //5登陆成功发心跳包
		{&pb.LoginEnterS2C{}, nil},                 //登录发送
		{&pb.EnterSceneS2C{}, c.EnterSceneS2C},     //7进入场景成功
		{&pb.RoleViewListS2C{}, c.RoleViewListS2C}, //8获取九宫格视野
		{&pb.MovePosS2C{}, c.MovePosS2C},           //9移动
		{&pb.RoleViewDelS2C{}, nil},                //9移动
	}
	c.cbFnMap = make(map[uint32]CbFn)
	for _, cbFn := range cbFnList {
		code, _, _ := pb.GetAllCmdByMsg(cbFn.msg)
		c.cbFnMap[code] = cbFn

	}
}
