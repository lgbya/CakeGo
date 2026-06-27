package uuid

import (
	"cake/env"
	"cake/internal/game/def"
	"fmt"
	"sync/atomic"
)

// 位分配常量
const (
	typeBits   = 4
	platBits   = 8
	serverBits = 16
	seqBits    = 36

	typeShift   = platBits + serverBits + seqBits // 60
	platShift   = serverBits + seqBits            // 52
	serverShift = seqBits                         // 36

	typeMask   = 0x0F
	platMask   = 0xFF
	serverMask = 0xFFFF
	seqMask    = 0xFFFFFFFFF
)

// 业务ID类型

// GenID 外部传入序列号指针，多类型全局唯一ID生成
// typ: 业务类型 0~15
// serverID: 服务器ID
// platID: 平台ID
// nextSeq: 外部原子序列号指针
func GenID(typ uint64, nextSeq *uint64) uint64 {
	// 边界校验
	if typ > typeMask {
		panic(fmt.Sprintf("type %d exceeds max %d", typ, typeMask))
	}
	platID := uint64(env.PlatID())
	if platID > platMask {
		panic(fmt.Sprintf("platID %d exceeds max %d", platID, platMask))
	}
	serverID := uint64(env.ServerID())
	if serverID > serverMask {
		panic(fmt.Sprintf("serverID %d exceeds max %d", serverID, serverMask))
	}

	seq := atomic.AddUint64(nextSeq, 1)
	if seq > seqMask {
		panic(fmt.Sprintf("seq %d overflow, max %d", seq, seqMask))
	}

	t := typ & typeMask
	p := platID & platMask
	s := serverID & serverMask

	return t<<typeShift | p<<platShift | s<<serverShift | seq
}

// ParseID 解析ID
func ParseID(id uint64) (typ, serverID, platID, seq uint64) {
	typ = (id >> typeShift) & typeMask
	platID = (id >> platShift) & platMask
	serverID = (id >> serverShift) & serverMask
	seq = id & seqMask
	return
}

func ParseRoleID(roleID uint64) (serverID, platID, seq uint64) {
	typ, s, p, seq := ParseID(roleID)
	if typ != def.IDTypeRole {
		panic("current id is not role type")
	}
	return s, p, seq
}
