package logger

import (
	"cake/env"
	"cake/internal/gate/packet"
	"cake/proto/pb"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"reflect"
)

// 协议专属格式化日志器，独立文件输出
var protoSugar *zap.SugaredLogger
var noLogCmd = map[uint32]bool{}
var Level = "debug"

func InitProto(cfg env.Log) {

	Level = cfg.Level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		panic("日志等级不存在")
	}
	// 1. 日志滚动配置，防止单个日志文件过大
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Path + "/pb/proto.log",
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 2. 正确初始化编码器配置
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.LevelKey = ""  // 隐藏 INFO/DEBUG 级别标识
	encoderCfg.CallerKey = "" // 隐藏文件路径、代码行号
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	// 3. 构建控制台格式编码器
	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	// 4. 创建独立日志核心
	// 多输出位置：文件 + 控制台
	multiSync := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(lumberJackLogger),
		zapcore.AddSync(os.Stdout),
	)
	core := zapcore.NewCore(encoder, multiSync, level)

	protoLogger := zap.New(core)
	protoSugar = protoLogger.Sugar()
	for _, v := range env.GetIntSlice("log.banProto") {
		noLogCmd[uint32(v)] = true
	}
}

func CheckLogCmd(cmd uint32) bool {
	return !noLogCmd[cmd]
}

func RecvProto(account string, roleID uint64, cmd uint32, msg proto.Message) {
	if Level == "debug" {
		if noLogCmd[cmd] {
			return
		}
		msgText := FormatProtoMsg(msg)
		if account != "" {
			protoSugar.Debugf("[%s|%d] %s:%d	%v", account, roleID, "接收", cmd, msgText)
		} else {
			protoSugar.Debugf("[] %s:%d	%v", "接收", cmd, msgText)
		}
	}
}

func SendProto(account string, roleID uint64, buf []byte) {
	if Level == "debug" {
		//var buf []byte
		//copy(buf, oldBuf)
		n := len(buf)
		cmd, data, _, _ := packet.DecodeMsg(buf[:n])
		if noLogCmd[cmd] {
			return
		}

		typ, ok := pb.GetS2CTypeByCmd(cmd)
		if !ok {
			return
		}
		msg := reflect.New(typ.Elem()).Interface().(proto.Message)

		if err := proto.Unmarshal(data, msg); err != nil {

			Debugf("cmd:%d 解析失败: %v", cmd, err)
			return
		}
		msgText := FormatProtoMsg(msg)
		protoSugar.Infof("[%s|%d] %s:%d	%v", account, roleID, "发送", cmd, msgText)
	}
}

func FormatProtoMsg(msg proto.Message) string {
	opt := protojson.MarshalOptions{
		UseProtoNames:   true,
		Indent:          "",
		EmitUnpopulated: true, // 关键：递归展开所有嵌套、零值字段
	}
	b, err := opt.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("marshal errx:%v", err)
	}
	return string(b)
}
