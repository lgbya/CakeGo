package rolesvc

import (
	"cake/internal/game/def/errcode"
	"cake/internal/game/model"
	rolepo "cake/internal/game/repo/role"
	"cake/internal/gate/router/irouter"
	"cake/internal/gensvc/rpc"
	"cake/internal/pkg/logger"
	"cake/internal/util/errx"
	"strconv"
	"time"
)

type Service struct {
	RoleID uint64
	*rpc.Service
	*model.Conn
}

func StartService(roleState *model.Role, cfg rpc.Cfg) (*rpc.Service, *Service, error) {
	s := &Service{RoleID: roleState.RoleID, Conn: roleState.Conn}
	svcName := s.SvcName(roleState.RoleID)
	roleRpc, err := rpc.StartWithCfg(svcName, s, cfg)
	return roleRpc, s, err
}

func (s *Service) SvcName(RoleID uint64) string {
	return "role_svc_" + strconv.FormatUint(RoleID, 10)
}

func (s *Service) Init(rpcSvc *rpc.Service, rawRole any) (any, error) {
	role, ok := rawRole.(*model.Role)
	if !ok {
		return nil, errx.New(errcode.LoginRoleWorkerFail)
	}
	role.RoleRpc = rpcSvc

	s.Service = rpcSvc
	if err := HandleLogin(role); err != nil {
		return nil, err
	}
	//s.registerRpc()
	s.registerTimer()
	logger.Debugf("启动角色进程 账号:%s, id:%d, 名称:%s ", role.Account, role.RoleID, role.Name)
	return newState(role), nil
}

func (s *Service) Stop(rawState any) {
	state, ok := rawState.(*State)
	if !ok {
		logger.Errorf("记录数据库错误 %v", rawState)
	}
	HandleLogout(state.Role)
	err := rolepo.Repo().UpdateRole(state.RolePO)
	if err != nil {
		logger.Errorf("记录数据库错误 %v, %v", state.RolePO, err)
	}

	logger.Debugf("[%d]角色进程正在关闭", s.RoleID)
}

// ========================= rpc 方法
func (s *Service) RpcRoleCmd(state *State, args any) (any, error) {
	cmd := args.(irouter.RoleCmd)
	_ = cmd.RoleFn(state.Role, cmd.Msg)
	return nil, nil
}

func (s *Service) RpcHeartbeat(state *State, _ any) (any, error) {
	nowTime := time.Now().Unix()
	state.Heartbeat.LastTime = nowTime
	return nil, nil
}

func (s *Service) RpcSaveSceneRole(state *State, rawSceneRole any) (any, error) {
	sceneRole := rawSceneRole.(*model.SceneRole)
	state.Data.Location = sceneRole.Location
	state.Save()
	return nil, nil
}

// 网关进程关会发一条消息到这里
func (s *Service) RpcConnClose(state *State, _ any) (any, error) {
	state.Conn.CloseConn()
	s.AddTimer("TimerStopRole", 10*time.Second, 1, s.TimerStopRole, nil)
	return nil, nil
}

var _ rpc.GenService = new(Service)
