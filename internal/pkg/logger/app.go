package logger

import (
	"cake/env"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// 颜色常量
const (
	colorDebug = "\033[1;36m" // 加粗青色
	colorInfo  = "\033[1;32m" // 加粗绿色
	colorWarn  = "\033[1;33m" // 加粗黄色
	colorError = "\033[1;31m" // 加粗红色
	colorReset = "\033[0m"
)

func InitApp(cfg env.Log) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		panic("日志等级不存在")
	}

	// 基础编码器配置（共用时间、调用栈）
	baseEncCfg := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		CallerKey:  "caller",
		MessageKey: "msg",
		LineEnding: zapcore.DefaultLineEnding,
		EncodeTime: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			// 固定 25 字符宽度，左对齐
			enc.AppendString(fmt.Sprintf("%-25s", caller.TrimmedPath()))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	var cores []zapcore.Core

	// 1. 控制台：彩色 + 固定列宽
	if cfg.Console {
		conEncCfg := baseEncCfg
		conEncCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			// 先补空格定宽，再加颜色，解决排版错乱
			raw := fmt.Sprintf("%-5s", l.CapitalString())
			var color string
			switch l {
			case zapcore.DebugLevel:
				color = colorDebug
			case zapcore.InfoLevel:
				color = colorInfo
			case zapcore.WarnLevel:
				color = colorWarn
			case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
				color = colorError
			}
			enc.AppendString(color + raw + colorReset)
		}

		consoleEncoder := zapcore.NewConsoleEncoder(conEncCfg)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	// 2. 文件日志：纯文本/JSON，禁用颜色，防止乱码
	if cfg.Path != "" {
		_ = os.MkdirAll(filepath.Dir(cfg.Path), 0755)
		fileEncCfg := baseEncCfg
		// 文件日志不用颜色，正常定宽
		fileEncCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("%-5s", l.CapitalString()))
		}

		fileEncoder := zapcore.NewJSONEncoder(fileEncCfg)
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Path + "app/app.log",
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, writer, level))
	}

	core := zapcore.NewTee(cores...)
	// AddCallerSkip(1) 跳过封装层，展示真实代码行号
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar = logger.Sugar()
}

// Sync 退出前刷新缓冲区
func Sync() {
	_ = logger.Sync()
}

// 原生方法
func Debug(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { logger.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { logger.Fatal(msg, fields...) }

// 格式化方法
func Debugf(template string, args ...interface{}) { sugar.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { sugar.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { sugar.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { sugar.Errorf(template, args...) }
func Fatalf(template string, args ...interface{}) { sugar.Fatalf(template, args...) }

func WithOptions(opts ...zap.Option) *zap.Logger {
	return logger.WithOptions(opts...)
}
