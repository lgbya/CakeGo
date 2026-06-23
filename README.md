模仿erlang的gen server实现进程单游戏服Actor 模型的mmo游戏框架
目录
CakeGo/
├── cmd                     # 服务启动入口
│   └── server
│       └── main.go         # 程序入口：初始化配置、启动网关、启动所有游戏Actor服务
├── env                     # 环境配置
│   ├── local.yaml          
│   └── init.go           
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

安装命令
sh run.sh install 平台号 服务器号
例：sh run.sh install 1 1

启动命令
sh run.sh start 平台号 服务器号
例：sh run.sh start 1 1

压测客户端
sh run.sh test 平台号 服务器号
例：sh run.sh test 1 1
