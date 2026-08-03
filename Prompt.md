# FinvQuant 项目需求文档（Prompt）

> 本文档是 FinvQuant 仓库（量化策略交易平台）的结构化需求说明，**支持后续持续更新与补充完善**。
> 更新方式：直接编辑本文件，保持分节结构，在对应章节追加或修订内容；重大变更请在文末"变更记录"登记。

## 1. 项目定位

FinvQuant 是一个**量化策略交易平台**，采用**前后端分离**架构：
- 服务端：Go 语言（Gin），提供量化策略与交易相关 API；
- 前端：Web 控制台（Vue3），提供策略管理、行情与交易界面。

## 2. 仓库结构

```
.
├── cmd/server/            # Go 服务端入口（端口 16001）
├── internal/              # Go 服务端内部模块
│   ├── api/               #   Gin 路由与处理器
│   ├── config/            #   配置加载（环境变量）
│   ├── database/          #   PostgreSQL 18 连接（pgx/v5）
│   └── redisclient/       #   Redis 8 连接（go-redis/v9）
├── Web/                   # 前端（Vue3 + Vite8 + Vuetify4，端口 16002）
├── deploy/                # Docker Compose 部署编排
├── .github/workflows/     # GitHub Actions：构建 + 推送 GHCR 镜像
├── Dockerfile             # 服务端镜像（多阶段构建）
├── VeritasQuant/          # 既有子项目（Python 量化平台，历史保留）
├── go.mod / go.sum        # Go 模块定义
├── .gitignore             # Go 版本忽略规则
└── Prompt.md              # 本需求文档
```

## 3. 技术选型（当前基线）

| 层 | 技术 | 版本 | 说明 |
|----|------|------|------|
| 服务端语言 | Go | **1.25.3** | 平台服务端唯一语言 |
| Web 框架 | Gin | **最新**（v1.12.x） | HTTP API 框架 |
| 数据库 | PostgreSQL | **18** | 业务主库（pgx/v5 驱动） |
| 缓存/消息 | Redis | **8** | go-redis/v9 客户端 |
| 前端框架 | Vue | 3.x | Composition API |
| 构建工具 | Vite | **8.x** | 前端构建/开发服务器 |
| UI 组件库 | Vuetify | **4.x** | Material Design 组件 |
| 包管理器 | npm | 最新 | 前端依赖管理 |
| Go 模块代理 | goproxy.cn | — | 国内构建加速（可选） |

> 版本策略：依赖采用"最新稳定版"；升级时同步更新本文档与 `go.mod`/`package.json`。

## 4. 服务端（Go）设计

### 4.1 端口与启动
- 默认监听端口：**16001**
- 配置方式：环境变量（`FINV_*` 前缀），见 `internal/config/config.go`
- 健康检查：`GET /api/v1/health/live`、`GET /api/v1/health/ready`
- 版本信息：`GET /api/v1/version`

### 4.2 模块划分（internal）
- `api`：Gin 路由注册、HTTP 处理器
- `config`：环境变量配置加载
- `database`：PostgreSQL 18 连接池（pgxpool + Ping 探活）
- `redisclient`：Redis 8 客户端（Ping 探活）

### 4.3 依赖基线
- `github.com/gin-gonic/gin` v1.12.x
- `github.com/redis/go-redis/v9` v9.22.x
- `github.com/jackc/pgx/v5` v5.x

## 5. 前端（Web）设计

- 默认端口：**16002**
- 技术栈：Vue 3.5 + Vite 8.2 + Vuetify 4.1（TypeScript）
- 目录：`Web/`（vite 标准结构）
- 开发代理：`/api` → `http://localhost:16001`（见 `Web/vite.config.ts`）
- 生产部署：Nginx 静态托管 + `/api` 反向代理到服务端 16001（见 `Web/nginx.conf`）

## 6. 部署与镜像

### 6.1 镜像（GitHub Packages / GHCR）
| 镜像 | 内容 | 端口 |
|------|------|------|
| `ghcr.io/acanx/finvquant-server` | Go 服务端 | 16001 |
| `ghcr.io/acanx/finvquant-web` | 前端（Nginx） | 16002 |

### 6.2 本地 Docker 部署
```bash
cp deploy/.env.example deploy/.env   # 按需修改
docker compose --env-file deploy/.env up -d
# 或直接使用已发布镜像：
docker pull ghcr.io/acanx/finvquant-server:latest
docker pull ghcr.io/acanx/finvquant-web:latest
```

### 6.3 依赖服务（Compose 内置）
- `postgres`：`postgres:18-alpine`（宿主映射 5433）
- `redis`：`redis:8-alpine`（宿主映射 6380）

## 7. CI/CD（GitHub Actions）

- 文件：`.github/workflows/build-publish.yml`
- 触发：push（dev/main/FinvQuant）、tag `v*`、PR、手动
- 流程：
  1. `build-server`：Go vet + build + test（Go 1.25.3）
  2. `build-web`：npm ci + build（Node 24）
  3. `docker`（非 PR）：构建并推送 server/web 镜像到 GHCR（latest + 版本 tag）
  4. `docker-pr`（PR）：仅构建不推送，验证镜像可构建

## 8. 端口约定（汇总）

| 服务 | 端口 |
|------|------|
| Go 服务端 | **16001** |
| 前端 Web | **16002** |
| PostgreSQL（宿主映射） | 5433 |
| Redis（宿主映射） | 6380 |

## 9. 待办 / 规划（Roadmap）

- [ ] 服务端业务模块：行情、账户、策略、订单、风控 API
- [ ] 前端业务页面：策略管理、行情看板、交易面板
- [ ] 数据库迁移与 Schema 管理（golang-migrate / goose）
- [ ] Redis 缓存策略与消息通道接入
- [ ] 认证授权（JWT / RBAC）
- [ ] 与 `VeritasQuant/` 子项目的数据/能力集成（行情导入 PG 等）

## 10. 变更记录

| 日期 | 变更 | 说明 |
|------|------|------|
| 2026-08-04 | 初始版本 | 初始化 Go 服务端（Gin 最新 / PG18 / Redis8 / Go 1.25.3）+ Vue3+Vite8+Vuetify4 前端；端口 16001/16002；GHCR 镜像构建与 Docker Compose 部署；.gitignore 改 Go 版 |
