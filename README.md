# FinvQuant

**量化策略交易平台** v0.1.0 — 前后端分离架构，支持通用量化回测全链路。

- **服务端**：Go 1.25.3 + Gin + PostgreSQL 18 + Redis 8（端口 **16001**）
- **前端**：Vue 3 + Vite 8 + Vuetify 4（端口 **16002**）
- **部署**：Docker Compose / 单容器 All-in-One 镜像（GHCR）
- **分支**：`dev`（开发分支，默认）

---

## 功能概览

| 模块 | 说明 | 状态 |
|------|------|------|
| 历史行情导入 | MVSV 分钟行情上传导入（字段级覆盖 upsert） | ✅ |
| 历史行情查询 | K 线蜡烛图查询（红涨绿跌，悬停详情） | ✅ |
| 元数据管理 | 交易所 / 市场 / 证券字典维护 | ✅ |
| 通用量化回测 | 策略 / 账户 / 任务 / 报告 / 资金持仓 / 链路追踪 | ✅ |
| 环境管理 | 回测 / 模拟盘 / 仿真 / 实盘环境配置（交易时段、规则、成本） | ✅ |
| 模板管理 | 策略 / 账户 / 环境模板（内置 + 自定义） | ✅ |
| 链路追踪⑨ | 资金流水明细 / 持仓变化明细 / 事件追踪（触发原因·成交结果·委托耗时） | ✅ |
| 多用户隔离 | 策略 / 账户 / 任务 / 环境 / 模板按 `user_id` 隔离 | ✅ |
| 单用户多账户 | 账户支持 `group_id` 分组（主/子账户） | ✅ |
| 实盘链路 | 行情、账户、策略、订单、风控 API 规划中 | 🚧 |

---

## 快速开始（Docker）

```bash
# 1. 克隆代码（默认分支 dev）
git clone https://github.com/ACANX/FinvQuant.git
cd FinvQuant

# 2. 准备环境变量并启动
cp Deploy/.env.example Deploy/.env   # 按需修改
docker compose --env-file Deploy/.env up -d
```

- 前端控制台：http://localhost:16002
- 服务端 API：http://localhost:16001/API/V1/Health/Live

或直接使用 All-in-One 镜像（单容器 = 服务端 + 前端，无需独立 Nginx）：

```bash
docker pull ghcr.io/acanx/finvquant:latest
docker run -d --name finvquant -p 16001:16001 -p 16002:16002 \
  -e FINV_PG_HOST=host.docker.internal -e FINV_REDIS_ADDR=host.docker.internal:6379 \
  ghcr.io/acanx/finvquant:latest
```

---

## 本地开发

```bash
# 服务端（Go 1.25.3）
go run ./cmd/server

# 前端（端口 16002，/api 代理到 16001）
cd Web && npm install && npm run dev
```

---

## 目录结构

```
├── cmd/server/              # Go 服务端入口（端口 16001）
├── internal/                # Go 服务端内部模块
│   ├── api/                 #   Gin 路由与 HTTP 处理器
│   │   ├── router.go        #   路由注册（34+ 端点）
│   │   └── handler/         #   各领域处理器
│   ├── backtest/            #   量化回测模块（引擎/服务/指标/表达式/模型）
│   ├── config/              #   环境变量配置加载（FINV_* 前缀）
│   ├── database/            #   PostgreSQL 连接池（pgx/v5）+ 自动迁移
│   ├── meta/                #   元数据管理服务（交易所/市场/证券）
│   ├── mvsv/                #   MVSV 行情格式解析器
│   ├── quote/               #   历史行情导入/查询服务
│   ├── redisclient/         #   Redis 8 客户端
│   ├── static/              #   静态资源服务
│   └── webui/               #   go:embed 内嵌前端构建产物
├── Web/                     # 前端（Vue 3 + Vite 8 + Vuetify 4）
│   ├── src/                 #   源码
│   │   ├── views/           #     页面组件
│   │   ├── App.vue          #     根组件
│   │   ├── router.ts        #     路由定义
│   │   └── api.ts           #     API 客户端
│   └── vite.config.ts       #   Vite 配置（/api 代理）
├── Deploy/                  # 部署编排
│   ├── docker-compose.yml   #   Compose 一键部署（含 PG18 + Redis 8）
│   ├── .env.example         #   环境变量模板
│   ├── Migrations/          #   数据库迁移脚本（50+ 个迁移）
│   ├── upgrade.cmd          #   增量升级脚本
│   ├── rollback.cmd         #   回滚脚本
│   ├── Win11DockerDeploy.md #   Windows 11 部署指南
│   └── Win11DockerUpgrade.md#   Windows 11 增量升级指南
├── Docs/                    # 文档
│   ├── API/                 #   服务端 API 接口文档（34 个端点）
│   ├── Asset/Backtest/      #   回测架构设计图集（SVG）
│   ├── DataDictMapping/     #   数据字典映射文档
│   ├── DataFormat/          #   MVSV 行情格式说明
│   ├── DevSpec/             #   开发规范（API/策略/错误码/文件命名等）
│   ├── GitHubActionUpgrade.md # CI/CD 升级指南
│   └── Menu/                #   前端菜单文档
├── Dockerfile               # 多阶段构建镜像（前端 + 后端）
├── .github/workflows/CI.yml # CI/CD：构建 + 测试 + 推送 GHCR 镜像
├── go.mod / go.sum          # Go 模块定义（Go 1.25.3）
├── VERSION                  # 版本文件（当前 0.1.0）
├── Prompt.md                # 结构化需求文档（持续更新）
├── VeritasQuant/            # 既有子项目（Python 量化平台，历史保留）
└── README.md                # 本文件
```

---

## 服务端 API

基路径：`/API/V1`（统一大写），端口 **16001**

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/API/V1/Health/Live` | 存活检查 |
| GET | `/API/V1/Health/Ready` | 就绪检查 |
| GET | `/API/V1/Version` | 版本信息 |

### 历史行情

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/API/V1/Meta/Finv/Quant/Quote/Import/Upload` | MVSV 分钟行情上传导入 |
| GET | `/API/V1/Meta/Finv/Quant/Quote/History/QuoteQuery` | 历史行情查询（K 线） |

### 元数据管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/API/V1/Meta/Finv/Quant/Metadata/Exchange/*` | 交易所信息维护 |
| GET/POST | `/API/V1/Meta/Finv/Quant/Metadata/Market/*` | 市场信息维护 |
| GET/POST | `/API/V1/Meta/Finv/Quant/Metadata/Security/*` | 证券信息维护（含 Options/Lookup） |

### 量化回测

路径前缀：`/API/V1/Meta/Finv/Quant/Backtest/`

**策略管理**：`Strategy/List`, `Get`, `Save`, `Toggle`, `Delete`
**账户管理**：`Account/List`, `Get`, `Save`, `Toggle`, `Delete`
**任务管理**：`Run/Create`, `List`, `Get`, `Cancel`, `Delete`
**报告分析**：`Run/Report`, `Equity`, `Trades`
**链路追踪**：`Run/Cashflows`, `PositionLogs`, `EventTraces`
**删除任务**：`Run/DeleteTask/List`, `DeleteTask/Logs`, `DeleteTask/Retry`, `DeleteTask/Archives`
**环境管理**：`Environment/List`, `Get`, `Save`, `Toggle`, `Delete`
**模板管理**：`Template/List`, `Get`, `Save`, `Delete`

> 详细接口文档见：`Docs/API/README.md`

---

## 量化回测模块

### 能力矩阵

| 功能 | 说明 |
|------|------|
| 策略定义 | JSON 模型 v1（universe / data / indicators / signals / rules / risk / cost），保存时编译校验 |
| 指标计算 | MA / EMA / RSI / MACD / BOLL / ATR / STDDEV / HHV / LLV |
| 信号表达式 | 自研引擎（比较/逻辑/算术 + cross_up/down / ref / highest/lowest / abs），深度上限 64 |
| 回测引擎 | 逐 bar 回放（预热 → 挂单撮合 NEXT_BAR_OPEN → 止损止盈 → 信号 → 规则限制 → 账户更新 → 报告点） |
| 报告生成 | 余额/收益率/收益额/持仓金额曲线 + 最大回撤/夏普/胜率/盈亏比等技术指标 |
| 链路追踪⑨ | 资金流水明细 / 持仓变化明细 / 事件追踪（触发原因·成交结果·委托耗时·未成交原因分类） |
| 环境自适应 | 交易时段过滤（含跨午夜）、tick_size 对齐、T+N/涨跌停/合约乘数、撮合模式、币种校验、成本覆盖链（环境 > 任务 > 策略 > 账户） |
| 多用户隔离 | 策略/账户/任务/环境/模板按 `user_id` 隔离，所有 List/Get/Toggle/Delete/CreateRun 均做归属校验 |
| 异步调度 | 并发上限 4、进度/状态持久化、支持取消、重启悬挂自动标记 FAILED |
| 内置模板 | 双均线 / RSI / 布林带 / MACD 策略模板 + 默认环境模板（GCMain 黄金期货 / 沪深 ETF） |

### 数据库迁移

- 迁移文件：`Deploy/Migrations/`（50+ 个迁移，启动时自动应用）
- 分段约定：`V1~V99999` 为表结构/变更脚本（DDL），`V100000+` 为数据种子脚本（DML）
- 表名规范：业务表统一 `finv_` 前缀，行情表 `finv_quote_xxx`，回测表 `finv_quant_xxx`

---

## CI/CD

- 触发器：push（dev/main/FinvQuant）、tag `v*`、PR、手动
- 流程：
  1. **build-server**：Go vet + build + test（Go 1.25.3）
  2. **build-web**：npm ci + build（Node 24）
  3. **docker**（非 PR）：构建并推送 All-in-One 镜像 `ghcr.io/acanx/finvquant`（latest + 版本 tag）
  4. **docker-pr**（PR）：仅构建不推送，验证镜像可构建

---

## 端口约定

| 服务 | 端口 |
|------|------|
| Go 服务端 | **16001** |
| 前端 Web | **16002** |
| PostgreSQL（宿主映射） | 5433 |
| Redis（宿主映射） | 6380 |

---

## 文档

- [Prompt.md](Prompt.md) — 项目需求与技术基线（结构化，持续更新）
- [Docs/API/README.md](Docs/API/README.md) — 服务端 API 接口文档
- [Docs/DevSpec/BacktestStrategySpec.md](Docs/DevSpec/BacktestStrategySpec.md) — 策略定义模型与表达式语法
- [Docs/Asset/Backtest/README.md](Docs/Asset/Backtest/README.md) — 回测架构设计图集（SVG）
- [Deploy/Win11DockerDeploy.md](Deploy/Win11DockerDeploy.md) — Windows 11 Docker 部署文档
- [Deploy/Win11DockerUpgrade.md](Deploy/Win11DockerUpgrade.md) — Windows 11 Docker 增量升级文档
- [VeritasQuant/README.md](VeritasQuant/README.md) — 既有 Python 子项目说明

---

## 许可

[MIT](LICENSE) © 2026 ACANX