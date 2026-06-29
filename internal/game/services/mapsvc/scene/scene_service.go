package scene

import (
	"cake/internal/game/def/errcode"
	"cake/internal/game/model"
	"cake/internal/game/services/mapsvc/battle"
	"cake/internal/game/services/mapsvc/bcast"
	"cake/internal/gate/router/irouter"
	"cake/internal/gensvc/rpc"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
	"cake/internal/util/errx"
	"cake/proto/pb"
	"strconv"
)

type Service struct {
	*rpc.Service
	*MapBase               //地图基础信息
	ID        uint32       //地图唯一id
	*State                 //记录场景的数据
	BattleRpc *rpc.Service //战斗帧进程
	BcastRpc  *rpc.Service //广播进程
	recvChan  chan irouter.RoleCmd
	Grids     map[model.Pos]Cell // 二维网格
}

func StartService(id uint32, cfg rpc.Cfg, mapBase *MapBase) (*rpc.Service, *Service, error) {
	scene := newState(id)
	s := &Service{
		ID:      id,
		MapBase: mapBase,
		State:   scene,
	}
	sceneRpc, err := rpc.StartWithCfg(s.SvcName(id), s, cfg)
	return sceneRpc, s, err
}

func (s *Service) SvcName(ID uint32) string {
	return "scene_svc_" + strconv.Itoa(int(ID))
}

func (s *Service) Init(rpcSvc *rpc.Service, _ any) (any, error) {
	logger.Debugf("地图进程启动成功%d", s.MapID)
	s.Service = rpcSvc
	s.Grids = s.initNineGirds()
	//启动一个战斗帧进程
	battleRpc, err := battle.StartService(s.ID, rpcSvc)
	if err != nil {
		return nil, err
	}
	s.BattleRpc = battleRpc

	//启动一个广播进程
	bcastRpc, err := bcast.StartService(s.ID, s.GetCtx())
	if err != nil {
		return nil, err
	}
	s.BcastRpc = bcastRpc
	return s.State, nil
}

func (s *Service) Stop(_ any) {
	logger.Infof("[sceneID:%d|MapID:%d]地图进程关闭中", s.ID, s.MapID)
}

func (s *Service) GetBattleRpc() *rpc.Service {
	return s.BattleRpc
}

// ===================== rpc ===============
func (s *Service) RpcRoleCmd(state *State, rawCmd any) (any, error) {
	cmd := rawCmd.(irouter.RoleCmd)
	sceneRole, ok := state.sceneRoles[cmd.RoleID]
	if !ok {
		return nil, nil
	}
	_ = cmd.SceneFn(s, sceneRole, cmd.Msg)
	return nil, nil
}

func (s *Service) RpcEnterScene(state *State, rawSceneRole any) (any, error) {
	sceneRole := rawSceneRole.(*model.SceneRole)
	if sceneRole.RoleID <= 0 {
		return nil, errx.New(errcode.SceneRoleIDIllegal)
	}
	if _, ok := state.sceneRoles[sceneRole.RoleID]; ok {
		logger.Debugf("当前在scene进程 %+v", sceneRole)
		s.BcastRpc.Send5s(rpcid.RpcSaveConnRole, sceneRole)
		s.sendRoleViewList(state, sceneRole, false)
		return sceneRole.Location, nil
	}

	if sceneRole.MapID != s.MapID {
		sceneRole.UpdatePos(s.SpawnPos)
	}

	isUpdateGrid := s.updateRoleGrid(sceneRole)
	sceneRole.SceneID = s.sceneID
	sceneRole.MapID = s.MapID
	state.sceneRoles[sceneRole.RoleID] = sceneRole
	s.BcastRpc.Send5s(rpcid.RpcSaveConnRole, sceneRole)
	s.BattleRpc.Send5s(rpcid.RpcAddBattleRole, sceneRole.ToBattle())
	s.sendRoleViewList(state, sceneRole, isUpdateGrid)
	return sceneRole.Location, nil
}

// 退出场景
func (s *Service) RpcLeaveScene(state *State, rawRoleID any) (any, error) {

	roleID := rawRoleID.(uint64)
	logger.Debugf("玩家退出场景 %d", roleID)

	if roleID <= 0 {
		return nil, errx.New(errcode.SceneRoleIDIllegal)
	}
	sceneRole, ok := state.sceneRoles[roleID]
	if !ok {
		return nil, nil
	}

	//删除数据
	delete(state.sceneRoles, roleID)
	s.BcastRpc.Send5s(rpcid.RpcDelConnRole, roleID)
	s.BattleRpc.Send5s(rpcid.RpcDelBattleRole, roleID)

	isUpdateGrid := s.delRoleGrid(sceneRole)

	//获取当前九宫格玩家
	viewRoleIDs := s.get9GridViewRoles(sceneRole.GridPos)

	//通知视野里的角色删除玩家
	msg := &pb.RoleViewDelS2C{
		RoleId: roleID,
	}
	for viewRoleID := range viewRoleIDs {
		viewSceneRole, ok := state.sceneRoles[viewRoleID]
		if !ok {
			continue
		}
		//玩家更新格子才通知其他人
		if viewRoleID != sceneRole.RoleID && isUpdateGrid {
			viewSceneRole.Conn.SendMsg(msg)
		}
	}

	return nil, nil
}

// 刷新脏标记
func (s *Service) RpcSyncRoleStates(state *State, rawBattleRoles any) (any, error) {
	//logger.Infof("帧结束，计算广播内容")
	battleRoles := *rawBattleRoles.(*map[uint64]model.BattleRole)
	for roleID, battleRole := range battleRoles {
		sceneRole, ok := state.sceneRoles[roleID]
		if !ok {
			continue
		}
		sceneRole.SyncBattle(&battleRole)
		sceneRole.RoleRpc.Send5s(rpcid.RpcSaveSceneRole, sceneRole)
		//todo 偷个懒，直接根据类型做简单的广播同步,应该是存多种同步类型
		switch battleRole.SyncType {
		case model.BattleSyncTypeMove:
			s.bcastNineGridMsg(sceneRole, &pb.MovePosS2C{
				RoleId: sceneRole.RoleID,
				MapId:  s.MapID,
				Pos:    sceneRole.Pos.Pb(),
			})
		default:
		}
		//
	}
	return nil, nil
}
