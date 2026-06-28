package client

import (
	"math/rand"
	"time"
)

func init() {
	// 必须给独立随机源设置种子
	rand.Seed(time.Now().UnixNano())
}
func RandomInt(n int) int {
	return rand.Intn(n)
}
