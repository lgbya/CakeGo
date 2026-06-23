package router

import "reflect"

type RpcFn func(any, any, any) (any, error)

type Router interface {
	GetRoutes(reflect.Type) (map[string]RpcFn, bool)
}

var Gen Router
