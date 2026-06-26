package client

import (
	"cake/internal/game/def"
	"cake/internal/game/def/errcode"
	"cake/internal/game/model"
	"cake/internal/pkg/logger"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
)

var cnt atomic.Int32

func (c *Client) AccountAuthS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.AccountAuthS2C)
	if msg.IsAuth {
		logger.Debugf("启动客户端成功 账号：%s", c.Account)
		c.SelectRolesC2S()
	}

}

func (c *Client) SelectRolesS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.SelectRolesS2C)
	if msg.CommonNotice.ErrCode != errcode.Ok {
		logger.Debugf("查询角色失败：%v", msg)
	}
	if msg.RoleList == nil {
		logger.Debugf("查询角色成功，角色个数：%d", len(msg.RoleList))
		c.CreateRoleC2S()
	} else {
		c.RoleID = msg.RoleList[0].RoleId
		c.LoginRoleC2S()
	}

}

func (c *Client) CreateRoleS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.CreateRoleS2C)
	if msg.CommonNotice.ErrCode != errcode.Ok {
		logger.Debugf("创建角色失败：%v", msg)
	}
	c.RoleID = msg.RoleId
	c.LoginRoleC2S()

}

func (c *Client) LoginRoleS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.LoginRoleS2C)
	if msg.CommonNotice.ErrCode != errcode.Ok {
		logger.Debugf("玩家[%s][%d]登录角色失败：%v", c.Account, c.RoleID, msg)
		return
	}
	logger.Debugf("玩家[%s][%d]登录角色成功：%v", c.Account, c.RoleID, msg)
	go c.HeartbeatC2S()
	c.LoginEnterC2S()
}

func (c *Client) EnterSceneS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.EnterSceneS2C)
	logger.Infof("[%s]进入场[%d],坐标[%d,%d]成功：%v", c.Account, msg.MapId, msg.Pos.X, msg.Pos.Y, msg)
	c.Pos = model.Pos{X: int(msg.Pos.X), Y: int(msg.Pos.Y)}
	c.Location.SceneID = msg.SceneId
	c.Location.MapID = msg.MapId
	if msg.MapId == def.MainSceneMapID {
		cnt.Add(1)
		if cnt.Load()%2 == 0 {
			//随机进入另一张地图
			c.EnterSceneC2S(0, 1001)
			return
		}
		c.StartAutoWalk()
	} else {
		c.StartAutoWalk()
		//c.MovePosC2S(msg.Pos.X+10, msg.Pos.Y+10)
	}
}

func (c *Client) MovePosS2C(cmd uint32, rawMsg proto.Message) {
	//msg := rawMsg.(*pb.MovePosS2C)
	//if c.RoleID != msg.RoleId {
	//	return
	//}
	//logger.Infof("玩家[%s]在场景[%d]移动,原坐标[%d,%d], 新坐标[%d,%d]成功：%v",
	//	c.Account, msg.MapId, c.Pos.X, c.Pos.Y, msg.Pos.X, msg.Pos.Y, msg)
	//c.Pos = model.Pos{X: int(msg.Pos.X), Y: int(msg.Pos.Y)}
	//mapConf := conf.MapConfs[c.Location.MapID]
	//if mapConf.Width > int(msg.Pos.X+10) && mapConf.Height > int(msg.Pos.Y+10) {
	//	c.MovePosC2S(msg.Pos.X+10, msg.Pos.Y+10)
	//}
}

func (c *Client) RoleViewListS2C(cmd uint32, rawMsg proto.Message) {
	msg := rawMsg.(*pb.RoleViewListS2C)
	if msg.Type == def.RoleViewTypeAll {
		logger.Debugf("玩家[%s]在场景[%d] 当前视野人数：%v", c.Account, msg.MapId, len(msg.SceneRoles))
	}
}

func (c *Client) HeartbeatS2C(cmd uint32, rawMsg proto.Message) {

}
