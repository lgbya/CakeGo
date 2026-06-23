package env

import (
	"flag"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Config struct {
	// ... 其他配置
	Base Base `mapstructure:"base" validate:"required"`
	Log  Log  `mapstructure:"log" validate:"required"`
}

type Log struct {
	Level      string `mapstructure:"level"`      // debug/info/warn/error
	Path       string `mapstructure:"path"`       // 日志文件路径，如 ./logs/game.log
	MaxSize    int    `mapstructure:"maxSize"`    // 单个文件最大大小（MB）
	MaxBackups int    `mapstructure:"maxBackups"` // 最多保留文件数
	MaxAge     int    `mapstructure:"maxAge"`     // 文件保留天数
	Compress   bool   `mapstructure:"compress"`   // 是否压缩旧日志
	Console    bool   `mapstructure:"console"`    // 是否同时输出到控制台
}

type Base struct {
	ServerID uint32 `mapstructure:"serverID" validate:"required"`
	PlatID   uint32 `mapstructure:"platID" validate:"required"`
}

var v *viper.Viper
var cfg = &Config{}
var base = &Base{}

func Init() {
	env := flag.String("env", "local", "running environment: local")
	flag.Parse()

	// 初始化 viper 实例
	v = viper.New()

	// 加载配置的逻辑和之前一样
	v.SetConfigName("app")
	v.SetConfigType("yaml")
	v.AddConfigPath("./env")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("read base config failed: %v", err)
	}

	v.SetConfigName(*env)
	if err := v.MergeInConfig(); err != nil {
		log.Printf("warn: no %s config file", *env)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("GAME")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	log.Printf("running in %s mode", *env)

	if err := v.Unmarshal(cfg); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(base); err != nil {
		panic(err)
	}
}

func GetLog() Log {
	return cfg.Log
}

func ServerID() uint32 {
	return cfg.Base.ServerID
}

func PlatID() uint32 {
	return cfg.Base.PlatID
}

// 对外暴露 Get 方法，支持按 key 动态获取
func Get(key string) interface{} {
	return v.Get(key)
}

// 为了类型安全，再封装几个常用类型的 Getter
func GetString(key string) string {
	return v.GetString(key)
}

func GetInt(key string) int {
	return v.GetInt(key)
}

func GetBool(key string) bool {
	return v.GetBool(key)
}

func GetIntSlice(key string) []int {
	return v.GetIntSlice(key)
}
