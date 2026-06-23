package role

import "sync"

var (
	logicInst *logic
	logicOnce sync.Once
)

type logic struct {
}

func Logic() *logic {
	logicOnce.Do(func() {
		logicInst = &logic{}
	})
	return logicInst
}
