package main

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"cake/test/tcpclient/client"
)

func main() {
	env.Init()
	logger.Init()
	client.NewClient(20)
}
