package scene

import (
	"cake/internal/game/model"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/proto/pb"
	"sync"
)

var (
	logicInst *logic
	logicOne  sync.Once
)

type logic struct {
}

func Logic() *logic {
	logicOne.Do(func() {
		logicInst = &logic{}
	})
	return logicInst
}

// 账号认证
func (l *logic) MovePos(sceneSvc iscene.IService, sceneRole *model.SceneRole, movePosC2S *pb.MovePosC2S) {

	//这里加验证玩家能否移动的逻辑
	//通过后再发给战斗进程，保证战斗的序列化不会乱

	sceneSvc.GetBattleRpc().Send5s(rpcid.RpcMovePath, map[string]any{"id": sceneRole.RoleID, "msg": movePosC2S})
}
