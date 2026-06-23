// Code generated rpcgen.go; DO NOT EDIT
package rpcgen	

import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/mapsvc/battle"
	"cake/internal/game/services/mapsvc/bcast"
	"cake/internal/game/services/mapsvc/scene"
	"cake/internal/game/services/rolesvc"
	"cake/internal/game/services/testsvc"

	"reflect"
)
	
var regRoutes = map[reflect.Type]map[string]router.RpcFn{
	reflect.TypeOf((*battle.Service)(nil)):BattleMap,
	reflect.TypeOf((*bcast.Service)(nil)):BcastMap,
	reflect.TypeOf((*scene.Service)(nil)):SceneMap,
	reflect.TypeOf((*rolesvc.Service)(nil)):RolesvcMap,
	reflect.TypeOf((*testsvc.Service)(nil)):TestsvcMap,

}

func Init() {
	router.Gen = NewGenRouter()
}

type GenRouter struct {
}

func NewGenRouter() router.Router {
	return &GenRouter{}
}

func (*GenRouter) GetRoutes(typ reflect.Type) (map[string]router.RpcFn, bool) {
	routes, ok := regRoutes[typ]
	if !ok {
		return nil, false
	}
	return routes, true
}

	