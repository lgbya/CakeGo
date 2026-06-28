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
	1000: {1000, "奥利特尔城", 1, 1600, 1600, 100, 200, 800, 800},
	1001: {1001, "风花村", 1, 1600, 1600, 100, 200, 800, 800},
}
