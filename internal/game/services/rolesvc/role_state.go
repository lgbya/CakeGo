package rolesvc

import "cake/internal/game/model"

type State struct {
	*model.Role
}

func newState(role *model.Role) *State {
	return &State{role}
}
