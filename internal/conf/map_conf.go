package conf

type MapConf struct {
	ID        uint32
	Name      string
	Type      int
	Width     int
	Height    int
	CellSize  int
	BlockSize int
	SpawnX    int
	SpawnY    int
}

var MapConfs = map[uint32]MapConf{
	1000: {1000, "奥利特尔城", 1, 47000, 32000, 100, 2200, 2333, 3222},
	1001: {1001, "风花村", 1, 47000, 32000, 100, 2200, 2333, 3222},
}
