# CakeGo 游戏服务框架
## 借鉴erlang的gen server实现进程单游戏服Actor 模型的mmo游戏框架

## 一、项目简介
基于 Golang 协程实现轻量 Actor 模型，借鉴 Erlang 网游成熟进程架构，网关+玩家+场景多Actor隔离。

## 二、目录结构
```text
CakeGo/
├── cmd                     # 服务启动入口
│   └── server
│       └── main.go         # 程序入口：初始化配置、启动网关、启动所有游戏Actor服务
├── env                     # 环境配置
├── internal                # 内部业务代码（禁止外部导入）
│   ├── conf                # 配置表解析、导表工具（Excel转json/二进制配置）
│   ├── game                # 游戏核心业务逻辑
│   │   ├── def             # 游戏通用常量、枚举、错误码
│   │   ├── handle          # 协议路由层：客户端协议分发、参数校验
│   │   │   ├── login_handle.go
│   │   │   └── scene_hanele.go
│   │   ├── logic           # 业务逻辑层（纯业务，不碰网络、存储）
│   │   ├── model           # 游戏公用数据结构、内存实体
│   │   ├── repo            # 数据持久层封装（mysql/redis）
│   │   └── services        # 基于gensvc实现的各个Actor协程服务（核心Actor层）协程文件后缀必须带_service
│   ├── gate                # 网关服务
│   ├── gensvc              # 仿Erlang gen_server 核心RPC/Actor框架实现
│   ├── pkg                 # 内部封装第三方依赖（避免直接go mod到处引用）
│   └── util                # 通用工具库
├── proto                   # Protobuf协议定义文件
├── scripts                 # Shell运维脚本
├── sql                     # MySQL建表语句、初始化数据
└── test_client             # TCP压测客户端、模拟多玩家登录测试

```
## 三、环境配置
环境配置在env目录下的app.yaml
````yaml
gate:
  addr: "0.0.0.0:8888"  #网关端口

base:
  platId : 1 #平台id
  serverId: 1 #服务id

#监控
monitoring:
  metricAddr: ":9091"	#开启metric监控的端口
  pprofAddr: ":6060"	#开启pprof监控的端口

db:
  host: "127.0.0.1"		#数据库host
  port: "3306"			#数据库端口
  user: "root"			#数据库账号
  pass: "123456"		#数据库密码
  name: "game_db"		#数据库名

````
## 四、运行命令
```bash
#安装命令
#注意事项如果协议生成错误，确认google/protobuf的位置，修改scripts/proto.sh的/usr/local/include为你的路径
#需要安装proto3
#sh run.sh install 平台号 服务器号
例：sh run.sh install 1 1

#启动命令
#sh run.sh start 平台号 服务器号
例：sh run.sh start 1 1

#关闭命令
#sh run.sh stop 平台号 服务器号
例：sh run.sh stop 1 1

#压测客户端
#sh run.sh test 平台号 服务器号
例：sh run.sh test 1 1

#单独编译协议
#sh run.sh proto 
例：sh run.sh proto

```
# 五、绑定协议路由
### 1. 绑定网关的路由
```go
irouter.Reg().ConnCmd(协议结构体, 绑定方法)
//例子
irouter.Reg().ConnCmd(&pb.HeartbeatC2S{}, r.HeartbeatC2S)
```

### 2. 绑定角色的路由
```go
irouter.Reg().RoleCmd(协议结构体, 绑定方法)
//例子
irouter.Reg().RoleCmd(&pb.HeartbeatC2S{}, r.HeartbeatC2S)
```

### 3. 绑定场景的路由
```go
irouter.Reg().SceneCmd(协议结构体, 绑定方法)
//例子
irouter.Reg().SceneCmd(&pb.MovePosC2S{}, s.MovePosC2S)
```

## 六、启动gen server 协程
### 1. gen server 模板
注意事项：\
	1.文件必须放在internal/game/services目录下 \
	2.文件名必须带_service
```go
package testsvc

import (
	"cake/internal/gensvc/rpc"
	"fmt"
)

//包内必须带有结构体State
type State struct {
}

type Service struct {
	*rpc.Service
}

//启动方法
func Start() (*rpc.Service, error) {
	s := &Service{}
	//协程配置必须使用rpc.NewCfg()
	cfg := rpc.NewCfg()
	//Cfg{
	//	InitArgs 初始方法Init时的第二个参数
	//	StartTimeout 协程启动超时时间
	//	SendFn:      自定义send方法
	//	SendMaxCap:  缓冲chan最大容量多少
	//}
	cfg.SendMaxCap = 1
	//rpc启动方法 参数1 协程注册名，协程注册结构，协程配置
	//也可以用rpc.Start("test", s)启动协程，走默认配置
	roleRpc, err := rpc.StartWithCfg("test", s, cfg)
	return roleRpc, err
}

func (s *Service) SvcName() string {
	return "test"
}

// 必须实现，每次启动协程会自动调用一次该方法，必须返回结构体State和镶嵌*rpc.Service
func (s *Service) Init(r *rpc.Service, args any) (any, error) {
	s.Service = r
	return &State{}, nil
}


// 必须实现，每次协程结束会自动调用一次该方法 ，如果初始化失败不会执行
func (s *Service) Stop(rawState any) {
	state := rawState.(*State)
}

//rpc 方法，只有方法名前面带有rpc并且符合类型 func(state State, args any) (any, error) 会自动注册
func (s *Service) RpcTest(state State, args any) (any, error) {
	fmt.Println("send_test", args)
	//返回参数，如果是通过call调用，同步返回给调用方，send调用忽略
	return nil, nil
}

var _ rpc.GenService = &Service{}

```
### 2. ast静态分析
基于ast静态分析，会筛选出internal/game/services的Service结构体, 带有前缀的Rpc方法会自动注册并生成常量id,会自动生成常量rpcid.RpcXXX
```go
//会生成常量 const RpcTest = "RpcTest"
func (s *Service) RpcTest(state State, args any) (any, error)
```

通过id发送到chan，在协程内部调用对应的方法

生成的分析文件在internal/gensvc/rpcgen

### 3. 服务协程间通信方法
```go
//启动rpc服务
testRpc, err := rpc.StartWithCfg("test", s)

//有缓存chan发送，协程内chan接收到信息后会执行RpcTest方法
//会等待多少秒发送成功，发送不成功就丢弃
testRpc.SendTimeout(rpcid.RpcTest, "hello world", 5*time.Second)
//默认等待5秒，发送不成功就丢弃
testRpc.Send5S(rpcid.RpcTest, "hello world")
//无等待，chan满了直接丢弃
testRpc.Send(rpcid.RpcTest, "hello world")
//协程的chan接受到信息后延后3秒再执行
testRpc.SendAfter(3*time.Second,rpcid.RpcTest, "hello world",, 5*time.Second)

//无缓冲chan发送，同步返回结果
relust, err := testRpc.CallTimeout(rpcid.RpcTest, "hello world", 5*time.Second)
relust, err := testRpc.Call5S(rpcid.RpcTest, "hello world")

//通过协程名发送
rpc.Call5s("test", rpcid.RpcTest, "hello world")
rpc.CallTimeout("test", rpcid.RpcTest, "hello world", 5*time.Second)
rpc.Send("test", rpcid.RpcTest, "hello world")
rpc.AfterSend(3*time.Second, "test", rpcid.RpcTest, "hello world")
```

### 4. 服务协程内部的定时器
```go
func (s *Service) registerTimer() {
    s.AddTimer(定时器名唯一, 执行间隔, 执行次数（-1是循环）, 执行方法, 执行参数)
    s.AddTimer("TimerSaveRoleDB", 5*time.Second, -1, s.TimerSaveRoleDB, nil)
}
```

### 5. State结构体数据安全
State结构体实现Copy和Restore方法，在每次处理消息时，gensvc会自动执行Copy方法记录下要数据，当消息返回错误执行Restore方法

```go
type GenState interface {
	Copy(string)  (any, bool)
	Restore(any)
}

// 例子：
type State struct {
	*model.Role
}

// 在处理消息前深度复制数据
func (s *State) Copy(cmd string) (any, bool) {
	var isSkip bool
	//对于高频的心跳包和移动不做copy处理，实现黑名单制度
	switch cmd {
	case "TimerCheckHeartbeat":
		isSkip = true
	case rpcid.RpcHeartbeat:
		isSkip = true
	case rpcid.RpcMovePath:
		isSkip = true
	}
	if isSkip {
		return nil, false
	}
	return s.Role.CloneRolePO(), true
}

// 失败后调用接口数据恢复
func (s *State) Restore(rawData any) {
	rolePO, ok := rawData.(*model.RolePO)
	if !ok {
		return
	}
	s.Role.RolePO = rolePO
}

var _ rpc.GenState = new(State)
```