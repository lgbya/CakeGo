package bcast

import "cake/internal/game/model"

type State struct {
	connRoles map[uint64]*model.Conn
}

func newState() *State {
	return &State{connRoles: make(map[uint64]*model.Conn)}
}
