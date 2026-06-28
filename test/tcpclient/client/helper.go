package client

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r    *rand.Rand
	mux  sync.Mutex
	once sync.Once
)

func init() {
	once.Do(func() {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

func Rand(min, max int) int {
	if min >= max {
		return min
	}
	mux.Lock()
	defer mux.Unlock()
	// 如果 r 为 nil，再次初始化（安全）
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return r.Intn(max-min+1) + min
}
