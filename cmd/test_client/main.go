package main

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"cake/test_client/client"
)

func main() {
	env.Init()
	logger.Init()
	client.NewClient(1000)
}
