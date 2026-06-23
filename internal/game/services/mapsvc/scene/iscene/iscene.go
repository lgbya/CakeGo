package iscene

import (
	"cake/internal/gensvc/rpc"
	"sync"
)

type IManager interface {
	RpcBySceneID(uint32) *rpc.Service
	RpcByMapID(uint32) *rpc.Service
	MapIdToSceneId(uint32) uint32
}

var (
	defaultIMgr IManager
	imgrOne     sync.Once
)

type IService interface {
	GetBattleRpc() *rpc.Service
}

func Manager() IManager {
	return defaultIMgr
}

func InitManager(imgr IManager) {
	imgrOne.Do(func() {
		defaultIMgr = imgr
	})
}
