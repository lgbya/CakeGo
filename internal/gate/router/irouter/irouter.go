package irouter

import (
	"cake/internal/game/model"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"cake/internal/gate/igate"
	"google.golang.org/protobuf/proto"
	"sync"
)

type IRoute interface {
	Register()
}

type RoleCmd struct {
	RoleID  uint64
	Msg     proto.Message
	RoleFn  RoleRouteFn
	SceneFn SceneRouteFn
}

type ConnRouteFn = func(igate.ConnSvc, proto.Message)
type RoleRouteFn = func(*model.Role, proto.Message) error
type SceneRouteFn = func(iscene.IService, *model.SceneRole, proto.Message) error

type IRegister interface {
	ConnCmd(proto.Message, ConnRouteFn)
	RoleCmd(proto.Message, RoleRouteFn)
	SceneCmd(proto.Message, SceneRouteFn)
	Dispatch(any, []byte)
}

var (
	defaultReg IRegister
	regOnce    sync.Once
)

func Reg() IRegister {
	return defaultReg
}

func InitReg(ireg IRegister) {
	regOnce.Do(func() {
		defaultReg = ireg
	})
}
