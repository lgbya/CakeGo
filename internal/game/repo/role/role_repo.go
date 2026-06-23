package role

import (
	"cake/env"
	"cake/internal/game/model"
	"cake/internal/pkg/db"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	roleRepoInst *repo
	roleRepoOnce sync.Once
)

type repo struct {
	nextSeq uint64
}

func Repo() *repo {
	roleRepoOnce.Do(func() {
		roleRepoInst = &repo{}
		//todo:偷个懒，唯一id不应该这样生成，下个标记后续再改
		startSeq, err := roleRepoInst.Count()
		if err != nil {
			panic(err)
		}
		atomic.StoreUint64(&roleRepoInst.nextSeq, uint64(startSeq))
	})
	return roleRepoInst
}

func (r *repo) FindRoleByID(RoleID uint64) (*model.RolePO, error) {
	var role *model.RolePO
	err := db.DbInst().Where("role_id = ?", RoleID).First(&role).Error
	return role, err
}

// 账号查角色列表
func (r *repo) ListRolesByAccount(account string) ([]*model.RolePO, error) {
	var roles []*model.RolePO
	err := db.DbInst().Where("account = ?", account).Find(&roles).Error
	return roles, err
}

// 查询一共有多少个角色
func (r *repo) CountRolesByAccount(account string) (int64, error) {
	var count int64
	err := db.DbInst().Model(&model.RolePO{}).Where("account = ?", account).Count(&count).Error
	return count, err
}

// 查询一共有多少个角色
func (r *repo) Count() (int64, error) {
	var count int64
	err := db.DbInst().Model(&model.RolePO{}).Count(&count).Error
	return count, err
}

// 插入新角色
func (r *repo) CreateRole(role *model.RolePO) error {
	return db.DbInst().Create(role).Error
}

// 更新角色
func (r *repo) UpdateRole(role *model.RolePO) error {
	if role == nil {
		return errors.New("role data cannot be nil")
	}
	if role.RoleID <= 0 {
		return errors.New("invalid role_id")
	}
	return db.DbInst().Model(&model.RolePO{}).Where("role_id = ?", role.RoleID).
		Select("account", "server_id", "plat_id", "name", "gender", "career", "lv", "data").
		Updates(role).Error
}

// 检查角色名唯一
func (r *repo) CheckRoleNameUnique(name string) (bool, error) {

	tx := db.DbInst().Model(&model.RolePO{}).Where("name = ?", name)

	var cnt int64
	if err := tx.Count(&cnt).Error; err != nil {
		return false, err
	}
	if cnt > 0 {
		return false, nil
	}
	return true, nil
}

// 生成唯一角色id
func (r *repo) GenRoleID() uint64 {
	serverID := env.ServerID()
	platID := env.PlatID()
	id := atomic.AddUint64(&r.nextSeq, 1)
	platMask := uint64(platID) & 0xFF
	serverMask := uint64(serverID) & 0xFFFF
	seqMask := id & 0xFFFFFFFFFF // 40位掩码 2^40-1
	return platMask<<56 | serverMask<<40 | seqMask
}

// 反解唯一角色id
func (r *repo) ParseRoleID(roleID uint64) (uint64, uint64, uint64) {
	platID := (roleID >> 56) & 0x7F
	serverID := (roleID >> 40) & 0xFFFF
	id := roleID & 0xFFFFFFFFFF
	return serverID, platID, id
}
