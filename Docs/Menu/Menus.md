# Web 前端业务菜单总览（Menus）

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/Menus.md`
> 适用范围：`Web/` 前端侧边导航（`Web/src/App.vue` 中 `menuItems` 定义）中的全部业务菜单索引。
> 菜单文档规范见 [Docs/DevSpec/MenuSpec.md](../DevSpec/MenuSpec.md)；接口文档见 [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md)。

## 菜单清单

| # | 菜单名称 | key | 图标 | 对应视图组件 | 使用的后端接口 | 菜单文档 |
|---|----------|-----|------|--------------|----------------|----------|
| 1 | 仪表盘 | `dashboard` | `mdi-view-dashboard` | `Web/src/views/DashboardView.vue` | `GET /API/V1/health/live`（存活探针） | — |
| 2 | 历史行情查询 | `quote-query` | `mdi-chart-candlestick` | `Web/src/views/QuoteQueryView.vue` | `GET /API/V1/Quote/Query` | [HistoryQuote/HistoryQuoteQuery.md](HistoryQuote/HistoryQuoteQuery.md) |
| 3 | 历史行情数据导入 | `quote-import` | `mdi-database-import` | `Web/src/views/QuoteImportView.vue` | `POST /API/V1/Quote/Import/Upload` | — |

## 说明

- 菜单定义与视图组件一一对应，全部在 `Web/src/App.vue` 的 `menuItems` 中登记；新增菜单时须同步更新本索引。
- 「菜单文档」列指向对应业务菜单介绍文档（`Docs/Menu/xxx/XXXX.md`，见 [MenuSpec.md](../DevSpec/MenuSpec.md)）；暂无文档的菜单以 `—` 标注，待补充。
- 「使用的后端接口」列与对应 API 接口文档的「已使用位置登记」互为索引（见 [ApiSpec.md](../DevSpec/ApiSpec.md) 第 6 节）。
- 视图组件名称与菜单 key 的对应关系：

| 菜单 key | 视图组件 |
|----------|----------|
| `dashboard` | `DashboardView.vue` |
| `quote-query` | `QuoteQueryView.vue` |
| `quote-import` | `QuoteImportView.vue` |

## 相关文档

- [Docs/DevSpec/MenuSpec.md](../DevSpec/MenuSpec.md) — 前端菜单开发规范
- [Docs/API/README.md](../API/README.md) — 服务端 API 端点索引
