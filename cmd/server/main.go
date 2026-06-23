package main

import (
	"cake/internal/game"
	"cake/internal/pkg/logger"
)

func main() {
	defer logger.Sync()
	game.Init()
	game.Stop()
}
