package rolesvc

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
)

type State struct {
	*model.Role
}

func newState(role *model.Role) *State {
	return &State{role}
}

// 创建副本
func (s *State) Copy(cmd string) (any, bool) {
	// 跳过不需要保存的命令
	skipCmds := map[string]bool{
		"TimerCheckHeartbeat": true,
		rpcid.RpcHeartbeat:    true,
		rpcid.RpcMovePath:     true,
	}
	if skipCmds[cmd] {
		return nil, false
	}
	//性能上不用deepcopy，自己手动拷贝
	data := model.RoleBizData{
		Exp:      s.Data.Exp,
		Location: s.Location(),
	}
	clone := &model.RolePO{
		RoleID:   s.RolePO.RoleID,
		Account:  s.RolePO.Account,
		ServerID: s.RolePO.ServerID,
		PlatID:   s.RolePO.PlatID,
		Name:     s.RolePO.Name,
		Gender:   s.RolePO.Gender,
		Career:   s.RolePO.Career,
		Lv:       s.RolePO.Lv,
		Data:     data,
	}

	return clone, true
}

// 失败后数据恢复
func (s *State) Restore(rawData any) {
	clone, ok := rawData.(*model.RolePO)
	if !ok {
		logger.Errorf("恢复角色数据失败: 数据类型错误")
		return
	}
	s.Role.RolePO = clone
}

var _ rpc.GenState = new(State)
