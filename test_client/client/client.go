package client

import (
	"cake/internal/conf"
	"cake/internal/game/model"
	"cake/internal/gate/packet"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"cake/proto/pb"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Client struct {
	Account string
	RoleID  uint64
	conn    net.Conn
	model.Location
	cbFnMap        map[uint32]CbFn
	autoWalkTicker *time.Ticker
	autoWalkStopCh chan struct{}
}

var wg sync.WaitGroup

func NewClient(max int) {
	for i := 0; i < max; i++ {

		account := "user" + strconv.Itoa(i)
		wg.Add(1)
		sys.SafeGo(func() {
			client := &Client{Account: account}
			client.regCbFn()
			client.start()
		})
	}
	wg.Wait()
}

func (c *Client) start() {
	defer wg.Done()

	//=============================
	// 连接服务端
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		fmt.Println("连接失败：", err)
		return
	}
	defer conn.Close()

	//fmt.Println("连接服务端成功")
	c.conn = conn
	c.AccountAuthC2S()
	c.readLoop(conn)
	// 读取控制台输入
	//scanner := bufio.NewScanner(os.Stdin)
	//for {
	//	fmt.Print("请输入消息：")
	//	if !scanner.Scan() {
	//		return
	//	}
	//	//c.inputHandle(conn, scanner)
	//}
}

//func (c *Client) inputHandle(conn net.Conn, scanner *bufio.Scanner) {
//	input := scanner.Text()
//	code, errx := strconv.ParseUint(input, 10, 32)
//	if errx != nil {
//		fmt.Println("输入非法，请输入纯数字协议号！")
//		return
//	}
//
//	if c.send(conn, uint32(code)) == false {
//		return
//	}
//}

func (c *Client) send(msg proto.Message) bool {
	code, _, _ := pb.GetC2SCmdByMsg(msg)
	sendBuf, _ := packet.EncodeMsg(msg)
	if _, err := c.conn.Write(sendBuf); err != nil {
		fmt.Println("发送失败:", err)
		return false
	}
	if logger.CheckLogCmd(code) {
		logger.Debugf("%s:发送:%d %v", c.Account, code, msg)
	}
	return true
}

func (c *Client) readLoop(conn net.Conn) {
	defer conn.Close()
	// 缓存上一次未处理完的粘包剩余数据
	var remainBuf []byte
	tempBuf := make([]byte, 1024)

	for {
		n, err := conn.Read(tempBuf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				logger.Errorf("读取服务端回复失败:%v", err)
			}

			return
		}
		// 拼接：上次残留数据 + 本次新读到的数据
		allData := append(remainBuf, tempBuf[:n]...)
		// 清空残留，本次循环解析后重新赋值
		remainBuf = nil

		// 循环解析，直到缓冲区不够一个完整包
		for {
			respCmd, respBody, left, ok := packet.DecodeMsg(allData)
			if !ok {
				// 数据不足一个完整包，把剩余数据存起来，下次继续解析
				remainBuf = left
				break
			}

			// 解析成功，执行业务回调
			cbFn, okCb := c.cbFnMap[respCmd]
			if !okCb {
				logger.Errorf("服务端数据解包失败:未注册指令 %d", respCmd)
				// 把剩下的数据继续循环解析
				allData = left
				continue
			}

			if err := proto.Unmarshal(respBody, cbFn.msg); err != nil {
				logger.Errorf("protobuf反序列化失败 cmd:%d, err:%v", respCmd, err)
				allData = left
				continue
			}

			if logger.CheckLogCmd(respCmd) {
				logger.Debugf("%s:接收:%d %v", c.Account, respCmd, cbFn.msg)
			}
			if cbFn.fn != nil {
				cbFn.fn(respCmd, cbFn.msg)
			}

			// 剩余数据继续循环拆包（处理一次读到多个包的粘包场景）
			allData = left
		}
	}
}

func (c *Client) StartAutoWalk() {
	// 防止重复启动定时器
	if c.autoWalkTicker != nil {
		return
	}
	// 初始化1秒定时器
	c.autoWalkTicker = time.NewTicker(1 * time.Second)
	c.autoWalkStopCh = make(chan struct{})

	go func() {
		defer func() {
			c.autoWalkTicker.Stop()
			c.autoWalkTicker = nil
			close(c.autoWalkStopCh)
			c.autoWalkStopCh = nil
			logger.Infof("玩家[%s]自动行走任务已停止", c.Account)
		}()

		mapConf := conf.MapConfs[c.Location.MapID]
		for {
			select {
			case <-c.autoWalkStopCh:
				// 收到停止信号，退出协程
				return
			case <-c.autoWalkTicker.C:
				// 每秒触发：随机生成坐标
				maxX := int32(mapConf.Width - 1)
				maxY := int32(mapConf.Height - 1)
				// 在地图范围内随机坐标
				randomX := rand.Int31n(maxX)
				randomY := rand.Int31n(maxY)

				// 发送客户端上报移动协议
				c.MovePosC2S(uint32(randomX), uint32(randomY))
				//logger.Infof("玩家[%s]自动行走：随机坐标(%d,%d)", c.Account, randomX, randomY)
			}
		}
	}()
	//logger.Infof("玩家[%s]开启每秒自动随机行走", c.Account)
}
