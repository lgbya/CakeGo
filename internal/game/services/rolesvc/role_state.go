package rolesvc

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"cake/internal/gensvc/rpcgen/rpcid"
)

type State struct {
	*model.Role
}

func newState(role *model.Role) *State {
	return &State{role}
}

// 创建副本
func (s *State) Copy(cmd string) (any, bool) {
	var isSkip bool
	switch cmd {
	case "TimerCheckHeartbeat":
		isSkip = true
	case rpcid.RpcHeartbeat:
		isSkip = true
	case rpcid.RpcMovePath:
		isSkip = true

	}
	if isSkip {
		return nil, false
	}
	return s.Role.CloneRolePO(), true
}

// 失败后数据恢复
func (s *State) Restore(rawData any) {
	rolePO, ok := rawData.(*model.RolePO)
	if !ok {
		return
	}
	s.Role.RolePO = rolePO
}

var _ rpc.GenState = new(State)
