package model

import (
	consts2 "cake/internal/game/def"
	"cake/internal/gate/packet"
	"cake/internal/gensvc/rpc"
	"cake/internal/util/sys"
	"github.com/mohae/deepcopy"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync/atomic"
	"time"
)

type Heartbeat struct {
	LastTime int64 //最近一次心跳时间
	BadCnt   int   //连续错误的心跳次数
}

// 持久化数据
type RolePO struct {
	RoleID    uint64 `gorm:"primaryKey"`
	Account   string
	ServerID  uint32
	PlatID    uint32
	Name      string
	Gender    consts2.Gender //性别
	Career    consts2.Career //职业
	Lv        uint32         `gorm:"default:1"` //等级
	Data      RoleBizData    `gorm:"column:data;type:JSON;serializer:json"`
	CreatedAt time.Time
	//UpdateAt  time.Time
}

// 业务聚合数据 武器，背包，宠物都放在这里
type RoleBizData struct {
	Exp      uint32   `json:"exp"`
	Location Location `json:"location"`
}

type Location struct {
	SceneID uint32 `json:"scene_id"`
	//SceneSvcID int    `json:"scene_svc_id"`
	MapID   uint32 `json:"map_id"`
	Pos     Pos    `json:"pos"`
	OldPos  Pos    `json:"old_pos"`
	GridPos Pos    `json:"grid_pos"`
	Dir     int    `json:"dir"`
}

func (l *Location) UpdatePos(pos Pos) {
	oldPos := l.Pos
	l.Pos = pos
	l.OldPos = oldPos
}

type Role struct {
	*RolePO
	*Conn
	RoleRpc   *rpc.Service
	isSave    bool
	Heartbeat Heartbeat
}

func (r *RolePO) ID() string {
	return strconv.FormatUint(r.RoleID, 10)
}

func (r *RolePO) TableName() string {
	return "role"
}

func (r *Role) Save() {
	r.isSave = true
}
func (r *Role) RestSave() {
	r.isSave = false
}

func (r *Role) IsSave() bool {
	return r.isSave
}

func (r *Role) CloneRoleDB() *RolePO {
	return deepcopy.Copy(r).(*RolePO)
}

func (r *Role) Location() Location {
	return r.Data.Location
}
func (r *Role) SetLocation(location Location) {
	r.Data.Location = location
}

func (r *Role) SendSuccess(msg proto.Message) {
	data := packet.Success(msg)
	r.SendMsg(data)
}

func (r *Role) SendFail(msg proto.Message, errCode uint32) {
	data := packet.Fail(msg, errCode)
	r.SendMsg(data)
}

type Conn struct {
	ID       uint32
	MsgQueue chan proto.Message
	SceneRpc *rpc.Service
	StopFn   func()
	Closed   *atomic.Bool
}

func (c *Conn) SendMsg(msg proto.Message) {
	if c.IsClosed() {
		return
	}
	sys.SafeSend5s(c.MsgQueue, msg)

}

func (c *Conn) CloseConn() {
	if c.IsClosed() {
		return
	}
	c.StopFn()
	c.SceneRpc = nil
	c.MsgQueue = nil
	c.StopFn = nil
}

func (c *Conn) IsClosed() bool {
	return (*c.Closed).Load()
}
