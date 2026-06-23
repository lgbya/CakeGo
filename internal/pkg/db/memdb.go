package db

import (
	"github.com/hashicorp/go-memdb"
)

func MemCfg() map[string]*memdb.TableSchema {
	return map[string]*memdb.TableSchema{
		"role": {
			Name: "role",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "BcastRoleID"},
				},
			},
		},
		"scene": {
			Name: "scene",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "sceneID"},
				},
			},
		},
	}
}

// 全局DB单例
var memInst *memdb.MemDB

func MemInst() *memdb.MemDB {
	return memInst
}

func InitMemDb() {
	schema := &memdb.DBSchema{
		Tables: MemCfg(),
	}
	inst, err := memdb.NewMemDB(schema)
	if err != nil {
		panic("创建 go-memdb错误")
	}
	memInst = inst
}
