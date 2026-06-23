package rolesvc

import (
	"cake/internal/game/logic/role"
	"cake/internal/game/model"
)

type loginFn func(role *model.Role) error

var loginFns = []loginFn{
	role.Logic().LoginEnter,
}

// 登录初始化玩家
func HandleLogin(roleState *model.Role) error {
	for _, fn := range loginFns {
		if err := fn(roleState); err != nil {
			return err
		}
	}
	return nil
}
