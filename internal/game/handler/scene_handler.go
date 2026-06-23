package handler

import (
	"cake/internal/game/logic/role"
	"cake/internal/game/logic/scene"
	"cake/internal/game/model"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"cake/internal/gate/router/irouter"
	"cake/internal/util/errx"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
)

type SceneRoute struct {
}

func (s *SceneRoute) Register() {
	irouter.Reg().RoleCmd(&pb.EnterSceneC2S{}, s.EnterSceneC2S)
	irouter.Reg().SceneCmd(&pb.MovePosC2S{}, s.MovePosC2S)
}

func (*SceneRoute) EnterSceneC2S(roleMod *model.Role, rawMsg proto.Message) error {
	msg, ok := rawMsg.(*pb.EnterSceneC2S)
	if !ok {
		return nil
	}

	if msg.SceneId > 0 {
		if err := role.Logic().EnterScene(roleMod, msg.SceneId, false); err != nil {
			roleMod.SendFail(&pb.EnterSceneS2C{}, errx.GetCode(err))
			return err
		}
	}

	if err := role.Logic().EnterMap(roleMod, msg.MapId); err != nil {
		roleMod.SendFail(&pb.EnterSceneS2C{}, errx.GetCode(err))
		return err
	}

	location := roleMod.Location()
	roleMod.SendSuccess(&pb.EnterSceneS2C{MapId: location.MapID, Pos: location.Pos.Pb()})
	return nil
}

func (*SceneRoute) MovePosC2S(sceneSvc iscene.IService, sceneRole *model.SceneRole, rawMsg proto.Message) error {
	movePosC2S, ok := rawMsg.(*pb.MovePosC2S)
	if !ok {
		return nil
	}

	scene.Logic().MovePos(sceneSvc, sceneRole, movePosC2S)
	return nil
}
