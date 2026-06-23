package timer

import (
	"cake/internal/util/sys"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"time"
)

type timerFn func(any, any) error

type timeTask struct {
	name      string        //定时器id
	maxTimes  int           //最大调用次数 -1无限
	runTimes  int           //已经调用次数
	nextRunAt time.Time     //下一次调用时间
	interval  time.Duration //间隔多少秒执行
	fn        timerFn
	args      any
}

// ---------------------- 定时器管理器 ----------------------

type GbTree struct {
	tree *treemap.Map // key: nextRunAt.UnixNano(), value: *timeTask
}

func NewGbTree() *GbTree {
	return &GbTree{
		tree: treemap.NewWith(utils.Int64Comparator),
	}
}

// AddTimer 注册定时任务
// name:唯一标识；interval:间隔；maxTimes:-1无限；fn:回调
func (tt *GbTree) AddTimer(name string, interval time.Duration, maxTimes int, fn timerFn, args any) {
	now := time.Now()

	t := &timeTask{
		name:      name,
		maxTimes:  maxTimes,
		runTimes:  0,
		interval:  interval,
		nextRunAt: now.Add(interval),
		fn:        fn,
		args:      args,
	}
	// 按下次执行时间排序存入红黑树
	tt.tree.Put(t.nextRunAt.UnixNano(), t)
}

// RemoveTimer 按名字移除任务
func (tt *GbTree) RemoveTimer(name string) {
	it := tt.tree.Iterator()
	for it.Next() {
		t, ok := it.Value().(*timeTask)
		if !ok {
			continue
		}
		if t.name == name {
			tt.tree.Remove(it.Key())
			return
		}
	}
}

// Tick 每帧/每轮询调用一次，执行到期任务
func (tt *GbTree) Tick(s IGenService) {
	defer sys.Recover(s.Name())

	now := time.Now()
	nowNano := now.UnixNano()

	// 循环取最前面到期的
	minKey, minVal := tt.tree.Min()
	if minVal == nil {
		tt.tree.Remove(minKey)
		return
	}

	t, ok := minVal.(*timeTask)
	if !ok {
		tt.tree.Remove(minKey)
		return
	}

	// 还没到期，结束本轮
	if t.nextRunAt.UnixNano() > nowNano {
		return
	}

	//先移除旧节点，再执行
	tt.tree.Remove(minKey)

	// 执行回调

	if err := t.fn(s.State(), t.args); err != nil {
		return
	}
	t.runTimes++

	// 判断是否继续
	if t.maxTimes != -1 && t.maxTimes >= t.runTimes {
		return
	}
	// 计算下次执行时间（避免漂移：基于上次nextRunAt，非now）
	t.nextRunAt = t.nextRunAt.Add(t.interval)
	tt.tree.Put(t.nextRunAt.UnixNano(), t)
}

func (tt *GbTree) Close(s IGenService) {
	UnRegister(s)
	tt.tree.Clear()
}

// ---------------------- 角色进程示例 ----------------------

//func main() {
//	tm := NewGbTree()
//
//	// 注册三个任务：10s、5s、3s，都是无限循环
//	tm.AddTimer("role_10s", 10*time.Second, -1, func() {
//		fmt.Println("[10s] 执行：", time.Now().Format("15:04:05"))
//	})
//	tm.AddTimer("role_5s", 5*time.Second, -1, func() {
//		fmt.Println("[5s] 执行：", time.Now().Format("15:04:05"))
//	})
//	tm.AddTimer("role_3s", 3*time.Second, -1, func() {
//		fmt.Println("[3s] 执行：", time.Now().Format("15:04:05"))
//	})
//
//	// 模拟角色进程主循环（10ms tick一次）
//	timer := time.NewTicker(10 * time.Millisecond)
//	defer timer.StopFn()
//	for range timer.C {
//		tm.tick()
//	}
//}
