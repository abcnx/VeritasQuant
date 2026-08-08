# FinvQuant

**量化策略交易平台**（V0.1.0）—— 通用量化回测引擎 + 黄金期货合约回测验证，覆盖行情导入/查询、元数据管理、结构化策略回测、报告分析与链路追踪，为后续扩展 ETF/股票/场外基金/国内期货/美股期货/黄金石油等商品期货的量化分析奠定通用模型基础。

## ✨ 核心特性

### 通用量化回测引擎（internal/backtest）

- **结构化策略定义模型**：universe（标的池）/ data（周期·撮合）/ indicators（MA/EMA/RSI/MACD/BOLL/ATR/STDDEV/HHV/LLV）/ signals（自研表达式引擎：cross_up/cross_down/ref/highest/lowest/abs + 比较/逻辑/算术）/ rules（数量模式 ALL_IN·ALL·FIXED·PERCENT·AMOUNT + 频率/次数/时间点）/ risk（止损止盈/仓位/分批建仓减仓/滑动窗口限制）/ cost（覆盖链）
- **无未来函数回放**：信号 bar 收盘确认、次根开盘成交（NEXT_BAR_OPEN）；ref 负偏移编译期拦截 + 表达式深度上限，杜绝前视
- **报告 8+2 项**：余额/收益率/收益额/持仓金额曲线、最大/平均投入、到期收益率、年化收益、最大回撤（含区间）、夏普、波动率、胜率、盈亏比、信号归因
- **⑨ 持仓链路追踪**：资金流水（连续可校验）+ 持仓变化（OPEN/ADD/REDUCE/CLOSE）+ 事件追踪（8 项登记：触发原因/时间/成交结果/委托下单/成交耗时/存活时间/未成交原因）
- **环境/模板模型**：BACKTEST/PAPER/SIMULATION/LIVE × 地区市场；交易时段/T+N/涨跌停/合约乘数/tick_size 引擎自适应；成本覆盖链 环境>任务>策略>账户
- **多用户隔离**：策略/账户/任务带 user_id，账户 group_id 子账户分组
- **任务管理**：异步执行（并发上限 4）、进度持久化、可取消、任务删除（归档留痕 + 审计日志）

### 行情与元数据

- **历史行情导入**：MVSV 文件解析（双列布局自动识别、20 个必填头部键、时区/字段校验）→ 分钟级 K 线落库（finv_quote_secu_kline_min）
- **历史行情查询**：ts 范围全量返回 + K 线图（黑金主题、三行图例显隐、均线/布林带、dataZoom 拖拽缩放、成交量/成交额万分位）
- **元数据字典**：交易所（finv_exchange）/ 市场（finv_market）/ 证券（finv_security）/ 币种（finv_currency）/ 地区（finv_region）/ 富途映射表（security/exchange/market_code/cs_market），种子数据开箱即用

### 工程化

- **单镜像 All-in-One**：ghcr.io/acanx/finvquant（Go embed 前端，单进程双端口 16001+16002），多平台 amd64/arm64
- **版本单一来源**：项目根 `VERSION` 文件（后端 -ldflags / 前端 package.json / 镜像 tag 三处共用）
- **一键部署**：deploy.cmd / upgrade.cmd（版本检测升级）/ rollback.cmd（回滚）+ Win11 部署文档
- **开发规范文档化**：ApiSpec（查询参数小驼峰、一接口一文档）/ MenuSpec（路由大驼峰、一菜单一文档）/ BacktestStrategySpec / UiSpec / LogSpec / ErrorCodeSpec 等 10 篇

## 🚀 快速开始（Docker）

### 方式一：Docker Compose（推荐）

```bash
# 1. 克隆代码（默认分支 dev）
git clone https://github.com/ACANX/FinvQuant.git
cd FinvQuant

# 2. 准备环境变量并启动（自动拉起 PostgreSQL 18 + Redis 8 + FinvQuant）
cp Deploy/.env.example Deploy/.env   # 按需修改（端口/密码/数据目录等）
docker compose --env-file Deploy/.env up -d
```

- 前端控制台：http://localhost:16002
- 服务端健康检查：http://localhost:16001/API/V1/Health/Live

### 方式二：All-in-One 镜像

```bash
docker pull ghcr.io/acanx/finvquant:latest
docker run -d --name finvquant -p 16001:16001 -p 16002:16002 \
  -e FINV_PG_HOST=host.docker.internal -e FINV_REDIS_ADDR=host.docker.internal:6379 \
  ghcr.io/acanx/finvquant:latest
```

> 镜像 tag 规则：分支推送打 `latest` + `v{VERSION}-YYYYMMDDHHMM`；推送 `v{VERSION}` git tag（正式发布）打 `latest` + `v{VERSION}`（全架构 amd64+arm64）。

## 💻 本地开发

```bash
# 服务端（Go 1.25.3，需本地 PostgreSQL 18 + Redis 8）
go run ./cmd/server

# 前端（端口 16002，/api 代理到 16001）
cd Web && npm install && npm run dev
```

## 🗂 业务菜单（Web 控制台）

| 菜单 | 路由 | 说明 |
|------|------|------|
| 仪表盘 | /dashboard | 系统状态总览 |
| 历史行情查询 / 导入 | /Meta/Finv/Quote/History/* | K 线查询、MVSV 导入 |
| 元数据管理（交易所/市场/证券） | /meta/exchange 等 | 字典维护 |
| 配置管理 / 环境管理 / 模板管理 | /Meta/Finv/Quant/{Config,Environment,Template} | 回测环境与模板 |
| 账户 / 资金 / 持仓管理 | /Meta/Finv/Quant/{Account,Fund,Position} | 回测账户与结果查看 |
| 黄金期货合约回测验证 | /Meta/Finv/Quant/Backtest/GoldFutures | 回测条件配置与启动 |
| 策略管理 | /Meta/Finv/Quant/Strategy | 结构化策略定义 |
| 回测分析 | /Meta/Finv/Quant/Backtest/Analysis | 报告指标卡 + 曲线 + ⑨链路追踪 + 导出 |
| 仿真/模拟盘/实盘（占位） | /Meta/Finv/Quant/Simulation/* 等 | 规划中，环境类型已建模 |

## 🔌 API 概览（基路径 /API/V1/Meta/Finv/Quant/）

- **Backtest/**：Strategy / Account / Environment / Template 四组 CRUD（List/Get/Save/Toggle/Delete）+ Run（Create/List/Get/Cancel/Report/Equity/Trades/Cashflows/PositionLogs/EventTraces/Delete + DeleteTask 系列），共 29 个端点
- **Quote/**：`Quote/Import/Upload`（导入）、`Quote/History/QuoteQuery`（查询）
- **Metadata/**：Exchange / Market / Security（List/Save/Toggle + Options/Lookup）
- **Health/**：`/API/V1/Health/Live`、`/API/V1/Health/Ready`、`/API/V1/Version`

> 完整接口文档见 [Docs/API/APIs.md](Docs/API/APIs.md)（#1~45 索引）与 [Docs/API/README.md](Docs/API/README.md)。

## 📁 目录结构

```
├── cmd/server/            # Go 服务端入口
├── internal/
│   ├── api/               # Gin 路由 + handler
│   ├── backtest/          # 通用量化回测引擎（模型/表达式/指标/引擎/服务）
│   ├── mvsv/              # MVSV 行情文件解析器
│   ├── meta/              # 元数据字典服务（交易所/市场/证券）
│   ├── quote/             # 历史行情导入与查询服务
│   ├── database/          # PostgreSQL 迁移
│   ├── config/ redisclient/ static/ webui/
├── Web/                   # Vue3 + Vite8 + Vuetify4 前端（views/Meta/Finv 分层）
├── Deploy/                # docker-compose + 部署/升级/回滚脚本 + 迁移 SQL
├── Docs/                  # API / Menu / DevSpec / DataFormat / DataDictMapping / Asset
├── .github/workflows/     # CI：多平台镜像构建 + GHCR 推送
├── VERSION                # 版本号单一来源（当前 0.1.0）
├── Prompt.md              # 结构化需求文档（持续更新）
└── VeritasQuant/          # 既有 Python 子项目（历史保留）
```

## 📚 文档索引

- [Prompt.md](Prompt.md) — 项目需求与技术基线
- [Docs/API/README.md](Docs/API/README.md) + [Docs/API/APIs.md](Docs/API/APIs.md) — API 接口文档（一接口一文档）
- [Docs/Menu/Menus.md](Docs/Menu/Menus.md) — 业务菜单文档（一菜单一文档）
- [Docs/DevSpec/](Docs/DevSpec/) — 开发规范（ApiSpec / MenuSpec / BacktestStrategySpec / UiSpec / LogSpec / ErrorCodeSpec / DocSpec / FileNamingSpec / FileEncodingSpec / GitSpec）
- [Docs/DataFormat/MvsvFileFormat.md](Docs/DataFormat/MvsvFileFormat.md) — MVSV 行情文件格式规范（含 NVDA/GCMain 示例文件）
- [Docs/DataDictMapping/](Docs/DataDictMapping/) — 数据字典与映射说明
- [Deploy/Win11DockerDeploy.md](Deploy/Win11DockerDeploy.md) — Windows 11 Docker 部署文档
- [Deploy/Win11DockerUpgrade.md](Deploy/Win11DockerUpgrade.md) — 增量升级文档
- [Deploy/DeployUpgradeGuide.md](Deploy/DeployUpgradeGuide.md) — 部署/升级/回滚脚本使用手册
- [VeritasQuant/README.md](VeritasQuant/README.md) — 既有 Python 子项目说明

## 📦 版本历史

- **v0.1.0**（2026-08-08）：通用量化回测引擎 + 黄金期货合约回测验证（PR #338/#339）、MVSV 双布局解析、元数据字典全量种子、多平台 CI 发布、一键部署脚本

## 许可

[MIT](LICENSE) © 2026 ACANX
