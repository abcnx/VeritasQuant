# P0-003 独立文档定位演练记录

## 演练控制

- 演练 ID：`P0-003-DOC-DISCOVERY-待分配`
- 工作项：`P0-003`
- 参与者：待填写；必须为未编写文档索引或受审治理材料的人类。
- 观察者：待填写；建议为 TL 或独立 QA。
- 开始时间：待填写 UTC 时间
- 结束时间：待填写 UTC 时间
- 总耗时：待填写；不得超过 15 分钟。
- 结论：`NOT_EXECUTED`

本记录验证 [文档索引](DevelopmentDocumentIndex.md) 满足 P0-003 的验收条件。它不检查交易行为、不会改变技术方案，也不代替 P0-013 的 M0 Gate 签署。

## 演练规则

1. 演练开始前，观察者只向参与者提供仓库根目录与本页；不得直接提供各答案的文件路径。
2. 参与者可使用仓库内搜索和文档链接，但不得向作者或开发执行代理询问答案。
3. 每个条目记录定位到的准确文件、章节或登记 ID，以及完成 UTC 时间。
4. 任一必达项无法在 15 分钟内定位、链接失效或将归档材料误认为权威设计时，结论为 `FAILED`；修复后须重新演练并使用新 ID。

## 必达定位项

| 序号 | 参与者需定位的事实 | 正确目标 | 参与者找到的位置 | 完成 UTC 时间 | 观察者判定 |
| ---: | --- | --- | --- | --- | --- |
| 1 | 唯一权威技术设计及其阶段 Gate 契约 | `Docs/VeritasQuantTechSpec.md` 第 13 章 | 待填写 | 待填写 | 待填写 |
| 2 | 当前 P0-P6 开发计划 | `Docs/VeritasQuantDevelopmentPlan.md` | 待填写 | 待填写 | 待填写 |
| 3 | 工作项状态机与独立验收规则 | `Docs/VeritasQuantDevelopmentWorkflow.md` 第 4、7、11 节 | 待填写 | 待填写 | 待填写 |
| 4 | P0 工作项、依赖和当前阻断项 | `Docs/DevelopmentWorkflow/WorkItemRegister.yml` | 待填写 | 待填写 | 待填写 |
| 5 | 当前风险、行动项和事故登记 | `RiskRegister.yml`、`ActionRegister.yml`、`IncidentRegister.yml` | 待填写 | 待填写 | 待填写 |
| 6 | P0-006 的 CI/分支保护待办 | `P0-006CiGovernance.md`、`ACT-P0-004` | 待填写 | 待填写 | 待填写 |
| 7 | P0-008 的 Compose 技术证据与独立 SRE 缺口 | `P0-008ComposeDrillEvidence.md`、`ACT-P0-005` | 待填写 | 待填写 | 待填写 |
| 8 | M0 当前预审结论和必须关闭的阻断 | `M0PreReview.md`、`M0StageGateReport.md` | 待填写 | 待填写 | 待填写 |
| 9 | 非作者评审与 IC 替补的填写位置 | `M0IndependentReviewEvidence.md`、`ACT-P0-007` | 待填写 | 待填写 | 待填写 |
| 10 | ADR 是否存在及其权威索引 | `Docs/Adr/AdrIndex.md` | 待填写 | 待填写 | 待填写 |

## 观察者结论

- 必达项通过数：`0/10`
- 耗时合格：待填写
- 是否误用 Archive：待填写
- 发现的问题与修复行动：待填写；问题应登记至 `BugRegister.yml`、`RiskRegister.yml` 或 `ActionRegister.yml`。
- 最终结论：`NOT_EXECUTED`
- 观察者姓名/目录引用：待填写
- 观察者确认 UTC 时间：待填写

只有结论为 `PASSED`、耗时不超过 15 分钟且由合格非作者确认时，P0-003 才具备进入独立验收的证据。该结果仍不自动将 P0-003 标为 `ACCEPTED`。
