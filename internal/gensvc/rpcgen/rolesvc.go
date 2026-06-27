// Code generated rolesvc.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/rolesvc"
	"errors"
)
var RolesvcMap = map[string]router.RpcFn{
	"RpcRoleCmd": Rolesvc_RpcRoleCmd,
	"RpcHeartbeat": Rolesvc_RpcHeartbeat,
	"RpcSaveSceneRole": Rolesvc_RpcSaveSceneRole,
	"RpcUpdateConn": Rolesvc_RpcUpdateConn,
	"RpcConnClose": Rolesvc_RpcConnClose,

}

func Rolesvc_RpcRoleCmd(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*rolesvc.Service)
		state,ok := rawState.(*rolesvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcRoleCmd(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Rolesvc_RpcHeartbeat(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*rolesvc.Service)
		state,ok := rawState.(*rolesvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcHeartbeat(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Rolesvc_RpcSaveSceneRole(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*rolesvc.Service)
		state,ok := rawState.(*rolesvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcSaveSceneRole(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Rolesvc_RpcUpdateConn(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*rolesvc.Service)
		state,ok := rawState.(*rolesvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcUpdateConn(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Rolesvc_RpcConnClose(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*rolesvc.Service)
		state,ok := rawState.(*rolesvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcConnClose(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

