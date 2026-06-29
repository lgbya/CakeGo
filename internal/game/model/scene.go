package model

import (
	"cake/internal/game/def"
	"cake/internal/gensvc/rpc"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
	"sync"
)

const BattleSyncTypeMove = 1 //同步移动
const BattleSyncTypeAttr = 2 //同步属性

type SceneRole struct {
	RoleID   uint64
	RoleName string
	ServerID uint32
	PlatID   uint32
	Career   def.Career
	Location
	*Conn
	mu      sync.RWMutex
	RoleRpc *rpc.Service
}

func NewSceneRole(roleState *Role) *SceneRole {
	return &SceneRole{
		RoleID:   roleState.RoleID,
		Location: roleState.Data.Location,
		Conn:     roleState.Conn,
		RoleRpc:  roleState.RoleRpc,
		RoleName: roleState.Name,
		ServerID: roleState.ServerID,
		PlatID:   roleState.PlatID,
		Career:   roleState.Career,
	}
}

func (s *SceneRole) Pb() *pb.SceneRole {
	return &pb.SceneRole{
		RoleId:   s.RoleID,
		RoleName: s.RoleName,
		ServerId: s.ServerID,
		PlatId:   s.PlatID,
		Pos:      s.Pos.Pb(),
		Career:   uint32(s.Career),
	}
}

func (s *SceneRole) ToBattle() *BattleRole {
	return &BattleRole{
		RoleID: s.RoleID,
		Pos:    s.Pos,
	}
}

func (s *SceneRole) SyncBattle(battleRole *BattleRole) {
	s.UpdatePos(battleRole.Pos)
}

type BattleRole struct {
	RoleID   uint64
	Pos      Pos
	SyncType int //同步类型
}

func (b *BattleRole) SyncTypeMove() {
	b.SyncType = BattleSyncTypeMove
}

type AoiInfo struct {
	RoleIDs map[uint64]struct{}
	Msg     proto.Message
}
