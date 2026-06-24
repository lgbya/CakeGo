package db

import (
	"cake/env"
	zlogger "cake/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
)

var cacheInst *redis.Client

func CacheInst() *redis.Client {
	return cacheInst
}

func InitRedis() {
	addr := env.GetString("redis.addr")
	cacheInst = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if cacheInst == nil {
		panic("redis init failed")
	}
	zlogger.Info("success	redis 连接成功")
}
