# 历史行情查询（QuoteQuery）— 菜单业务文档

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/HistoryQuote/HistoryQuoteQuery.md`
> 对应视图：`Web/src/views/QuoteQueryView.vue`
> 接口契约：见 [Docs/API/HistoryQuote/HistoryQuote.md](../../API/HistoryQuote/HistoryQuote.md)

## 1. 菜单入口

- **菜单位置**：左侧导航栏（`Web/src/App.vue` 中 `menuItems` 定义）
- **菜单名称**：历史行情查询
- **菜单 key**：`quote-query`
- **菜单图标**：`mdi-chart-candlestick`
- **对应视图组件**：`Web/src/views/QuoteQueryView.vue`

## 2. 业务功能概述

历史行情查询菜单为用户提供**按证券代码 + 交易日查询分钟级 K 线**的能力，核心功能点：

| 功能 | 说明 |
|------|------|
| 分钟级 K 线查询 | 按证券代码 + 交易日查询 1 分钟 K 线（周期目前仅支持 `Min`） |
| 多日回溯 | 支持 **1 日 / 5 日**切换，5 日模式展示最近 5 个有数据交易日的分钟级 K 线 |
| 日历点选日期 | 交易日期通过日历组件点选，不可晚于当天；自动转为 `yyyymmdd` 提交 |
| 分页翻页 | 每页固定 240 根（约全天 4 小时交易时段的分钟线数），底部页码翻页 |
| 富途风格 K 线 | 红涨绿跌配色 + MA5/MA10/MA20/MA30 均线 + 成交量副图，悬停查看明细 |
| 图表缩放 | 支持鼠标滚轮 / 拖拽缩放（dataZoom），价格轴与成交量轴联动 |

## 3. 操作流程（用户视角）

```
输入证券代码 → 日历点选交易日期 → 选择周期（Min）→ 切换 1日/5日 → 点击「查询」
    → 查看 K 线图（悬停看明细、滚轮缩放）→ 底部页码翻页查看其余分钟线
```

详细步骤：

1. **输入证券代码**：支持股票 / ETF 等代码（如 `NVDA`、`518880`）。
2. **点选交易日期**：点击日期输入框弹出日历，选择交易日（最大为当天）。多日模式（5 日）时该日期作为**回溯截止日**。
3. **选择周期**：目前仅提供「1分（Min）」选项。
4. **切换 1 日 / 5 日**：按钮组切换回溯交易日数。
5. **点击「查询」**：校验通过后请求后端接口并渲染图表。
6. **翻页查看**：底部显示「每页 240 根 · 共 N 根」，通过分页组件翻页（翻页自动重新查询）。

> 提示：切换查询条件（代码 / 日期 / 天数 / 周期）后会自动回到第 1 页，但**不会自动触发查询**，需手动点击「查询」。

## 4. 前端处理逻辑

### 4.1 查询参数校验

| 校验项 | 规则 | 失败提示 |
|--------|------|----------|
| 证券代码 | 非空（去除首尾空格） | `证券代码不能为空` |
| 交易日期 | 必须通过日历选择，格式 `yyyymmdd`（8 位数字） | `请通过日历选择交易日期` |

### 4.2 请求与响应处理

- **请求**：`GET /API/V1/Quote/Query`，Query 参数如下：

| 参数 | 取值 | 说明 |
|------|------|------|
| `secu_code` | 用户输入 | 证券代码 |
| `date` | `yyyymmdd` | 交易日期（5 日模式为回溯截止日） |
| `period` | `Min` | 周期 |
| `days` | `1` / `5` | 回溯交易日数 |
| `page` | 当前页码 | 默认 1 |
| `page_size` | `240` | 每页条数（常量 `PAGE_SIZE`） |

- **响应处理**：
  - `body.code !== 0` → 展示 `body.message` 为错误提示。
  - 请求异常（网络失败）→ 提示 `网络错误：无法连接服务端`。
  - 成功 → 保存 `bars` 与 `total`，生成摘要信息，渲染图表；无数据时清空图表。

- **摘要信息格式**：
  `{证券代码} {日期范围} {周期} · {N}日 · 共 {total} 根 · 第 {page}/{totalPages} 页`
  - 单日模式日期范围直接显示交易日期；5 日模式显示首尾 K 线的 `MMDD-MMDD` 区间。

### 4.3 K 线渲染逻辑（ECharts）

- **配色**：红涨绿跌（富途牛牛风格）——上涨 `#ef232a`，下跌 `#14b143`。
- **K 线序列**：每根蜡烛数据 `[open, close, low, high]`；缺失价格视为 0（不参与展示）。
- **均线**：MA5 / MA10 / MA20 / MA30 四条，颜色 `#f6c343` / `#3b8ff7` / `#c56cf0` / `#2fb28a`。
  - 采用**滑动窗口**计算；窗口内存在缺失值时该点输出 `null`，避免断点被错误填充（`connectNulls: false`）。
- **成交量副图**：底部独立网格，柱色按当根 K 线涨跌着色（红涨绿跌），宽度 60%。
- **双轴联动**：价格轴（右）与成交量轴（右）通过 `axisPointer.link` 联动十字光标。
- **坐标轴标签**：
  - 1 日模式：仅显示 `HH:MM`；
  - 5 日模式：显示 `MM/DD HH:MM`，便于区分不同交易日分段。
- **缩放**：内置 `dataZoom`（`type: inside`），主图与成交量图同步缩放。

### 4.4 悬停详情（Tooltip）

悬停任意 K 线显示：

- 交易日期 + 时间（`yyyymmdd hh:mm:ss`）
- 开盘 / 收盘（红绿着色）/ 最高 / 最低
- 涨跌额与涨跌幅（红涨绿跌着色；仅当后端返回 `change` / `change_pct` 时展示，否则显示 `—`）
- 成交量 / 成交额（有值才展示，千分位格式化）
- 当前点 MA5 / MA10 / MA20 / MA30 数值
- 导入备注（`remark` 有值时展示）

## 5. 后端处理逻辑

> 完整接口契约见 [Docs/API/HistoryQuote/HistoryQuote.md](../../API/HistoryQuote/HistoryQuote.md)，本节为与菜单强相关的处理链路。

```
GET /API/V1/Quote/Query
  → handler.QuoteQuery.Query（参数解析与校验）
  → service.QueryBars（多日回溯 + 分页查询 + 涨跌计算）
  → 返回 { secu_code, date, period, days, page, page_size, total, count, bars }
```

### 5.1 参数解析与校验（`internal/api/handler/query.go`）

| 参数 | 默认值 | 校验规则 |
|------|--------|----------|
| `secu_code` | — | 必填，为空返回 4001 |
| `date` | — | 必须为合法正整数 `yyyymmdd`，否则返回 4001 |
| `period` | `Min` | 空则默认 `Min`；非 `Min` 由服务层拒绝 |
| `days` | `1` | 非法 / 小于 1 时回退默认值 |
| `page` | `1` | 同上 |
| `page_size` | `240` | 同上 |

### 5.2 查询服务（`internal/quote/service.go`）

- **周期限制**：目前仅支持 `Min`（1 分钟），其他周期返回错误。
- **参数截断**：`days` 上限 10（前端仅提供 1 / 5）；`page_size` 上限 1000。
- **多日回溯**：取 `date` 当天及之前最近 N 个**有数据**的交易日（按 `date` 倒序去重取前 N），保证 5 日模式跳过无数据日。
- **分页查询**：`total` 统计满足条件的总分钟线条数；当前页按 `date ASC, ts ASC` 排序，`LIMIT/OFFSET` 取数，前端据此计算总页数。
- **涨跌计算**：`change = close - prev_close`，`change_pct = change / prev_close * 100`；任一价格缺失或 `prev_close = 0` 时输出 `null`。
- **数据表**：`finv_quote_secu_kline_min`（分钟级 K 线，主键 `ts + market_code + secu_code`）。

## 6. 常见问题与提示

| 现象 | 原因 / 处理 |
|------|-------------|
| 提示「证券代码不能为空」 | 未输入代码，补全后查询 |
| 提示「请通过日历选择交易日期」 | 日期为空或格式非法，需通过日历点选 |
| 提示「查询失败: ...」 | 后端返回业务错误（如周期不支持、无数据源错误），按 `message` 定位 |
| 提示「网络错误：无法连接服务端」 | 服务端不可达或接口异常，检查服务端是否启动 |
| 图表为空但提示查询完成 | 该证券代码在所选日期范围无分钟级数据，可尝试切换 5 日或更换日期 |
| 5 日模式日期区间不足 5 日 | 回溯按「有数据的交易日」计算，历史数据不足时以实际存在数据为准 |

## 7. 关联文档与代码

| 类型 | 位置 |
|------|------|
| 前端视图 | `Web/src/views/QuoteQueryView.vue` |
| 菜单定义 | `Web/src/App.vue` |
| 使用的后端接口 | `GET /API/V1/Quote/Query`（[API 文档 `Docs/API/HistoryQuote/HistoryQuote.md`](../../API/HistoryQuote/HistoryQuote.md)） |
| 后端处理器 | `internal/api/handler/query.go` |
| 查询服务 | `internal/quote/service.go` |
| 接口规范 | `Docs/DevSpec/ApiSpec.md` |
| 文档规范 | `Docs/DevSpec/DocSpec.md` |
