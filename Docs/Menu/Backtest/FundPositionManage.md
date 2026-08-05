# 资金管理 / 持仓管理（FundManage · PositionManage）— 菜单业务文档

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/Backtest/`
> 对应视图：`Web/src/views/FundManageView.vue` · `Web/src/views/PositionManageView.vue`
> 接口契约：见 [Docs/API/Backtest/BacktestRunQuery.md](../../API/Backtest/BacktestRunQuery.md)

## 1. 菜单入口

| 菜单 | 位置 | key | 图标 | 路由 | 视图 |
|------|------|-----|------|------|------|
| 资金管理 | 左侧导航栏 → 资金管理 | `fund` | `mdi-cash-multiple` | `/meta/finvquant/fund` | `FundManageView.vue` |
| 持仓管理 | 左侧导航栏 → 持仓管理 | `position` | `mdi-briefcase-variant-outline` | `/meta/finvquant/position` | `PositionManageView.vue` |

## 2. 业务功能概述

| 菜单 | 功能 |
|------|------|
| 资金管理 | 选择已成功回测任务 → 查看现金与总资产曲线（ECharts，按报告精度）；资金流水的结构化数据可通过接口获取 |
| 持仓管理 | 选择已成功回测任务 → 查看持仓数量与持仓市值双轴曲线 + 开平仓记录（最多 500 条） |

## 3. 使用方法

```
选择回测任务（下拉，仅列出 SUCCEEDED）→ 自动加载并渲染曲线
持仓管理页下方表格展示开平仓记录（时间/方向/价格/数量/金额/信号）
```

## 4. 处理逻辑

| 数据 | 接口 |
|------|------|
| 任务列表 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/List?status=SUCCEEDED&page=1&page_size=100` |
| 资金/持仓曲线 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/Equity?run_id=&page=1&page_size=5000` |
| 开平仓记录 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/Trades?run_id=&page=1&page_size=500` |

- 资金曲线：`equity`（总资产=现金+持仓市值）与 `cash`（现金）双线；
- 持仓曲线：`position_value`（持仓市值）与 `position_qty`（持仓数量）双轴；
- x 轴标签按报告精度生成（Day=日期，Min/Hour=日期+时间）。

## 5. 注意事项

- 仅展示 SUCCEEDED 任务；曲线精度与任务创建时的 report_precision 一致；
- 更完整的资金流水/持仓变化/事件追踪明细见「回测分析」菜单的 ⑨ 链路追踪页签。

## 6. 使用的后端接口索引

| 接口 | 接口文档 |
|------|----------|
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/List` | [BacktestRunQuery.md](../../API/Backtest/BacktestRunQuery.md) |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Equity` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Trades` | 同上 |
