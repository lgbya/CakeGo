package role

import (
	"cake/internal/game/def"
	"cake/internal/game/def/errcode"
	"cake/internal/game/model"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/internal/pkg/logger"
	"cake/internal/util/errx"
	"cake/proto/pb"
	"errors"
	"fmt"
	"time"
)

func (l *logic) LoginEnter(roleMod *model.Role) error {
	//玩家登录没有地图进入默认地图
	sceneID := roleMod.Location().SceneID
	sceneRpc := iscene.Manager().RpcBySceneID(sceneID)
	if sceneRpc == nil {
		sceneID = iscene.Manager().MapIdToSceneId(def.MainSceneMapID)
	}

	if err := l.EnterScene(roleMod, sceneID, true); err != nil {
		return err
	}
	location := roleMod.Location()
	roleMod.SendSuccess(&pb.EnterSceneS2C{MapId: location.MapID, Pos: location.Pos.Pb()})
	return nil
}

func (l *logic) LogoutLeave(roleMod *model.Role) error {
	//玩家登录没有地图进入默认地图
	logger.Debugf("[%d]执行下线后退出地图逻辑,", roleMod.RoleID)
	sceneRpc := iscene.Manager().RpcBySceneID(roleMod.Location().SceneID)
	if sceneRpc == nil {
		return errors.New(fmt.Sprintf("[%d]没有找到对应的地图 %v", roleMod.RoleID, roleMod.Location()))
	}
	sceneRpc.Send5s(rpcid.RpcLeaveScene, roleMod.RoleID)
	return nil
}

// 进入大场景
func (l *logic) EnterMap(roleMod *model.Role, mapID uint32) error {
	sceneID := iscene.Manager().MapIdToSceneId(mapID)
	if sceneID == 0 {
		return errx.New(errcode.SceneMapNonExistent)
	}

	return l.EnterScene(roleMod, sceneID, false)
}

func (l *logic) EnterScene(roleMod *model.Role, sceneID uint32, isLogin bool) error {
	roleSceneID := roleMod.Location().SceneID
	sceneRpc := iscene.Manager().RpcBySceneID(sceneID)
	if sceneRpc == nil {
		return errx.New(errcode.SceneNonExistent)
	}

	//非登录才退出上一个场景
	if !isLogin {
		if roleSceneID == sceneID {
			return errx.New(errcode.SceneAlreadyIn)
		}

		//先退出场景
		if roleSceneID == 0 && roleMod.SceneRpc != nil {
			return errx.New(errcode.SceneLeaveFail)
		}

		if roleMod.SceneRpc != nil {
			_, err := roleMod.SceneRpc.CallTimeout(rpcid.RpcLeaveScene, roleMod.RoleID, 10*time.Second)
			if err != nil {
				return errx.New(errcode.SceneLeaveFail)
			}
		}
	}

	//在进入新场景
	sceneRole := model.NewSceneRole(roleMod)
	rawLocation, err := sceneRpc.CallTimeout(rpcid.RpcEnterScene, sceneRole, 10*time.Second)

	if err != nil {
		return err
	}

	location, ok := rawLocation.(model.Location)
	if !ok {
		return errx.New(errcode.SceneLocationFail)
	}
	roleMod.SetLocation(location)
	roleMod.SceneRpc = sceneRpc
	logger.Debugf("[%d]玩家进入场景[%d]", roleMod.RoleID, roleMod.Location().MapID)
	//roleMod.SendSuccess(&pb.EnterSceneS2C{MapId: location.MapID, Pos: location.Pos.Pb()})
	return nil
}
