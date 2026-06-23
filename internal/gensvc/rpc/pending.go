package rpc

import (
	"cake/internal/pkg/logger"
	"sync"
)

type pending struct {
	list []*Msg
	head int // 队头下标
	tail int // 队尾下标
	mu   sync.Mutex
}

func newPending() *pending {
	return &pending{list: make([]*Msg, 10)}
}

func (p *pending) pushLocked(name string, msg *Msg) {
	//超过最大上限，直接抛弃
	if p.tail >= PendingMaxLen {
		logger.Errorf("[%s][%s]进程消息队列堵塞超标,直接抛弃,当前积压:%d", name, msg.Cmd, p.pendingLen())
		return
	}

	//动态扩容
	if p.tail >= len(p.list) {
		newArr := make([]*Msg, len(p.list)*2)
		copy(newArr, p.list[p.head:p.tail])
		p.list = newArr
		p.tail = p.pendingLen()
		p.head = 0
	}
	p.list[p.tail] = msg
	p.tail++
	// 只在堆积达到警戒阈值时打印告警，避免日志刷屏
	if p.pendingLen() == PendingMaxLen/2 {
		logger.Errorf("[%s][%s]进程消息队列出现消息堵塞,当前积压:%d", name, msg.Cmd, p.pendingLen())
	}
}

func (p *pending) popLocked() (*Msg, bool) {
	if p.pendingLen() == 0 {
		var empty *Msg
		return empty, false
	}
	msg := p.list[p.head]
	// 置空引用，防止内存泄漏
	p.list[p.head] = nil
	p.head++

	// 优化：队头偏移超过一半，原地缩容，释放空闲内存
	if p.head > len(p.list)/2 {
		copy(p.list, p.list[p.head:p.tail])
		p.tail = p.pendingLen()
		p.head = 0
	}
	return msg, true
}
func (p *pending) pendingLen() int {
	return p.tail - p.head
}

func (p *pending) getBatchMsgs(sendLen int, sendMaxCap int) ([]*Msg, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	freeLen := sendMaxCap - sendLen
	if freeLen <= 0 {
		return nil, 0
	}
	pendingCnt := p.pendingLen()
	if pendingCnt == 0 {
		return nil, 0
	}

	batchCount := min(freeLen, pendingCnt)
	// 批量取出待发送消息，锁内只做一次拷贝，快速释放锁
	batchMsgs := make([]*Msg, 0, batchCount)
	for i := 0; i < batchCount; i++ {
		msg, ok := p.popLocked()
		if !ok {
			break
		}
		batchMsgs = append(batchMsgs, msg)
	}
	return batchMsgs, len(batchMsgs)
}

func (p *pending) Clear(s *Service) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, msg := range p.list {
		s.PutMsg(msg)
	}
	p.list = nil
}
