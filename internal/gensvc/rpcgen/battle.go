// Code generated battle.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/mapsvc/battle"
	"errors"
)
var BattleMap = map[string]router.RpcFn{
	"RpcAddBattleRole": Battle_RpcAddBattleRole,
	"RpcDelBattleRole": Battle_RpcDelBattleRole,
	"RpcMovePath": Battle_RpcMovePath,
	"RpcTest": Battle_RpcTest,

}

func Battle_RpcAddBattleRole(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*battle.Service)
		state,ok := rawState.(*battle.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcAddBattleRole(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Battle_RpcDelBattleRole(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*battle.Service)
		state,ok := rawState.(*battle.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcDelBattleRole(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Battle_RpcMovePath(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*battle.Service)
		state,ok := rawState.(*battle.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcMovePath(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Battle_RpcTest(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*battle.Service)
		state,ok := rawState.(*battle.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcTest(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

