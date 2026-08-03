# FinvQuant

量化策略交易平台，**前后端分离**架构：

- **服务端**：Go 1.25.3 + Gin + PostgreSQL 18 + Redis 8（端口 **16001**）
- **前端**：Vue3 + Vite8 + Vuetify4（端口 **16002**）
- **部署**：Docker Compose / GitHub Packages (GHCR) 镜像

## 快速开始（Docker）

```bash
cp deploy/.env.example deploy/.env   # 按需修改
docker compose --env-file deploy/.env up -d
```

- 前端控制台：http://localhost:16002
- 服务端 API：http://localhost:16001/api/v1/health/live

或直接使用已发布镜像：

```bash
docker pull ghcr.io/acanx/finvquant-server:latest
docker pull ghcr.io/acanx/finvquant-web:latest
```

## 本地开发

```bash
# 服务端（Go 1.25.3）
go run ./cmd/server

# 前端（端口 16002，/api 代理到 16001）
cd Web && npm install && npm run dev
```

## 目录结构

```
├── cmd/  internal/       # Go 服务端
├── Web/                  # Vue3+Vite8+Vuetify4 前端
├── deploy/               # Docker Compose 编排
├── .github/workflows/    # CI/CD（构建 + 推送 GHCR）
├── Prompt.md             # 结构化需求文档（持续更新）
└── VeritasQuant/         # 既有子项目（Python 量化平台，历史保留）
```

## 文档

- [Prompt.md](Prompt.md) — 项目需求与技术基线（结构化，持续更新）
- [VeritasQuant/README.md](VeritasQuant/README.md) — 既有 Python 子项目说明

## 许可

[MIT](LICENSE) © 2026 ACANX
