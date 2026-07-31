# M0 阶段 Gate 报告

## 报告控制

- 报告 ID：`M0-STAGE-GATE-待分配`
- 报告版本：`1.0`
- `StageGatePolicyVersion`：待填写
- 证据窗口：待填写 UTC 起止时间
- 候选提交：待填写完整 commit SHA
- 报告状态：`DRAFT`
- Gate 结论：`INSUFFICIENT_EVIDENCE`

本文件是 P0-013 的可填写 Gate 工件，不是签署记录，也不冻结首个迭代 Backlog。根据 [技术方案第 13 章](../VeritasQuantTechSpec.md#13-分阶段实施路线与验收结果)，最终报告必须记录政策版本、证据窗口、样本量、指标与阈值、测试报告哈希、审批签名以及唯一的 `PASS`、`FAIL` 或 `INSUFFICIENT_EVIDENCE` 结论。缺少任何强制证据时不得填写 `PASS`。

## 审查范围与排除项

- 范围：M0/P0 工程基线、治理记录和首个迭代 Backlog 冻结条件。
- 不在范围：环境晋级、Release 创建、模拟盘、券商连接、实盘授权或凭据。
- 代表资产和阶段边界：见 [P0-001 范围记录](P0-001ScopeAndGateRecord.md)。
- 当前预审结论和已知阻断：见 [M0PreReview.md](M0PreReview.md)。正式 Gate 审查须先更新预审，而不能以本模板替代预审结论。

## 强制检查表

| 检查 | 所需客观证据 | 当前状态 | 独立审阅结论与引用 |
| --- | --- | --- | --- |
| P0-001 范围、非目标与未来 Gate 签署名单 | [P0-001 范围记录](P0-001ScopeAndGateRecord.md)；非作者 PO/TL 审阅记录 | 待独立审阅 | 待填写 |
| P0-002 RACI、升级和值班替补 | [P0-002 RACI](P0-002RaciAndEscalation.md)；[独立角色证据包](M0IndependentReviewEvidence.md) 中的评审者与 IC 替补确认 | 待独立审阅 | 待填写 |
| P0-003 文档可定位性 | [文档索引](DevelopmentDocumentIndex.md)；已完成的 [定位演练记录](P0-003DocumentDiscoveryDrill.md) | 待独立演练 | 待填写 |
| P0-004 方案 A 目录边界 | [工作项证据](WorkItemRegister.yml) 中 P0-004 的脚本、契约测试和 Linux 验证；非作者 TL 审阅 | 待独立审阅 | 待填写 |
| P0-005 Python 3.13+ 与锁定策略 | [依赖策略](P0-005DependencyPolicy.md)；双平台 CI 工件和独立 QA 记录 | 待独立审阅 | 待填写 |
| P0-006 CI 合并治理 | [CI 治理记录](P0-006CiGovernance.md)；管理员分支保护审计链接；三项必需检查配置证据 | 阻断 | 待填写 |
| P0-007 违规阻断与定位 | [P0 独立验收包](P0IndependentAcceptancePackage.md) 中的受保护远程 CI 负例证据 | 阻断 | 待填写 |
| P0-008 Compose 演练 | [Compose 技术演练](P0-008ComposeDrillEvidence.md)；独立 SRE 复核 | 待独立审阅 | 待填写 |
| P0-009 测试证据格式 | [P0 独立验收包](P0IndependentAcceptancePackage.md) 中的 JUnit、coverage、环境和哈希复核 | 待独立审阅 | 待填写 |
| P0-010 登记表 | [登记表规范](P0-010RegisterSchema.md)；六类登记表；非作者 TL/QA 复核 | 待独立审阅 | 待填写 |
| P0-011 安全与许可证 | [P0 独立验收包](P0IndependentAcceptancePackage.md) 中的 Security baseline、负例和许可证复核 | 待独立审阅 | 待填写 |
| P0-012 追踪矩阵 | [P0 独立验收包](P0IndependentAcceptancePackage.md) 中的映射复核 | 待独立审阅 | 待填写 |
| 开放风险、事故、变更和行动项 | [风险](RiskRegister.yml)、[事故](IncidentRegister.yml)、[变更](ChangeRegister.yml)、[行动项](ActionRegister.yml)；阻断项为零或有已批准且未过期的例外 | 阻断 | 待填写 |

## 必须附加的治理证据

1. GitHub 管理员对默认分支保护、`Quality (ubuntu-latest)`、`Quality (windows-latest)` 和 `Security baseline` 必需检查的配置审计链接或截图。普通管理员绕过必须禁用；详见 `ACT-P0-004`。
2. 一次在受保护远程 CI 中执行的 P0-007 故意违规失败演练，须包含候选提交、Run URL、失败步骤、定位到的文件或字段、清理/回滚记录及恢复后 Run URL。
3. 至少一名未编写受审代码的人类评审者，以及一名 Incident Commander 替补的目录引用、确认 UTC 时间、审阅范围和 Gate 回避规则；填写 [独立角色证据包](M0IndependentReviewEvidence.md)。
4. 独立 QA 与 SRE 对各自证据包的实际复核运行或可重复审阅记录。作者测试、AI 代理执行或自我复核均不计入独立验收。

## 指标、样本和工件哈希

| 项目 | 阈值/要求 | 实际值 | 证据哈希或不可变链接 |
| --- | --- | --- | --- |
| 双平台 CI | Windows/Linux Python 3.13 质量任务与 Security baseline 成功 | 待填写 | 待填写 |
| CI 合并阻断 | 三项必需检查已启用，普通管理员绕过已禁用 | 待填写 | 待填写 |
| P0-007 负例 | 受保护 CI 明确失败并定位违规 | 待填写 | 待填写 |
| P0-003 定位演练 | 非作者在 15 分钟内完成全部必达项 | 待填写 | 待填写 |
| Compose 演练 | 启动、健康、停止与无遗留清理由独立 SRE 复核 | 待填写 | 待填写 |
| 测试证据 | JUnit、coverage、环境、种子和工件 SHA-256 齐全 | 待填写 | 待填写 |
| 开放阻断项 | 0 | 待填写 | 待填写 |

## 风险、例外与决议

| 记录 ID | 当前状态 | Gate 影响 | 决议/批准人/到期时间 |
| --- | --- | --- | --- |
| `RSK-P0-001` | 待填写 | 人员独立性与值班替补 | 待填写 |
| `RSK-P0-002` | 待填写 | 依赖、CI 与独立验收 | 待填写 |
| `RSK-P0-003` | 待填写 | 分支保护与必需检查 | 待填写 |
| `RSK-P0-004` | 待填写 | Compose 独立 SRE 复核 | 待填写 |
| `ACT-P0-004` | 待填写 | 必须关闭 | 待填写 |
| `ACT-P0-005` | 待填写 | 必须关闭 | 待填写 |
| `ACT-P0-007` | 待填写 | 必须关闭 | 待填写 |

任何例外都不得把缺少的 M0 强制证据转换为 `PASS`。若有开放阻断项，Gate 结论保持 `INSUFFICIENT_EVIDENCE` 或按实际失败填写 `FAIL`。

## 审批与结论

| 角色 | 人员与目录引用 | 审阅范围 | 独立性声明 | 签署 UTC 时间 | 签署/结论 |
| --- | --- | --- | --- | --- | --- |
| PO | 待填写 | 范围、非目标、风险接受与 Backlog 冻结 | 待填写 | 待填写 | 待填写 |
| TL | 待填写 | P0-001、P0-003、P0-004、P0-010、P0-012 | 待填写 | 待填写 | 待填写 |
| QA | 待填写 | P0-003、P0-005、P0-007、P0-009、P0-012 | 待填写 | 待填写 | 待填写 |
| SRE/安全 | 待填写 | P0-006、P0-008、P0-011 | 待填写 | 待填写 | 待填写 |

- 最终 Gate 结论：`INSUFFICIENT_EVIDENCE`
- 结论时间：待填写 UTC 时间
- Backlog 冻结引用：待填写；仅在最终结论为 `PASS` 后填写。
- 后续动作：完成全部阻断证据后更新 [M0PreReview.md](M0PreReview.md)，由人类 Gate 签署人重新执行本报告。
