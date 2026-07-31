# 开发文档索引

新成员应从本页定位当前权威设计、计划、治理记录、风险和证据。技术行为以技术方案为唯一事实来源；归档文档不作为设计依据。

| 类别 | 位置 | 用途 |
| --- | --- | --- |
| 权威技术设计 | [VeritasQuantTechSpec.md](../VeritasQuantTechSpec.md) | 事件、账户、风控、执行、API、配置和验收契约。 |
| 开发计划 | [VeritasQuantDevelopmentPlan.md](../VeritasQuantDevelopmentPlan.md) | P0-P6 任务、DoR/DoD、里程碑和风险基线。 |
| 开发工作流 | [VeritasQuantDevelopmentWorkflow.md](../VeritasQuantDevelopmentWorkflow.md) | 状态机、评审、验证、Gate 和证据规则。 |
| ADR 索引 | [AdrIndex.md](../Adr/AdrIndex.md) | 架构决策记录；当前尚无已批准 ADR。 |
| 范围/Gate 待签记录 | [P0-001ScopeAndGateRecord.md](P0-001ScopeAndGateRecord.md) | 代表资产、非目标和待签署人。 |
| RACI 与升级 | [P0-002RaciAndEscalation.md](P0-002RaciAndEscalation.md) | 角色责任、联系人和事故升级路径。 |
| 工作项登记 | [WorkItemRegister.yml](WorkItemRegister.yml) | P0 工作项状态、依赖、证据和阻断。 |
| 风险登记 | [RiskRegister.yml](RiskRegister.yml) | 当前风险、责任、缓解与升级。 |
| 缺陷/事故/变更/行动项 | [BugRegister.yml](BugRegister.yml)、[IncidentRegister.yml](IncidentRegister.yml)、[ChangeRegister.yml](ChangeRegister.yml)、[ActionRegister.yml](ActionRegister.yml) | 正式登记模板与审计历史。 |
| 阶段追踪矩阵 | `TraceabilityMatrix.yml` | 技术契约、任务、测试和 Gate 映射。 |
| P0 工程实施记录 | [P0-003-P0-012ImplementationEvidence.md](P0-003-P0-012ImplementationEvidence.md) | P0 作者实施边界、工件和未决证据。 |
| 单 Agent 治理材料 | [SingleAgentGovernanceMaterial.md](SingleAgentGovernanceMaterial.md) | ClaudeCode 的执行边界、待确认人类角色与最小确认模板。 |
| P0 测试证据 | [P0-003-P0-012TestEvidence.json](P0-003-P0-012TestEvidence.json) | 机器可读 JUnit、coverage、环境、种子和工件哈希。 |
| CI 治理记录 | [P0-006CiGovernance.md](P0-006CiGovernance.md) | 必需检查、证据保留和待管理员配置。 |
| 测试证据规范 | [TestEvidencePolicy.md](TestEvidencePolicy.md) | 稳定测试 ID、JUnit、覆盖率、种子和哈希规范。 |
| 安全许可证策略 | [../../Configs/Security/LicensePolicy.yml](../../Configs/Security/LicensePolicy.yml) | 许可证白名单及漏洞处置 SLA；当前待批准。 |
| Docker 开发依赖 | [../../Docker/DevelopmentEnvironment.md](../../Docker/DevelopmentEnvironment.md) | 临时 PostgreSQL/Redis 的启动和清理说明。 |
| P1 实现证据 | [P1-001-P1-013Evidence.md](P1-001-P1-013Evidence.md) | P1 作者测试、工件和残余风险。 |

所有新增开发过程文档应保存于 `Docs/DevelopmentWorkflow/`；设计决策变更仍必须同步修改技术方案。
