## 性能优化记录

### 2026-07-01: 战斗帧计算内存优化

**问题**：
- `TimerFrameCalculation` 占总分配的 76%（1.17GB）
- 每帧调用 `make([]*rpc.Msg, 0, cap)` 分配新底层数组
- 每帧创建临时 Map `make(map[uint64]model.BattleRole)`

**优化方案**：
- `s.msgCache = s.msgCache[:0]` 复用切片

**优化效果**：
- 存活内存: 1.5GB+ → 12.5MB
- GC 停顿: 明显降低
- 帧计算性能: 提升

**监控指标**：
- PPROF: `http://localhost:6060/debug/pprof/heap`
- 预期存活内存: < 50MB (空闲时)


**问题**：
- `rpc.StartWithCfg和sys.SafeSendTimeout` 占总分配的 26.74%（15MB）
- 热路径使用 `<-time.After()` 频繁创建 Timer

**优化方案**：
- `timer:=time.NewTick();` 复用切片
- 改用 `timer:=time.NewTick()` 复用 Timer

**优化效果**：
- 存活内存: 15MB → 0MB
- GC 停顿: 明显降低
- 帧计算性能: 提升

**监控指标**：
- PPROF: `http://localhost:6060/debug/pprof/heap`
- 预期存活内存: < 50MB (空闲时)
