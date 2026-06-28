package client

import (
	"math/rand"
)

func Rand(min, max int) int {
	if min >= max {
		return min
	}
	// 直接使用全局随机数（Go 1.20+ 自动种子化，内部已加锁，并发安全）
	return rand.Intn(max-min+1) + min
}
