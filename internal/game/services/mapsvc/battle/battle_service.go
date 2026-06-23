package battle

import (
	"cake/internal/game/model"
	"cake/internal/gensvc/rpc"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
	"cake/internal/util/sys"
	"cake/proto/pb"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Service struct {
	*rpc.Service
	ID       uint32
	msgCache []*rpc.Msg
	sceneRpc *rpc.Service
}

func StartService(id uint32, sceneRpc *rpc.Service) (*rpc.Service, error) {
	s := &Service{
		ID:       id,
		msgCache: make([]*rpc.Msg, 0, 32),
		sceneRpc: sceneRpc,
	}
	cfg := rpc.NewCfg()
	cfg.Ctx = sceneRpc.GetCtx()
	cfg.SendFn = s.AddMsgCache
	cfg.SendMaxCap = 10000
	return rpc.StartWithCfg(s.SvcName(id), s, cfg)
}

func (s *Service) SvcName(id uint32) string {
	return "battle_svc_" + strconv.Itoa(int(id))
}

func (s *Service) Init(rpcSvc *rpc.Service, _ any) (any, error) {
	logger.Debugf("帧计算启动")
	s.Service = rpcSvc
	s.AddTimer("TimerFrameCalculation", 33*time.Millisecond, -1, s.TimerFrameCalculation, nil)
	return newState(), nil
}

// 帧计算
func (s *Service) TimerFrameCalculation(rawState, _ any) error {
	//logger.Debugf("帧计算计算启动")
	state := rawState.(*State)
	msgs := s.msgCache
	// 复用容量清空，不会扩容GC
	s.msgCache = make([]*rpc.Msg, 0, cap(s.msgCache))
	for _, msg := range msgs {
		sys.SafeRun(func() {
			defer s.PutMsg(msg)
			_, err := s.DoMsgFn(msg)
			if err != nil {
				logger.Errorf("战斗进程处理协议错误 %v %v", msg, err)
			}
		})
	}

	emptyBattleRoles := make(map[uint64]model.BattleRole, len(state.dirtyRoleIDs))
	for roleID := range state.dirtyRoleIDs {
		battleRole, ok := state.battleRoles[roleID]
		if !ok {
			continue
		}
		emptyBattleRoles[roleID] = battleRole
	}
	//帧结束
	state.dirtyRoleIDs = make(map[uint64]struct{})
	s.sceneRpc.Send5s(rpcid.RpcSyncRoleStates, &emptyBattleRoles)
	return nil
}

func (s *Service) AddMsgCache(msg *rpc.Msg) error {
	if len(s.msgCache) >= cap(s.msgCache) {
		return errors.New(fmt.Sprintf("战斗消息缓存已满，丢弃消息:%v", msg))
	}
	s.msgCache = append(s.msgCache, msg)
	return nil
}

func (s *Service) Stop(_ any) {

}

// ================rpc方法=========================
func (s *Service) RpcAddBattleRole(state *State, rawMsg any) (any, error) {
	battleRole := rawMsg.(*model.BattleRole)
	state.battleRoles[battleRole.RoleID] = *battleRole
	return nil, nil
}

func (s *Service) RpcDelBattleRole(state *State, rawRoleID any) (any, error) {
	roleID := rawRoleID.(uint64)
	delete(state.battleRoles, roleID)
	return nil, nil
}

func (s *Service) RpcMovePath(state *State, rawArgs any) (any, error) {
	args := rawArgs.(map[string]any)
	roleID := args["id"].(uint64)
	movePosC2S := args["msg"].(*pb.MovePosC2S)
	battleRole := state.battleRoles[roleID]
	battleRole.Pos.X = int(movePosC2S.Pos.X)
	battleRole.Pos.Y = int(movePosC2S.Pos.Y)
	battleRole.SyncTypeMove()
	state.SaveBattleRole(battleRole)
	state.SetDirty(roleID)
	return nil, nil
}

func (s *Service) RpcTest(_, _ any) (any, error) {
	return nil, nil
}
