package mapsvc

import (
	"cake/internal/conf"
	"cake/internal/game/services/mapsvc/scene"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"fmt"
)

func Init() {
	iscene.InitManager(scene.Manager())
	sceneMgr := scene.Manager()
	for mapID, _ := range conf.MapConfs {
		sceneMgr.AddMapBase(mapID)
		if sceneMgr.StartScene(mapID) == nil {
			panic(fmt.Sprintf("启动地图失败%d", mapID))
		}
	}
}

func Stop() {
	scene.Manager().Stop()
}
