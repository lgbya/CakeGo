package ws

import (
	"cake/internal/gate/packet"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// 升级配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 跨域放行，H5调试用，生产环境按需限制域名
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Conn struct {
	conn *websocket.Conn
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.conn.ReadMessage()
}

func (c *Conn) Send(buf []byte) {
	defer sys.Recover("SendPacket")

}

func (c *Conn) Close(id uint32) error {
	err := c.conn.Close()
	if err != nil {
		logger.Errorf("[%d]关闭网络错误：%v", id, err)
		return err
	}
	return nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// 兜底捕获panic，单个客户端异常不会让整个服务崩溃
	defer sys.Recover("wsHandler")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("握手失败:", err)
		return
	}
	defer conn.Close()

	log.Println("客户端连接成功")

	var buf []byte // 粘包缓冲区
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			// 读取失败立刻跳出循环，终止当前连接
			log.Printf("连接读取异常，关闭连接: %v", err)
			break
		}
		// 只处理二进制消息，过滤文本、心跳控制帧
		if msgType != websocket.BinaryMessage {
			continue
		}

		// 追加到粘包缓冲区
		buf = append(buf, data...)
		// 循环解包处理粘包
		for {
			code, body, left, ok := packet.DecodeMsg(buf)
			if !ok {
				break
			}
			log.Printf("收到数据: %x %d", body, code)
			buf = left
			// 这里执行protobuf解析、业务路由
		}
	}
}
