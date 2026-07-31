# P0-010 登记表规范

`WorkItemRegister.yml`、`BugRegister.yml`、`RiskRegister.yml`、`IncidentRegister.yml`、`ChangeRegister.yml` 与 `ActionRegister.yml` 是当前版本化登记表。每个文件均声明 `RegistryVersion`、`RecordType` 和 `RequiredFields`；记录 ID 在本表范围内不可复用。

状态机以 [VeritasQuantDevelopmentWorkflow.md](../VeritasQuantDevelopmentWorkflow.md) 第 4 节为唯一依据：工作项、缺陷、风险、事故、变更和行动项不能发明私有终态。每次状态变化均追加 `AuditHistory`，其中记录 UTC `Ts`、操作者、动作和证据位置。`Blocked` 仅是工作项状态标志，必须填写原因、阻断方、下一动作、责任人和升级时间。

尚未有外部基线日期时，`TargetDate`、`DueDate` 与 `EscalateTs` 显式标为“待指定”，不得伪造日期或将其解释为已批准时限。首次把登记表迁移到工作项平台时，必须保留本文件的 ID、审计历史和引用关系。
