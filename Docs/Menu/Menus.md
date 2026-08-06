# Web 前端业务菜单总览（Menus）

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/Menus.md`
> 适用范围：`Web/` 前端侧边导航（`Web/src/App.vue` 中 `menuItems` 定义）中的全部业务菜单索引。
> 菜单文档规范见 [Docs/DevSpec/MenuSpec.md](../DevSpec/MenuSpec.md)；接口文档见 [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md)。

## 菜单层级结构

```
仪表盘（/dashboard）
历史行情（分组）
  └─ 历史行情查询（/quote/query）
元数据管理（分组）
  └─ 业务元数据维护（二级分组）
      ├─ 交易所信息维护（/meta/exchange）
      ├─ 交易所下设市场信息维护（/meta/market）
      ├─ 规范证券信息维护（/meta/security）
      └─ 历史行情数据导入（/meta/import）
量化策略验证（分组）
  ├─ 黄金期货合约回测验证（/meta/finvquant/backtest/gold-futures）
  └─ 环境与模板管理（/meta/finvquant/env-template）
账户管理（/meta/finvquant/account）
资金管理（/meta/finvquant/fund）
持仓管理（/meta/finvquant/position）
策略管理（/meta/finvquant/strategy）
回测分析（/meta/finvquant/backtest/analysis）
仿真数据验证（/meta/finvquant/simulation/data）
模拟盘验证（/meta/finvquant/simulation/paper）
实盘仿真验证（/meta/finvquant/simulation/live-sim）
实盘交易（/meta/finvquant/live-trading）
```

> 每个叶子菜单对应独立 URL 路由路径（见 `Web/src/router.ts`）；侧边导航支持两级分组展开。
> **路径规范**：量化回测模块菜单/路由统一加 `Meta/FinvQuant/` 前缀（`/meta/finvquant/...`），
> 与后端 `/API/V1/Meta/FinvQuant/Backtest/**` 接口路径保持同一前缀体系。

## 菜单清单

| # | 菜单名称 | 层级 | key | 图标 | 路由 | 对应视图组件 | 使用的后端接口 | 菜单文档 |
|---|----------|------|-----|------|------|--------------|----------------|----------|
| 1 | 仪表盘 | 一级叶子 | `dashboard` | `mdi-view-dashboard` | `/dashboard` | `Web/src/views/DashboardView.vue` | `GET /API/V1/health/live`（存活探针） | — |
| 2 | 历史行情 | 一级分组 | `history-quote` | `mdi-chart-box` | — | — | — | — |
| 3 | 历史行情查询 | 二级叶子 | `quote-query` | `mdi-chart-line` | `/quote/query` | `Web/src/views/QuoteQueryView.vue` | `GET /API/V1/Quote/Query`、`GET /API/V1/Meta/FinvQuant/Metadata/Security/Options` | [HistoryQuote/HistoryQuoteQuery.md](HistoryQuote/HistoryQuoteQuery.md) |
| 4 | 元数据管理 | 一级分组 | `metadata` | `mdi-database-cog` | — | — | — | — |
| 5 | 业务元数据维护 | 二级分组 | `meta-maintenance` | `mdi-database-search` | — | — | — | — |
| 6 | 交易所信息维护 | 三级叶子 | `meta-exchange` | `mdi-office-building` | `/meta/exchange` | `Web/src/views/MetaExchangeView.vue` | `GET /API/V1/Meta/FinvQuant/Metadata/Exchange/List`、`POST /API/V1/Meta/FinvQuant/Metadata/Exchange/Save`、`POST /API/V1/Meta/FinvQuant/Metadata/Exchange/Toggle` | [Meta/MetaExchange.md](Meta/MetaExchange.md) |
| 7 | 交易所下设市场信息维护 | 三级叶子 | `meta-market` | `mdi-chart-areaspline` | `/meta/market` | `Web/src/views/MetaMarketView.vue` | `GET /API/V1/Meta/FinvQuant/Metadata/Market/List`、`POST /API/V1/Meta/FinvQuant/Metadata/Market/Save`、`POST /API/V1/Meta/FinvQuant/Metadata/Market/Toggle` | [Meta/MetaMarket.md](Meta/MetaMarket.md) |
| 8 | 规范证券信息维护 | 三级叶子 | `meta-security` | `mdi-tag-multiple` | `/meta/security` | `Web/src/views/MetaSecurityView.vue` | `GET /API/V1/Meta/FinvQuant/Metadata/Security/List`、`POST /API/V1/Meta/FinvQuant/Metadata/Security/Save`、`POST /API/V1/Meta/FinvQuant/Metadata/Security/Toggle`、`GET /API/V1/Meta/FinvQuant/Metadata/Security/Options`、`GET /API/V1/Meta/FinvQuant/Metadata/Security/Lookup` | [Meta/MetaSecurity.md](Meta/MetaSecurity.md) |
| 9 | 历史行情数据导入 | 三级叶子 | `meta-import` | `mdi-database-import` | `/meta/import` | `Web/src/views/QuoteImportView.vue` | `POST /API/V1/Quote/Import/Upload`、`GET /API/V1/Meta/FinvQuant/Metadata/Security/Options`、`GET /API/V1/Meta/FinvQuant/Metadata/Security/Lookup` | [HistoryQuote/HistoryQuoteImport.md](HistoryQuote/HistoryQuoteImport.md) |
| 10 | 量化策略验证 | 一级分组 | `quant-backtest` | `mdi-flask-outline` | — | — | — | — |
| 11 | 黄金期货合约回测验证 | 二级叶子 | `backtest-gold-futures` | `mdi-chart-bell-curve` | `/meta/finvquant/backtest/gold-futures` | `Web/src/views/BacktestGoldFuturesView.vue` | `GET .../Backtest/Strategy/List`、`GET .../Backtest/Account/List`、`GET .../Backtest/Environment/List`、`POST .../Backtest/Run/Create`、`GET .../Backtest/Run/List` | [Backtest/BacktestGoldFutures.md](Backtest/BacktestGoldFutures.md) |
| 12 | 环境与模板管理 | 二级叶子 | `env-template` | `mdi-application-cog-outline` | `/meta/finvquant/env-template` | `Web/src/views/EnvironmentTemplateView.vue` | `GET/POST .../Backtest/Environment/*`、`GET/POST .../Backtest/Template/*` | [Backtest/EnvironmentTemplate.md](Backtest/EnvironmentTemplate.md) |
| 13 | 账户管理 | 一级叶子 | `account` | `mdi-account-cog-outline` | `/meta/finvquant/account` | `Web/src/views/AccountManageView.vue` | `GET/POST .../Backtest/Account/*` | [Backtest/AccountManage.md](Backtest/AccountManage.md) |
| 14 | 资金管理 | 一级叶子 | `fund` | `mdi-cash-multiple` | `/meta/finvquant/fund` | `Web/src/views/FundManageView.vue` | `GET .../Backtest/Run/List`、`GET .../Backtest/Run/Equity` | [Backtest/FundManage.md](Backtest/FundManage.md) |
| 15 | 持仓管理 | 一级叶子 | `position` | `mdi-briefcase-variant-outline` | `/meta/finvquant/position` | `Web/src/views/PositionManageView.vue` | `GET .../Backtest/Run/List`、`GET .../Backtest/Run/Equity`、`GET .../Backtest/Run/Trades` | [Backtest/PositionManage.md](Backtest/PositionManage.md) |
| 16 | 策略管理 | 一级叶子 | `strategy` | `mdi-sitemap-outline` | `/meta/finvquant/strategy` | `Web/src/views/StrategyManageView.vue` | `GET/POST .../Backtest/Strategy/*` | [Backtest/StrategyManage.md](Backtest/StrategyManage.md) |
| 17 | 回测分析 | 一级叶子 | `backtest-analysis` | `mdi-chart-timeline-variant` | `/meta/finvquant/backtest/analysis` | `Web/src/views/BacktestAnalysisView.vue` | `GET .../Backtest/Run/List`、`GET .../Backtest/Run/Report`、`GET .../Backtest/Run/Equity`、`GET .../Backtest/Run/Trades`、`GET .../Backtest/Run/Cashflows`、`GET .../Backtest/Run/PositionLogs`、`GET .../Backtest/Run/EventTraces` | [Backtest/BacktestAnalysis.md](Backtest/BacktestAnalysis.md) |
| 18 | 仿真数据验证 | 一级叶子 | `simulation-data` | `mdi-database-sync-outline` | `/meta/finvquant/simulation/data` | `Web/src/views/PlaceholderView.vue` | —（规划中） | [Backtest/SimulationData.md](Backtest/SimulationData.md) |
| 19 | 模拟盘验证 | 一级叶子 | `simulation-paper` | `mdi-account-cash-outline` | `/meta/finvquant/simulation/paper` | `Web/src/views/PlaceholderView.vue` | —（规划中） | [Backtest/SimulationPaper.md](Backtest/SimulationPaper.md) |
| 20 | 实盘仿真验证 | 一级叶子 | `simulation-live-sim` | `mdi-robot-outline` | `/meta/finvquant/simulation/live-sim` | `Web/src/views/PlaceholderView.vue` | —（规划中） | [Backtest/SimulationLiveSim.md](Backtest/SimulationLiveSim.md) |
| 21 | 实盘交易 | 一级叶子 | `live-trading` | `mdi-cash-register` | `/meta/finvquant/live-trading` | `Web/src/views/PlaceholderView.vue` | —（规划中） | [Backtest/LiveTrading.md](Backtest/LiveTrading.md) |

## 说明

- 菜单定义与视图组件一一对应，全部在 `Web/src/App.vue` 的 `menuItems` 中登记；新增菜单时须同步更新本索引。
- 图标统一使用 Material Design Icons（`@mdi/font`）；**注意：`mdi-chart-candlestick` 在 @mdi/font 7.4.47 中不存在**，K 线类菜单使用 `mdi-chart-line` / `mdi-chart-box`。
- 三个字典维护菜单（交易所/市场/证券）的 List 分页：**启用的（flag_enable='1'）优先展示，禁用的排后面**，同状态按 code 升序（见 `internal/meta/service.go`）。
- 「菜单文档」列指向对应业务菜单介绍文档（`Docs/Menu/xxx/XXXX.md`，见 [MenuSpec.md](../DevSpec/MenuSpec.md)）；暂无文档的菜单以 `—` 标注，待补充。
- 「使用的后端接口」列与对应 API 接口文档的「已使用位置登记」互为索引（见 [ApiSpec.md](../DevSpec/ApiSpec.md) 第 7 节）。
- 视图组件名称与菜单 key 的对应关系：

| 菜单 key | 视图组件 |
|----------|----------|
| `dashboard` | `DashboardView.vue` |
| `quote-query` | `QuoteQueryView.vue` |
| `meta-exchange` | `MetaExchangeView.vue` |
| `meta-market` | `MetaMarketView.vue` |
| `meta-security` | `MetaSecurityView.vue` |
| `meta-import` | `QuoteImportView.vue` |
| `backtest-gold-futures` | `BacktestGoldFuturesView.vue` |
| `env-template` | `EnvironmentTemplateView.vue` |
| `account` | `AccountManageView.vue` |
| `fund` | `FundManageView.vue` |
| `position` | `PositionManageView.vue` |
| `strategy` | `StrategyManageView.vue` |
| `backtest-analysis` | `BacktestAnalysisView.vue` |
| `simulation-data` / `simulation-paper` / `simulation-live-sim` / `live-trading` | `PlaceholderView.vue`（占位，四个菜单共用） |

## 相关文档

- [Docs/DevSpec/MenuSpec.md](../DevSpec/MenuSpec.md) — 前端菜单开发规范
- [Docs/API/README.md](../API/README.md) — 服务端 API 文档入口
- [Docs/API/APIs.md](../API/APIs.md) — 服务端 API 总览（接口清单索引）
