package router

import (
	"cake/internal/game/handler"
	"cake/internal/gate/router/irouter"
)

// Routes 自动生成路由注册列表，请勿手动修改
func Routes() []irouter.IRoute {
	return []irouter.IRoute{
		&handler.LoginRoute{},
		&handler.SceneRoute{},

	}
}
