package battle

import "cake/internal/game/model"

type State struct {
	battleRoles  map[uint64]model.BattleRole
	dirtyRoleIDs map[uint64]struct{}
}

func newState() *State {
	return &State{battleRoles: make(map[uint64]model.BattleRole)}
}

func (s *State) SetDirty(roleID uint64) {
	s.dirtyRoleIDs[roleID] = struct{}{}
}

func (s *State) SaveBattleRole(battleRole model.BattleRole) {
	s.battleRoles[battleRole.RoleID] = battleRole

}
