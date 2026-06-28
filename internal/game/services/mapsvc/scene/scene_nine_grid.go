package scene

import (
	"cake/internal/game/def"
	"cake/internal/game/model"
	"cake/internal/gensvc/rpcgen/rpcid"
	"cake/proto/pb"
	"google.golang.org/protobuf/proto"
)

// ================
// 必须在修改sceneRole的xy坐标前执行
func (s *Service) updateRoleGrid(sceneRole *model.SceneRole) bool {
	//  先获取角色当前场景数据，不存在直接返回
	RoleID := sceneRole.RoleID
	if RoleID <= 0 {
		return false
	}
	//  计算当前格子坐标,更是玩家身上格子
	newGridPos := s.GetGridPos(sceneRole.Pos)
	sceneRole.GridPos = newGridPos

	// 判断格子是否存在
	newCell, ok := s.Grids[newGridPos]
	if !ok {
		return false
	}
	//算出旧格子
	oldGridPos := s.GetGridPos(sceneRole.OldPos)

	// 已经在目标格子，无需更新
	_, ok = newCell.SceneRoles[RoleID]
	if oldGridPos == newGridPos && ok {
		return false
	}

	//  清理角色【旧格子】中的数据（核心修复点）
	if oldCell, ok := s.Grids[oldGridPos]; ok {
		delete(oldCell.SceneRoles, RoleID)
	}

	newCell.SceneRoles[RoleID] = struct{}{}
	//  更新角色身上绑定的格子坐标，必须同步维护
	return true
}

func (s *Service) delRoleGrid(sceneRole *model.SceneRole) bool {
	//  先获取角色当前场景数据，不存在直接返回
	RoleID := sceneRole.RoleID
	if RoleID <= 0 {
		return false
	}

	//  计算当前格子坐标,更是玩家身上格子
	gridPos := s.GetGridPos(sceneRole.Pos)

	// 判断格子是否存在就删除
	if cell, ok := s.Grids[gridPos]; ok {
		delete(cell.SceneRoles, RoleID)
	}
	return true
}

func (s *Service) initNineGirds() map[model.Pos]Cell {
	grids := make(map[model.Pos]Cell)
	for y := -1; y <= s.MaxGridY; y++ {
		for x := -1; x <= s.MaxGridX; x++ {
			grids[model.Pos{X: x, Y: y}] = Cell{
				SceneRoles: make(map[uint64]struct{}),
				Units:      make(map[uint64]struct{}),
			}
		}
	}
	return grids
}

// 向角色所在九宫格广播消息
func (s *Service) bcastNineGridMsg(sceneRole *model.SceneRole, msg proto.Message) {

	// 获取九宫格所有接收者
	viewRoleIDs := s.get9GridViewRoles(sceneRole.GridPos)
	aoiInfo := model.AoiInfo{RoleIDs: viewRoleIDs, Msg: msg}
	s.BcastRpc.Send5s(rpcid.RpcAoiNiceGrid, &aoiInfo)
}

// center 为格子索引坐标，X/Y 是整型格子下标
func (s *Service) get9GridViewRoles(center model.Pos) map[uint64]struct{} {
	viewSet := make(map[uint64]struct{})
	// 遍历3*3九宫偏移
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			gx := center.X + dx
			gy := center.Y + dy
			// 边界防护：过滤地图范围外的格子，避免无效查询
			if gx < 0 || gy < 0 || gx >= s.MaxGridX || gy >= s.MaxGridY {
				continue
			}
			grid := model.Pos{
				X: gx,
				Y: gy,
			}
			cell, ok := s.Grids[grid]
			if !ok {
				continue
			}
			// 收集当前格子内所有角色
			for rid := range cell.SceneRoles {
				viewSet[rid] = struct{}{}
			}
		}
	}
	return viewSet
}

func (s *Service) sendRoleViewList(state *State, sceneRole *model.SceneRole, isUpdateGrid bool) {
	//获取当前九宫格玩家
	viewRoleIDs := s.get9GridViewRoles(sceneRole.GridPos)

	//其他人发新增
	msg := &pb.RoleViewListS2C{
		Type:       def.RoleViewTypeAdd,
		SceneId:    s.ID,
		MapId:      s.MapID,
		SceneRoles: []*pb.SceneRole{sceneRole.Pb()},
	}
	msgSceneRoles := make([]*pb.SceneRole, 0, len(viewRoleIDs))
	for viewRoleID := range viewRoleIDs {
		viewSceneRole, ok := state.sceneRoles[viewRoleID]
		if !ok {
			continue
		}
		msgSceneRoles = append(msgSceneRoles, viewSceneRole.Pb())
		//玩家更新格子才通知其他人
		if viewRoleID != sceneRole.RoleID && isUpdateGrid {
			viewSceneRole.Conn.SendMsg(msg)
		}
	}

	//玩家自己发全部，
	sceneRole.Conn.SendMsg(&pb.RoleViewListS2C{
		Type:       def.RoleViewTypeAll,
		SceneId:    s.ID,
		MapId:      s.MapID,
		SceneRoles: msgSceneRoles,
	})
}
