package rolesvc

import (
	"cake/internal/game/def"
	"cake/internal/game/repo/role"
	"cake/internal/pkg/logger"
	"time"
)

func (s *Service) registerTimer() {
	s.AddTimer("TimerCheckHeartbeat", def.HeartbeatInterval*time.Second, -1, s.TimerCheckHeartbeat, nil)
	s.AddTimer("TimerSaveRoleDB", def.SaveRoleDBInterval*time.Second, -1, s.TimerSaveRoleDB, nil)
}

func (*Service) TimerSaveRoleDB(rawRole any, _ any) error {
	state := rawRole.(*State)
	if !state.IsSave() {
		return nil
	}
	//定点入库
	err := role.Repo().UpdateRole(state.RolePO)
	if err != nil {
		logger.Errorf("[%d]定点存数据失败%v, %v", state.RoleID, state.RolePO, err)
		return err
	}
	state.RestSave()
	return nil
}

func (*Service) TimerCheckHeartbeat(rawRole any, _ any) error {
	state := rawRole.(*State)
	if state.Conn == nil {
		return nil
	}
	nowTime := time.Now().Unix()
	heartbeat := state.Heartbeat
	if heartbeat.LastTime <= 0 {
		heartbeat.LastTime = nowTime
	}
	if nowTime-heartbeat.LastTime >= def.HeartbeatInterval+2 {
		heartbeat.BadCnt++
		if heartbeat.BadCnt >= 6 {
			//连续差不多1分钟心跳超时了，直接踢吧
			state.CloseConn()
		}
	} else {
		heartbeat.BadCnt = 0
	}
	state.Heartbeat = heartbeat
	logger.Debugf("心跳%+v", state.Heartbeat)

	return nil
}

func (s *Service) TimerStopRole(rawRole any, _ any) error {
	state := rawRole.(*State)
	if !state.IsClosed() {
		return nil
	}
	mgr.StopRole(s.RoleID)
	return nil
}
