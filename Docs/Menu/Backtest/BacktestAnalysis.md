# 回测分析（BacktestAnalysis）— 菜单业务文档

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/Backtest/BacktestAnalysis.md`
> 对应视图：`Web/src/views/BacktestAnalysisView.vue`
> 接口契约：见 [Docs/API/Backtest/BacktestRunQuery.md](../../API/Backtest/BacktestRunQuery.md)

## 1. 菜单入口

- **菜单位置**：左侧导航栏 → 回测分析
- **菜单名称**：回测分析
- **菜单 key**：`backtest-analysis`
- **菜单图标**：`mdi-chart-timeline-variant`
- **URL 路由**：`/meta/finvquant/backtest/analysis`
- **对应视图组件**：`Web/src/views/BacktestAnalysisView.vue`
- **支持 URL 参数**：`?run_id=xxx`（自动打开指定任务的报告，由黄金期货回测验证页跳转携带）

## 2. 业务功能概述

回测任务列表 + 投资策略回测收益分析报告：

| 功能 | 说明 |
|------|------|
| 任务列表 | 按状态/标的/关键字过滤，分页展示（任务号/策略/标的/区间/周期/状态/进度） |
| 报告指标卡 | 12 项核心指标：期末总资产/总收益额/到期收益率/年化/最大回撤/夏普/胜率/盈亏比/最大投入/平均投入/持仓天数/交易笔数 |
| 四类曲线 | ① 账户余额（总资产+现金）② 投资收益率 ③ 累计收益额 ⑦ 持仓金额（ECharts，按报告精度） |
| ⑨ 链路追踪 | 事件统计卡片（触发/成交/拒绝/过期/平均委托耗时/未成交原因分布）+ 四页签明细（成交记录/资金流水/持仓变化/事件追踪） |
| 信号归因 | 各触发信号（买入信号/卖出信号/止损/止盈）的成交笔数分布 |

## 3. 使用方法

```
左侧列表选择 SUCCEEDED 任务（或从黄金期货回测验证页点击「查看报告」带 run_id 直达）
→ 查看指标卡与曲线 → 下方页签切换：成交记录 / 资金流水明细 / 持仓变化明细 / 事件追踪
→ 成交记录支持分页浏览
```

## 4. 处理逻辑

- 选择任务后并行拉取 6 类数据：

| 数据 | 接口 |
|------|------|
| 汇总报告 | `GET .../Backtest/Run/Report?run_id=` |
| 净值曲线 | `GET .../Backtest/Run/Equity?run_id=&page=1&page_size=5000` |
| 成交记录 | `GET .../Backtest/Run/Trades?run_id=`（分页） |
| 资金流水 | `GET .../Backtest/Run/Cashflows?run_id=&page=1&page_size=1000` |
| 持仓变化 | `GET .../Backtest/Run/PositionLogs?run_id=&page=1&page_size=1000` |
| 事件追踪 | `GET .../Backtest/Run/EventTraces?run_id=&page=1&page_size=1000` |

- ECharts 按报告精度生成 x 轴标签（Day=日期 / Hour=日期+时 / Min=日期+时分）；切换任务时 dispose 旧图实例；
- 报告指标语义（最大投入/平均投入（持仓期时间加权）/回撤区间/夏普（rf=0，按精度年化因子）等）
  见 [BacktestStrategySpec.md](../../DevSpec/BacktestStrategySpec.md) 第 7 章。

## 5. 注意事项

- 仅 SUCCEEDED 任务可打开报告；FAILED 任务在列表显示 error_message；
- 曲线点数与报告精度相关（Day 精度 8 年约 2000 点；Min 精度点数极大，建议用 Day/Hour 精度查看）；
- 报告数据持久化在 `finv_quant_backtest_run.report` 与 `finv_quant_backtest_equity` 等表，可随时回看。

## 6. 使用的后端接口索引

| 接口 | 接口文档 |
|------|----------|
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/List` | [BacktestRunQuery.md](../../API/Backtest/BacktestRunQuery.md) |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Report` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Equity` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Trades` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/Cashflows` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/PositionLogs` | 同上 |
| `GET /API/V1/Meta/FinvQuant/Backtest/Run/EventTraces` | 同上 |
