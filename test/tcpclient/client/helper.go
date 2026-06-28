package client

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r   *rand.Rand
	mux sync.Mutex
)

func init() {
	// 必须给独立随机源设置种子
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Rand(min int, max int) int {
	if min >= max {
		return min
	}
	mux.Lock()
	defer mux.Unlock()
	return r.Intn(max-min+1) + min
}
