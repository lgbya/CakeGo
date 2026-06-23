// Code generated testsvc.go; DO NOT EDIT
package rpcgen
import (
	"cake/internal/gensvc/router"
	"cake/internal/game/services/testsvc"
	"errors"
)
var TestsvcMap = map[string]router.RpcFn{
	"RpcTest": Testsvc_RpcTest,

}

func Testsvc_RpcTest(rawSvc, rawState, rawArgs any) (any, error) {
		svc := rawSvc.(*testsvc.Service)
		state,ok := rawState.(*testsvc.State)
		if !ok {
			return nil, errors.New("invalid state")
		}
		res, errx := svc.RpcTest(state, rawArgs)
		if errx != nil {
			return nil, errx
		}
		return res, nil
}

