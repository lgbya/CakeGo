package scene

import (
	"cake/internal/conf"
	"cake/internal/game/model"
	"math"
)

type Cell struct {
	SceneRoles map[uint64]struct{} // 玩家ID→玩家
	Units      map[uint64]struct{} // 怪物ID→怪物
}

type MapBase struct {
	MapID    uint32
	Name     string
	Type     int
	Width    int
	Height   int
	CellSize int //阻挡大小
	CellHypo int //阻挡斜边
	MaxGridX int // 横向大区块数量
	MaxGridY int // 纵向大区块数量
	//GridSize  int //格子大小
	GridSize int
	SpawnPos model.Pos // 出生点XY坐标（像素）
}

func NewMapBase(mapID uint32) *MapBase {
	mapConf, ok := conf.MapConfs[mapID]
	if !ok {
		return nil
	}
	//小格子
	cellHypo := int(float64(mapConf.CellSize) * math.Sqrt2)

	//大格子九宫格
	maxGridX := (mapConf.Width + mapConf.BlockSize - 1) / mapConf.BlockSize
	maxGridY := (mapConf.Height + mapConf.BlockSize - 1) / mapConf.BlockSize
	spawnPos := model.Pos{X: mapConf.SpawnX, Y: mapConf.SpawnY}
	return &MapBase{
		MapID:    mapID,
		Name:     mapConf.Name,
		Type:     mapConf.Type,
		Width:    mapConf.Width,
		Height:   mapConf.Height,
		CellSize: mapConf.CellSize,
		CellHypo: cellHypo,
		MaxGridX: maxGridX,
		MaxGridY: maxGridY,
		SpawnPos: spawnPos,
		GridSize: mapConf.BlockSize,
	}
}

func (m *MapBase) GetGridPos(pos model.Pos) model.Pos {
	GX := (pos.X + m.GridSize - 1) / m.GridSize
	GY := (pos.Y + m.GridSize - 1) / m.GridSize
	return model.Pos{X: GX, Y: GY}
}
