package rolesvc

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
)

type State struct {
	*model.Role
}

func newState(role *model.Role) *State {
	return &State{role}
}

// 创建副本
func (s *State) Copy() any {
	return s.Role.CloneRolePO()
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
