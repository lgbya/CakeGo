package packet

import (
	"cake/internal/game/def/errcode"
	"cake/internal/util/errx"
	"cake/proto/pb"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync"
)

const MaxPacketBufSize = 4096

var bufPool = sync.Pool{
	New: func() interface{} {
		// 预分配容量，初始长度0
		return make([]byte, 0, MaxPacketBufSize)
	},
}

// GetBuf 获取缓冲区
func GetBuf() []byte {
	buf := bufPool.Get().([]byte)
	// 关键：重置长度，清空旧数据，容量保留不销毁
	return buf[:0]
}

// PutBuf 归还缓冲区
func PutBuf(buf []byte) {
	// 超出预设最大容量的直接丢弃，防止超大内存缓存不释放
	if cap(buf) > MaxPacketBufSize {
		return
	}
	bufPool.Put(buf[:0])
}

func Success(body proto.Message) proto.Message {
	return build(body, &pb.Notice{ErrCode: errcode.Ok})
}

func Fail(body proto.Message, errCode uint32) proto.Message {
	return build(body, &pb.Notice{ErrCode: errCode})
}

// DecodeMsg 解包
func DecodeMsg(data []byte) (code uint32, body []byte, left []byte, ok bool) {
	if len(data) < 8 {
		return 0, nil, data, false // 包头不够，等待后续数据
	}
	code = binary.BigEndian.Uint32(data[:4])
	bodyLen := binary.BigEndian.Uint32(data[4:8])
	total := 8 + int(bodyLen)
	if len(data) < total {
		return 0, nil, data, false // 包体不全，继续缓存
	}
	body = data[8:total]
	left = data[total:] // 多余字节留作下次解析
	return code, body, left, true
}

// EncodeMsg 封包：code + bodyLen + body
func EncodeMsg(msg proto.Message) ([]byte, error) {

	code, _, ok := pb.GetAllCmdByMsg(msg)
	if !ok {
		return nil, errx.New(errcode.EncodeFail)
	}

	body, err := proto.Marshal(msg)
	//fmt.Println(code)
	if err != nil {
		return nil, err
	}
	bodyLen := uint32(len(body))

	buf := make([]byte, 8+len(body))
	// 写入code
	binary.BigEndian.PutUint32(buf[0:4], code)
	// 写入body长度
	binary.BigEndian.PutUint32(buf[4:8], bodyLen)
	// 拷贝body
	copy(buf[8:], body)
	return buf, nil
}

// Build 自动把CommonNotice塞到结构体的Resp字段
func build(data proto.Message, cr *pb.Notice) proto.Message {
	val := reflect.ValueOf(data).Elem()
	f := val.FieldByName("CommonNotice")
	if f.CanSet() {
		f.Set(reflect.ValueOf(cr))
	}
	return data
}
