package bcast

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"context"
	"strconv"
)

type Service struct {
	*rpc.Service
	ID uint32
}

func StartService(id uint32, ctx context.Context) (*rpc.Service, error) {
	s := &Service{
		ID: id,
	}
	cfg := rpc.NewCfg()
	cfg.Ctx = ctx
	cfg.SendMaxCap = 10000
	return rpc.StartWithCfg(s.SvcName(id), s, cfg)
}

func (s *Service) SvcName(id uint32) string {
	return "bcast_svc_" + strconv.Itoa(int(id))
}

func (s *Service) Init(rpcSvc *rpc.Service, _ any) (any, error) {
	s.Service = rpcSvc
	return newState(), nil
}

func (s *Service) Stop(_ any) {

}

// ================rpc方法=========================
func (s *Service) RpcSaveConnRole(state *State, rawSceneRole any) (any, error) {
	sceneRole := rawSceneRole.(*model.SceneRole)
	state.connRoles[sceneRole.RoleID] = sceneRole.Conn
	return nil, nil
}

func (s *Service) RpcDelConnRole(state *State, rawRoleID any) (any, error) {
	roleID := rawRoleID.(uint64)
	delete(state.connRoles, roleID)
	return nil, nil
}

func (s *Service) RpcAoiNiceGrid(state *State, rawAoiInfo any) (any, error) {
	aoiInfo := rawAoiInfo.(*model.AoiInfo)
	for roleID := range aoiInfo.RoleIDs {
		if conn, ok := state.connRoles[roleID]; ok {
			conn.SendMsg(aoiInfo.Msg)
		}
	}
	return nil, nil
}
