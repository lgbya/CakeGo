package client

import (
	"cake/env"
	consts2 "cake/internal/game/def"
	"cake/internal/game/logic/sdk"
	"cake/proto/pb"
	"time"
)

func (c *Client) AccountAuthC2S() {
	c.send(&pb.AccountAuthC2S{
		Account:   c.Account,
		ChannelId: sdk.ChannelIdTemp,
	})
}

func (c *Client) SelectRolesC2S() {
	c.send(&pb.SelectRolesC2S{
		Account: c.Account,
	})
}

func (c *Client) CreateRoleC2S() {
	gender := Rand(int(consts2.GenderWoman), int(consts2.GenderMan))
	career := Rand(int(consts2.CareerWarrior), int(consts2.CareerPriest))
	c.send(&pb.CreateRoleC2S{
		Name:     "新手玩家" + c.Account,
		ServerId: env.ServerID(),
		PlatId:   env.PlatID(),
		Gender:   uint32(gender),
		Career:   uint32(career),
	})
}

func (c *Client) LoginRoleC2S() {
	c.send(&pb.LoginRoleC2S{
		RoleId: c.RoleID,
	})
}

func (c *Client) MovePosC2S(x, y uint32) {
	c.send(&pb.MovePosC2S{
		Pos: &pb.Pos{
			X: x,
			Y: y,
		},
	})
}

func (c *Client) LoginEnterC2S() {
	c.send(&pb.LoginEnterC2S{})
}

func (c *Client) EnterSceneC2S(sceneID, mapID uint32) {
	c.send(&pb.EnterSceneC2S{SceneId: sceneID, MapId: mapID})
}

func (c *Client) HeartbeatC2S() {
	ticker := time.NewTicker(consts2.HeartbeatInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.send(&pb.HeartbeatC2S{ClientTime: time.Now().Unix()})
		}
	}
}
