package scene

import (
	"cake/internal/game/def"
	"cake/internal/game/services/mapsvc/scene/iscene"
	"cake/internal/gensvc/rpc"
	"cake/internal/pkg/logger"
	"context"
	"github.com/puzpuzpuz/xsync/v3"
	"sync"
	"sync/atomic"
)

var (
	mgrOne     sync.Once
	defaultMgr *manager
)

type manager struct {
	wg        *sync.WaitGroup
	mu        sync.RWMutex
	cxt       context.Context
	cancel    context.CancelFunc
	nextSeq   uint32
	MapBases  map[uint32]*MapBase            // map[地图配置id]
	SceneSvcs *xsync.MapOf[uint32, *Service] //所有场景
	MainSvcs  map[uint32]*Service            //主场景信息 MapID 初始化后不能改
}

func Manager() *manager {
	mgrOne.Do(func() {
		cxt, cancel := context.WithCancel(context.Background())
		defaultMgr = &manager{
			MapBases:  make(map[uint32]*MapBase),
			SceneSvcs: xsync.NewMapOf[uint32, *Service](),
			MainSvcs:  make(map[uint32]*Service),
			wg:        &sync.WaitGroup{},
			cxt:       cxt,
			cancel:    cancel,
		}
	})
	return defaultMgr
}

func (m *manager) AddMapBase(mapID uint32) {
	m.MapBases[mapID] = NewMapBase(mapID)
}

func (m *manager) AddSceneSvc(sceneSvc *Service) {
	m.SceneSvcs.Store(sceneSvc.ID, sceneSvc)
	if sceneSvc.Type == def.MapTypeMain {
		m.MainSvcs[sceneSvc.MapID] = sceneSvc
	}
}

func (m *manager) RpcBySceneID(sceneID uint32) *rpc.Service {
	sceneSvc, ok := m.SceneSvcs.Load(sceneID)
	if !ok {
		return nil
	}
	return sceneSvc.Service
}

func (m *manager) RpcByMapID(mapID uint32) *rpc.Service {
	if sceneSvc, ok := m.MainSvcs[mapID]; ok {
		return sceneSvc.Service
	}
	return nil
}

func (m *manager) MapIdToSceneId(mapID uint32) uint32 {
	if sceneSvc, ok := m.MainSvcs[mapID]; ok {
		return sceneSvc.sceneID
	}
	return 0
}

func (m *manager) StartScene(mapID uint32) *rpc.Service {
	mapBase, ok := m.MapBases[mapID]
	if !ok {
		logger.Errorf("地图id不存在 %d", mapID)
		return nil
	}
	sceneID := m.GenID()
	cfg := rpc.NewCfg()
	cfg.Wg = m.wg
	cfg.Ctx = m.cxt
	cfg.SendMaxCap = 10000
	sceneRpc, sceneSvc, err := StartService(sceneID, cfg, mapBase)
	if err != nil {
		logger.Errorf("MapID进程启动失败 %d", mapID)
		return nil
	}
	m.AddSceneSvc(sceneSvc)
	return sceneRpc
}

func (m *manager) GenID() uint32 {
	return atomic.AddUint32(&m.nextSeq, 1)
}

func (m *manager) Stop() {
	m.cancel()
	m.wg.Wait()
}

var _ iscene.IManager = new(manager)
