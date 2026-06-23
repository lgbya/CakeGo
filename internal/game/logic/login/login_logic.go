package login

import (
	"cake/env"
	"cake/internal/game/def"
	"cake/internal/game/def/errcode"
	"cake/internal/game/logic/sdk"
	"cake/internal/game/model"
	rolerepo "cake/internal/game/repo/role"
	"cake/internal/game/services/rolesvc"
	"cake/internal/gate/igate"
	"cake/internal/util/errx"
	"cake/proto/pb"
	"strings"
	"sync"
	"time"
)

var (
	logicInst *logic
	logicOnce sync.Once
)

type logic struct {
}

func Logic() *logic {
	logicOnce.Do(func() {
		logicInst = &logic{}
	})
	return logicInst
}

// 账号认证
func (ll *logic) AccountAuth(connSvc igate.ConnSvc, loginAuthC2S *pb.AccountAuthC2S) {
	loginAuthS2C := &pb.AccountAuthS2C{IsAuth: true}
	if err := sdk.AuthChannel(loginAuthC2S); err != nil {
		loginAuthS2C.IsAuth = false
		connSvc.SendFail(loginAuthS2C, errx.GetCode(err))
		return
	}
	authData := map[string]any{
		"account":   loginAuthC2S.Account,
		"server_id": loginAuthC2S.ServerId,
		"plat_id":   loginAuthC2S.PlatId,
	}
	connSvc.SetAuthData(authData)

	connSvc.SendSuccess(loginAuthS2C)
}

// 查询角色
func (ll *logic) SelectRoles(connSvc igate.ConnSvc, selectRolesC2S *pb.SelectRolesC2S) {
	account := connSvc.GetAccount()
	if account == "" && selectRolesC2S.Account != account {
		connSvc.SendFail(&pb.SelectRolesS2C{}, errcode.LoginAccountErr)
		return
	}

	roles, err := rolerepo.Repo().ListRolesByAccount(account)
	if err != nil {
		connSvc.SendFail(&pb.SelectRolesS2C{}, errcode.LoginSelectRoles)
		return
	}
	var roleList []*pb.RoleInfo
	for _, roleState := range roles {
		roleInfo := &pb.RoleInfo{
			RoleId:   roleState.RoleID,
			ServerId: roleState.ServerID,
			PlatId:   roleState.PlatID,
			Name:     roleState.Name,
			Lv:       roleState.Lv,
			Gender:   uint32(roleState.Gender),
			Career:   uint32(roleState.Career),
		}
		roleList = append(roleList, roleInfo)
	}

	selectRolesS2C := &pb.SelectRolesS2C{
		RoleList: roleList,
	}
	connSvc.SendSuccess(selectRolesS2C)
}

// 创建角色
func (ll *logic) CreateRole(connSvc igate.ConnSvc, createRoleC2S *pb.CreateRoleC2S) {
	account := connSvc.GetAccount()
	if account == "" {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginAccountErr)
		return
	}

	roleCnt, err := rolerepo.Repo().CountRolesByAccount(account)
	if err != nil && int(roleCnt) <= def.CreateRoleMaxCnt {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginCreateRoleMax)
		return
	}

	career := def.Career(createRoleC2S.Career)
	if !career.IsValid() {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginCareerNotExists)
		return
	}

	gender := def.Gender(createRoleC2S.Gender)
	if !gender.IsValid() {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginGenderNotExists)
		return
	}
	// todo 后续这里要加上锁防击穿
	roleRepo := rolerepo.Repo()
	name := createRoleC2S.Name
	if name == "" {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginRoleNameEmpty)
		return
	}
	ok, err := roleRepo.CheckRoleNameUnique(name)
	if err != nil {
		err = errx.From(err)
		connSvc.SendFail(&pb.CreateRoleS2C{}, errx.GetCode(err))
		return
	}

	if ok == false {
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginRoleNameExists)
		return
	}

	//生成唯一id
	roleID := roleRepo.GenRoleID()
	//插入数据
	rolePO := &model.RolePO{
		RoleID:    roleID,
		Account:   account,
		ServerID:  env.ServerID(),
		PlatID:    env.PlatID(),
		Name:      name,
		Gender:    gender,
		Career:    career,
		CreatedAt: time.Now(),
	}
	roleState := &model.Role{
		RolePO: rolePO,
	}

	if err = roleRepo.CreateRole(rolePO); err != nil {
		if strings.Contains(err.Error(), "uk_name") {
			connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginRoleNameExists)
			return
		}
		connSvc.SendFail(&pb.CreateRoleS2C{}, errcode.LoginCreateRoleFail)
	}

	connSvc.SendSuccess(&pb.CreateRoleS2C{
		RoleId:   roleState.RoleID,
		ServerId: roleState.ServerID,
		PlatId:   roleState.PlatID,
		Name:     roleState.Name,
		Gender:   uint32(roleState.Gender),
		Career:   uint32(roleState.Career),
	})
}

func (ll *logic) LoginRole(connSvc igate.ConnSvc, loginRoleC2S *pb.LoginRoleC2S) {
	account := connSvc.GetAccount()
	if account == "" {
		connSvc.SendFail(&pb.LoginRoleS2C{}, errcode.LoginAccountErr)
		return
	}

	roleRepo := rolerepo.Repo()

	RoleID := loginRoleC2S.RoleId
	if serverID, platID, Id := roleRepo.ParseRoleID(RoleID); serverID <= 0 || platID <= 0 || Id <= 0 {
		connSvc.SendFail(&pb.LoginRoleS2C{}, errcode.LoginInvalidRoleID)
		return
	}

	roleDB, err := roleRepo.FindRoleByID(RoleID)
	if err != nil {
		connSvc.SendFail(&pb.LoginRoleS2C{}, errcode.LoginRoleNotExists)
		return
	}

	if account != roleDB.Account {
		connSvc.SendFail(&pb.LoginRoleS2C{}, errcode.LoginAccountErr)
		return
	}

	roleState := &model.Role{
		RolePO: roleDB,
		Conn:   connSvc.GetRoleConn(),
	}
	roleRpc, err := rolesvc.StartRole(roleState)
	if err != nil {
		errCode := errx.GetCode(err)
		if errCode != errcode.System {
			connSvc.SendFail(&pb.LoginRoleS2C{}, errCode)
			return
		}
		connSvc.SendFail(&pb.LoginRoleS2C{}, errcode.LoginRoleWorkerFail)
		return
	}
	connSvc.SetRoleRpc(roleState.RoleID, roleRpc)
	loginRoleS2C := &pb.LoginRoleS2C{
		RoleId:   roleDB.RoleID,
		ServerId: env.ServerID(),
		PlatId:   env.PlatID(),
		Name:     roleDB.Name,
		Gender:   uint32(roleDB.Gender),
		Career:   uint32(roleDB.Career),
		Lv:       roleDB.Lv,
	}
	connSvc.SendSuccess(loginRoleS2C)
}
