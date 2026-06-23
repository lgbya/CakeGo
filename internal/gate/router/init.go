package router

import (
	"cake/internal/gate/router/irouter"
)

func Init() {
	irouter.InitReg(NewRouter())
	for _, route := range Routes() {
		route.Register()
	}
}
