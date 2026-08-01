# M0 独立角色与审阅证据包

## 使用边界

本证据包用于关闭 `ACT-P0-007`，并支持 P0-001、P0-002、P0-004、P0-008、P0-010 和 P0-013 的独立审阅。填写者必须是人类；开发执行代理、作者测试执行者及受审代码作者不得将本文件作为独立验收或 Gate 签署证据。

填写时不得记录个人电话、令牌、凭据或敏感身份信息。目录引用应使用企业目录、值班组别或工单链接。角色冲突规则以 [SingleAgentGovernanceMaterial.md](SingleAgentGovernanceMaterial.md) 和 [开发工作流第 7、11 步](../VeritasQuantDevelopmentWorkflow.md) 为准。

## 人员与独立性确认

| 角色 | 姓名/稳定别名 | 企业目录或值班组引用 | 不得审阅的作者范围 | 确认 UTC 时间 | 人类确认方式 |
| --- | --- | --- | --- | --- | --- |
| 非作者评审者 | ACANX | `245818784@qq.com`（企业目录引用为 git 提交身份） | P0-003 文档索引/演练记录（由 AI 代理生成，本人未编写） | 2026-08-01T16:41:00Z | 独立执行定位演练并签署执行表 `P0-003HumanDrillWorksheet.md` |
| 独立 QA | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 |
| 独立 SRE/安全 | 待填写 | 待填写 | 待填写 | 待填写 | 待填写 |
| Incident Commander 替补 | 待填写 | 待填写 | 不适用 | 待填写 | 待填写 |

每名审阅者应明确确认：其未编写其所验收的代码或证据；如存在交叉贡献，须写明范围并由另一名合格非作者承担该范围的独立验收。

## Incident Commander 替补确认

| 字段 | 填写内容 |
| --- | --- |
| 主 Incident Commander | 待填写；应与 [P0-002 RACI](P0-002RaciAndEscalation.md) 一致 |
| 替补 Incident Commander | 待填写 |
| 值班窗口与时区 | 待填写 |
| S0/S1 触发后的首个动作 | 停止相关新风险工作、保全证据、创建或更新 `IncidentRegister.yml` 记录 |
| 升级路径确认 | 待填写；应与 P0-002 的 5/15 分钟路径一致 |
| 替补确认 UTC 时间 | 待填写 |
| 确认记录链接 | 待填写 |

## 独立审阅清单

| 审阅项 | 审阅角色 | 需检查的来源 | 结论：通过/需修改/证据不足 | 证据链接或哈希 | 审阅 UTC 时间 |
| --- | --- | --- | --- | --- | --- |
| P0-001 范围、非目标与 Gate 名单 | 非作者 PO/TL | [P0-001 范围记录](P0-001ScopeAndGateRecord.md)、技术方案第 13/14 章 | 待填写 | 待填写 | 待填写 |
| P0-002 RACI、唯一 Accountable 与升级路径 | 非作者 PO/TL | [P0-002 RACI](P0-002RaciAndEscalation.md)、本文件的替补确认 | 待填写 | 待填写 | 待填写 |
| P0-003 文档索引与定位演练 | 非作者 QA/TL | [文档索引](DevelopmentDocumentIndex.md)、[定位演练](P0-003DocumentDiscoveryDrill.md) | 通过（ACANX 独立执行定位演练，10/10 正确，见 `P0-003HumanDrillWorksheet.md`；独立角色边界是否满足由独立验收最终判定） | [P0-003HumanDrillWorksheet.md](P0-003HumanDrillWorksheet.md) | 2026-08-01T16:41:00Z |
| P0-004 目录边界与 Linux 证据 | 非作者 TL | `WorkItemRegister.yml` 的 P0-004、[Linux 证据](M0LinuxValidationEvidence.md) | 待填写 | 待填写 | 待填写 |
| P0-008 Compose 启停、健康与清理 | 独立 SRE | [Compose 演练](P0-008ComposeDrillEvidence.md)、`Docker/docker-compose.yml`、`Docker/DevelopmentEnvironment.md` | 待填写 | 待填写 | 待填写 |
| P0-010 登记表字段与审计关系 | 非作者 QA/TL | 六类登记表、[登记表规范](P0-010RegisterSchema.md) | 待填写 | 待填写 | 待填写 |
| P0-006/007/009/011/012 支撑证据 | 独立 QA/SRE | [P0 独立验收包](P0IndependentAcceptancePackage.md) | 待填写 | 待填写 | 待填写 |

## Gate 回避与签署限制

1. 任何人不得对自己编写的代码、自己执行的作者测试或自己配置的管理员控制给出独立验收。
2. 发现审阅范围冲突时，该范围必须转交另一名合格非作者；本次审阅记录应标记为“证据不足”，不得以口头说明替代。
3. 非作者审阅通过不等于 M0 `PASS`。仅 PO、TL、QA 和 SRE/安全的 Gate 签署人在所有强制项齐全后，才能在 [M0 阶段 Gate 报告](M0StageGateReport.md) 中填写最终结论。
4. 在 M0 `PASS` 前，不创建 Release、不冻结正式迭代 Backlog、不进行环境晋级，也不启用实盘。

## 完成记录

- `ACT-P0-007` 状态：`OPEN`
- 本证据包状态：`DRAFT`
- 当前结论：尚无人类独立性、替补确认或审阅结果，不能关闭行动项或改变任何 P0 工作项状态。
