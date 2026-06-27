package ws

import (
	"cake/internal/gate/conn/connsvc"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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

func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{conn: conn}
}

func (c *Conn) Read() ([]byte, error) {
	msgType, data, err := c.conn.ReadMessage()
	if msgType != websocket.BinaryMessage {
		return nil, errors.New("非BinaryMessage")
	}
	return data, err
}

func (c *Conn) Send(buf []byte) {
	defer sys.Recover("SendPacket")
	c.conn.WriteMessage(websocket.BinaryMessage, buf)
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

	// 心跳超时设置
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	//fmt.Println("新客户端接入:", conn.RemoteAddr())
	wsConn := NewConn(conn)
	connsvc.StartService(wsConn)
}

//func wsHandler(w http.ResponseWriter, r *http.Request) {
//	defer func() {
//		if err := recover(); err != nil {
//			log.Printf("捕获ws panic: %v", err)
//		}
//	}()
//
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println("握手失败:", err)
//		return
//	}
//	defer conn.Close()
//
//	// 心跳超时设置
//	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
//	conn.SetPongHandler(func(string) error {
//		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
//		return nil
//	})
//
//	log.Println("客户端连接成功")
//	var buffer []byte // 粘包缓冲区
//
//	for {
//		msgType, data, err := conn.ReadMessage()
//		// 只要出错 OR 消息类型=-1 直接退出循环
//		if err != nil || msgType == -1 {
//			log.Printf("连接关闭, msgType:%d, err:%v", msgType, err)
//			break
//		}
//
//		// 只处理二进制业务消息
//		switch msgType {
//		case websocket.BinaryMessage:
//			buffer = append(buffer, data...)
//			// 循环解包处理粘包
//			for {
//				cmd, body, left, ok := packet.DecodeMsg(buffer)
//				if !ok {
//					break
//				}
//				log.Printf("收到cmd:%d 包体长度:%d", cmd, len(body))
//				// 执行protobuf反序列化+业务路由
//				buffer = left
//			}
//		case websocket.TextMessage:
//			logger.Warn("丢弃非法文本消息")
//		}
//	}
//}
