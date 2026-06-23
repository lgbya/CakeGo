// Code generated scene.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/mapsvc/scene"
	"errors"
)
var SceneMap = map[string]router.RpcFn{
	"RpcRoleCmd": Scene_RpcRoleCmd,
	"RpcEnterScene": Scene_RpcEnterScene,
	"RpcLeaveScene": Scene_RpcLeaveScene,
	"RpcSyncRoleStates": Scene_RpcSyncRoleStates,

}

func Scene_RpcRoleCmd(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*scene.Service)
		state,ok := rawState.(*scene.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcRoleCmd(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Scene_RpcEnterScene(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*scene.Service)
		state,ok := rawState.(*scene.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcEnterScene(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Scene_RpcLeaveScene(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*scene.Service)
		state,ok := rawState.(*scene.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcLeaveScene(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Scene_RpcSyncRoleStates(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*scene.Service)
		state,ok := rawState.(*scene.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcSyncRoleStates(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

