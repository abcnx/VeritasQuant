# P2-040 连续运行模拟盘（观测接线能力）— 证据

- **任务：** P2-040（ISSUE #169）
- **日期：** 2026-08-03
- **实现：** BeeAgent
- **PR：** 待第八批 PR 合并后回填

## 验收标准对照

| 验收标准 | 实现 | 证据 |
|----------|------|------|
| 每日运行清单 | `RuntimeSnapshotV1` 快照（指标文本 + SLO 摘要 + 时间） | `reliability/ObservabilityWiring.py` |
| 对账 | 由 P2-042 WeeklyReviewerV1 承接（观测接线提供指标输入） | `reliability/WeeklyReview.py` |
| SLO | `SloObservationRecorderV1` 按日记录 SLI 观测 | `ObservabilityWiring.py` |
| 风险/事故记录 | 由 P2-041 CrashDrillEvidenceWindowV1 承接 | `reliability/CrashDrill.py` |

## 技术方案

解决 P2-036/P2-037 遗留开放项（运行期埋点接线）：
- **ObservabilityWiringV1**：运行入口一键装配（MetricsCollector + SloCalculator + 探针），
  统一 runId/执行模式；
- **InstrumentedLedgerStore**：包装 LedgerStoreV1，每次提交观测账本延迟并计数
  （账本事务提交 SLI 埋点）；
- **InstrumentedGroupWorker**：包装 AccountGroupWorkerV1，处理事件时观测
  提交耗时（事件延迟 SLI 埋点）；
- **SloObservationRecorderV1**：按日把观测写入 SLO 计算器（P2-037 能力）；
- **RuntimeSnapshotV1**：指标文本 + SLO 摘要快照，供每日运行清单旁路写入
  （TechSpec 3.1 虚线旁路，不修改交易状态）。

## 测试

`tests/unit/reliability/test_observability_wiring.py`（6 用例）：
- wiring 装配、快照含指标与 SLO；
- 账本提交延迟/计数埋点、内层真实提交；
- 事件处理延迟埋点、handler 正常调用；
- SLO 观测记录与评估、传播到计算器。

## 验证结果

- ruff / mypy / Preflight：通过
- 全量 pytest：待第八批 PR 后确认

## 风险与开放项

- 实际 60 个有效交易日运行需要模拟盘环境与行情数据源持续运行；
  本任务提供运行期观测接线（数据采集侧），运行证据在模拟盘部署后补充。
