# Windows 11 本地部署教程（Docker 服务端 + 本地客户端）

> 对应 ISSUE #253：在 Windows 11 上通过 Docker（镜像容器）部署 VeritasQuant 服务端，
> 并在本地启动客户端连接服务端完成**模拟盘（PAPER）**与**券商仿真（SIMULATION）**实验。
>
> ⚠️ 实盘交易默认禁用；本教程不涉及实盘（LIVE）启用。

---

## 1. 环境要求

| 项目 | 要求 |
|------|------|
| 操作系统 | Windows 11（64 位，22H2 或更新；家庭版需支持 WSL2） |
| Docker Desktop | 4.x 或更新，启用 **WSL2 后端**（Settings → General → Use the WSL 2 based engine） |
| WSL2 | 内置 Linux 内核；`wsl --version` 可查（无则 `wsl --update`） |
| 内存 | 建议 ≥ 8 GB（Docker 引擎 + PostgreSQL + Redis + API 服务） |
| 磁盘 | 建议 ≥ 10 GB 可用空间（镜像 + 数据卷） |
| 网络 | 可拉取 Docker Hub 镜像（国内可配置镜像加速器） |
| Python | 仅客户端需要 ≥ 3.13（服务端在容器内运行，宿主无需 Python） |

验证 Docker 环境：

```powershell
docker --version
docker compose version
docker info
```

若 `docker info` 报错，先启动 Docker Desktop 并等待引擎就绪（托盘鲸鱼图标变绿）。

---

## 2. 获取代码

```powershell
git clone https://github.com/ACANX/VeritasQuant.git
cd VeritasQuant
```

（如已克隆，先拉取最新：`git pull`）

---

## 3. 部署物清单

| 文件 | 用途 |
|------|------|
| `Docker/Dockerfile` | 服务端应用镜像（多阶段构建，非 root 运行） |
| `Docker/docker-compose.deploy.yml` | 服务端编排：API + PostgreSQL + Redis（持久卷、健康检查） |
| `Docker/.env.deploy.example` | 环境变量模板（密码、端口、数据目录） |
| `scripts/DeployServer.py` | 部署脚本（check/build/start/status/logs/stop） |

---

## 4. 依赖说明

### 4.1 服务端（容器内）

| 组件 | 镜像/来源 | 说明 |
|------|-----------|------|
| API 服务 | `ghcr.io/acanx/veritasquant:<tag>`（GitHub Packages 构建，默认 `latest`） | FastAPI + Uvicorn，默认 `0.0.0.0:18000`；也可本地构建（`veritasquant/server:local`） |
| PostgreSQL | `postgres:16.4-alpine` | 事实/投影持久化（订单、成交、账户、审计） |
| Redis | `redis:7.4-alpine` | 跨进程事件分发（Redis Streams） |

> 服务端镜像由 GitHub Actions（CI `build-image` job）构建并发布到
> GitHub Container Registry（ghcr.io）。tag 推送（如 `V0.1.1`）会发布对应版本镜像与
> `latest`；push 到 main/dev 也会刷新 `latest`；手动触发（workflow_dispatch）时
> 会额外生成带版本前缀的时间戳镜像（如 `0.1.1-202608031217`，UTC）。
> Windows 部署直接拉取即可，无需本地构建。

### 4.2 客户端（Windows 本地）

客户端通过已安装的 Python 包运行（连接服务端 API）：

| 入口 | 用途 |
|------|------|
| `vq-api-server` | （通常只在容器内）API 服务 |
| `vq-run-backtest` | 回测 |
| `vq-run-paper-trading` | 模拟盘交易 |
| `vq-gui` / `vq-gui-client` | GUI 客户端 |
| `vq-import-market-data` / `vq-validate-market-data` | 数据导入与校验 |

---

## 5. 详细部署步骤

### 步骤 1：配置环境变量

```powershell
Copy-Item Docker\.env.deploy.example Docker\.env.deploy
# 用编辑器打开 Docker\.env.deploy，设置：
#   VQ_POSTGRES_PASSWORD=你的强密码（至少 12 位）
#   VQ_ENVIRONMENT=PAPER        （或 SIMULATION）
#   VQ_API_PORT=18000         （宿主映射端口，12000 以后避开常用端口）
```

### 步骤 2：环境自检（无需 Docker 也可运行）

```powershell
python scripts/DeployServer.py check
```

该命令校验编排文件语法（`docker compose config --quiet`）。

### 步骤 3：拉取服务端镜像（默认方式，无需本地构建）

服务端镜像由 GitHub Actions 构建并发布到 ghcr.io。默认编排直接拉取 `latest`：

```powershell
# 拉取 ghcr.io/acanx/veritasquant:latest
python scripts/DeployServer.py start   # 首次启动会自动拉取镜像
```

如需固定版本，在 `Docker/.env.deploy` 中设置（tag 必须与 GitHub Packages 页面实际发布的 tag 完全一致，见 6.6）：

```
VQ_IMAGE_TAG=latest
```

手动拉取验证：

```powershell
docker pull ghcr.io/acanx/veritasquant:latest
```

> **可选：本地构建**（不使用 GitHub 镜像时）——取消 `Docker/docker-compose.deploy.yml`
> 中 `server.build` 块注释，并把 `image` 改回 `veritasquant/server:local`，然后：
>
> ```powershell
> python scripts/DeployServer.py build
> ```

### 步骤 4：启动服务端

```powershell
python scripts/DeployServer.py start
```

- `--wait` 会等待三个容器健康检查全部通过；
- 查看状态：`python scripts/DeployServer.py status`；
- 查看日志：`python scripts/DeployServer.py logs --service server`。

### 步骤 5：验证服务端

```powershell
# liveness：进程存活
curl http://localhost:18000/health/live

# readiness：就绪门禁
curl http://localhost:18000/health/ready

# 版本
curl http://localhost:18000/api/v1/version
```

预期：`/health/live` 返回 `{"code":0,"data":{"status":"ALIVE",...}}`。

### 步骤 6：安装本地客户端

```powershell
# 使用项目 venv（建议）或全局环境
python -m venv .venv-client
.\.venv-client\Scripts\Activate.ps1
python -m pip install -e .
```

客户端入口即可用（`vq-run-backtest --help` 验证）。

### 步骤 7：执行模拟/仿真实验

**回测（离线，不依赖服务端）：**

```powershell
vq-run-backtest --config .\你的回测配置.yml
```

**模拟盘实验（连接服务端）：**

```powershell
vq-run-paper-trading --config .\你的模拟盘配置.yml --api http://localhost:18000
```

**GUI 客户端（连接服务端）：**

```powershell
vq-gui --api http://localhost:18000
```

> 说明：具体实验参数以各 console script 的 `--help` 与
> [Docs/VeritasQuantDevelopmentPlan.md](Docs/VeritasQuantDevelopmentPlan.md) 为准；
> 模拟盘/仿真配置中 `execution_mode` 使用 `PAPER` / `SIMULATION`。

### 步骤 8：停止服务端

```powershell
python scripts/DeployServer.py stop
```

- 停止并删除容器与网络；
- **数据卷保留**（PostgreSQL/Redis 数据不丢失）；
- 如需彻底清理数据：`docker compose -f Docker\docker-compose.deploy.yml down --volumes`。

---

## 6. 常见问题（FAQ）

### 6.1 Docker Desktop 启动失败 / WSL2 报错

- 确认 BIOS 开启虚拟化（任务管理器 → 性能 → CPU → 虚拟化：已启用）；
- `wsl --update` 更新内核；`wsl --status` 查看状态；
- Docker Desktop → Troubleshoot → Restart。

### 6.2 `VQ_POSTGRES_PASSWORD` 未设置

`docker compose` 会以 `VQ_POSTGRES_PASSWORD` 为必填变量，未设置时启动失败。
在 `Docker/.env.deploy` 中填入密码后重试。

### 6.3 端口 18000 被占用

- 修改 `Docker/.env.deploy` 中 `VQ_API_PORT=18001` 后重启；
- 或释放占用：`netstat -ano | findstr :18000` → `taskkill /PID <pid> /F`。

### 6.4 镜像拉取慢 / 超时

Docker Desktop → Settings → Docker Engine，配置镜像加速器后 `Apply & Restart`。

### 6.5 拉取 ghcr.io 镜像失败（denied / unauthorized）

- 确认 `Docker/.env.deploy` 中 `VQ_IMAGE_OWNER` 与镜像实际命名空间一致（默认 `acanx`）；
- 私有镜像需先登录：`echo $env:GITHUB_TOKEN | docker login ghcr.io -u <用户名> --password-stdin`（公开镜像无需登录）；
- 确认镜像已发布：在 GitHub 仓库 Packages 页面查看 `ghcr.io/acanx/veritasquant` 是否存在对应 tag。

### 6.6 拉取到不存在的版本 tag

- 检查 `VQ_IMAGE_TAG` 拼写：必须与 GitHub Packages 页面实际发布的 tag 完全一致（Git tag `V0.1.2` 触发构建时发布的镜像 tag 为 `0.1.2` 与 `latest`，镜像 tag 不带 `V` 前缀）；
- 查看已发布 tag：`docker manifest inspect ghcr.io/acanx/veritasquant:latest` 或 GitHub Packages 页面（`https://github.com/users/ACANX/packages/container/package/veritasquant`）。

### 6.7 容器启动后立即退出

```powershell
python scripts/DeployServer.py logs --service server
```

常见原因：端口冲突、`read_only: true` 下写入只读路径。确认 `.env.deploy` 中
`VQ_DATA_DIR` 指向宿主可写目录。

### 6.8 客户端连不上服务端

- 确认容器健康：`python scripts/DeployServer.py status`；
- 确认宿主端口：`curl http://localhost:18000/health/live`；
- 客户端配置中 API 地址使用 `http://localhost:18000`（勿用 `127.0.0.1` 时混用 IPv6）。

### 6.9 数据持久化

- PostgreSQL/Redis 使用命名卷（`vq-postgres` / `vq-redis`），`stop` 不删数据；
- 运行产物在 `vq-runtime` 卷与宿主 `VQ_DATA_DIR` 目录。

---

## 7. 安全与边界

- 本编排仅用于本地模拟盘/仿真；PostgreSQL 未对外暴露端口（仅容器网络内）；
- `.env.deploy` 含明文密码，已被 `.gitignore` 忽略，不要提交；
- 实盘（LIVE）默认禁用；任何实盘启用必须先走 Change 流程并满足
  [TechSpec 13 阶段 5 gate](Docs/VeritasQuantTechSpec.md)。
