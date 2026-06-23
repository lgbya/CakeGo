package logger

import "cake/env"

func Init() {
	cfg := env.GetLog()
	InitApp(cfg)
	InitProto(cfg)
}
