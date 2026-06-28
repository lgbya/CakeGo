// Code generated bcast.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/mapsvc/bcast"
	"errors"
)
var BcastMap = map[string]router.RpcFn{
	"RpcSaveConnRole": Bcast_RpcSaveConnRole,
	"RpcDelConnRole": Bcast_RpcDelConnRole,
	"RpcAoiNiceGrid": Bcast_RpcAoiNiceGrid,

}

func Bcast_RpcSaveConnRole(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*bcast.Service)
		state,ok := rawState.(*bcast.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcSaveConnRole(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Bcast_RpcDelConnRole(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*bcast.Service)
		state,ok := rawState.(*bcast.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcDelConnRole(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

func Bcast_RpcAoiNiceGrid(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*bcast.Service)
		state,ok := rawState.(*bcast.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcAoiNiceGrid(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

