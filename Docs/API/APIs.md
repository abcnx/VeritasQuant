# 服务端 API 总览（APIs）

> 所属：FinvQuant 服务端 · 存放：`Docs/API/APIs.md`
> 适用范围：服务端（Go + Gin）`/API/V1` 路由组下的全部接口清单索引。
> 基路径：`/API/V1`（统一大写，见 [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md)）；服务端端口 **16001**。
> 接口文档规范见 [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md)；菜单使用方索引见 [Docs/Menu/Menus.md](../Menu/Menus.md)。

## API 清单

| # | 方法 | 路径 | 说明 | 详细文档 | 已使用位置（业务菜单） |
|---|------|------|------|----------|------------------------|
| 1 | GET | `/API/V1/health/live` | 存活探针：进程存活即 200，返回 `status` / `server` | — | 仪表盘（`DashboardView.vue`） |
| 2 | GET | `/API/V1/health/ready` | 就绪探针：PG 与 Redis 均可达才 200，否则 503 | — | — |
| 3 | GET | `/API/V1/version` | 服务端版本信息：`name` / `version` / `go_version` / `commit` | — | — |
| 4 | POST | `/API/V1/Quote/Import/Upload` | 上传 MVSV 历史行情文件并导入 PG（字段级/整行覆盖） | [ImportsUpload.md](ImportsUpload.md) | 历史行情数据导入（`QuoteImportView.vue`） |
| 5 | GET | `/API/V1/Quote/Query` | 按证券代码+日期查询分钟级 K 线（周期 Min，1日/5日回溯，分页） | [HistoryQuote/HistoryQuote.md](HistoryQuote/HistoryQuote.md) | 历史行情查询（[Menus.md](../Menu/Menus.md#2) 之历史行情查询菜单，文档见 [HistoryQuote/HistoryQuoteQuery.md](../Menu/HistoryQuote/HistoryQuoteQuery.md)） |

## 说明

- 接口清单以 `internal/api/router.go` 中 `/API/V1` 路由组注册为准；新增接口须同步更新本索引及 [Docs/API/README.md](README.md) 端点索引。
- 「详细文档」列指向对应接口文档（`Docs/API/xxx/XXXX.md`，见 [ApiSpec.md](../DevSpec/ApiSpec.md) 第 6 节）；暂无独立文档的接口以 `—` 标注。
- 「已使用位置（业务菜单）」列与菜单文档中的接口引用互为索引（见 [MenuSpec.md](../DevSpec/MenuSpec.md) 第 3 节）；接口新增使用方或下线时须同步更新。
- 通用探针/信息接口（health、version）不强制要求独立文档，但契约变更须同步本索引。

## 相关文档

- [Docs/API/README.md](README.md) — API 文档入口（响应信封、错误码、端点索引）
- [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md) — API 开发规范
- [Docs/Menu/Menus.md](../Menu/Menus.md) — Web 前端业务菜单总览
