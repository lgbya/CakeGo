package rolesvc

import (
	"cake/internal/game/logic/role"
	"cake/internal/game/model"
	"cake/internal/pkg/logger"
)

type logoutFn func(role *model.Role) error

var logoutFns = []logoutFn{
	role.Logic().LogoutLeave,
}

// 登录初始化玩家
func HandleLogout(role *model.Role) {
	for _, fn := range logoutFns {
		if err := fn(role); err != nil {
			logger.Errorf("[%d]玩家下线错误", role.RoleID)
			continue
		}
	}
}
