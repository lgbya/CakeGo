package scene

import (
	"cake/internal/game/model"
	"strconv"
)

type State struct {
	sceneID    uint32
	sceneRoles map[uint64]*model.SceneRole
	connsPrt   *map[uint64]*model.Conn //广播进程只读，增和删在场景进程
}

func (s *State) ID() string {
	// uint64 转字符串作为唯一主键
	return strconv.Itoa(int(s.sceneID))
}

func newState(id uint32) *State {
	conns := make(map[uint64]*model.Conn)
	return &State{
		sceneID:    id,
		sceneRoles: make(map[uint64]*model.SceneRole),
		connsPrt:   &conns,
	}
}
