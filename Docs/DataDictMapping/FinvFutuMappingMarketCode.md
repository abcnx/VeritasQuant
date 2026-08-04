# FinvFutuMappingMarketCode — 富途市场代码映射

> 所属：FinvQuant 数据字典映射 · 存放：`Docs/DataDictMapping/FinvFutuMappingMarketCode.md`
> 数据表：`finv_futu_mapping_market_code`（表结构：[`Deploy/Migrations/V8__finv_futu_mapping_market_code.sql`](../../Deploy/Migrations/V8__finv_futu_mapping_market_code.sql)）
> 用途：FT（富途/moomoo）行情源市场代码与交易市场代码（finv_market.market_code）的字段映射表。

## 1. 表结构

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `futu_market_code` | INTEGER | PK | 富途行情源市场代码（如 `1` / `10` / `11` / `30` / `70` / `120` / `360`） |
| `finv_market_code` | INTEGER | NOT NULL，1~999999 | 交易市场代码（关联 [FinvMarket](FinvMarket.md) `market_code`） |
| `gmt_create` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 首次插入时间 |
| `gmt_update` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 最后更新时间（触发器维护） |

索引：

- `idx_finv_futu_mapping_market_code_finv`：`(finv_market_code, futu_market_code)`（按交易市场代码反查富途市场代码）

## 2. 数据清单

> 本表**暂无初始数据**，待确认富途市场代码与 finv_market_code 的完整映射后补充；届时新增数据种子脚本（`Deploy/Migrations/V100000+` 段位），并同步更新本节表格。

| futu_market_code | finv_market_code |
|-----------------:|-----------------:|
| （待补充） | |

## 3. 说明

- **映射方向**：`futu_market_code`（富途行情源市场代码）→ `finv_market_code`（交易市场代码）；富途侧代码值域由行情源决定，主键不设 CHECK 限制。
- **关联字典**：`finv_market_code` → [FinvMarket.md](FinvMarket.md)（交易市场）；关联不建物理外键，由程序层控制（项目惯例）。
- **审计字段**：`gmt_create` / `gmt_update` 与既有表规范一致，`gmt_update` 由 `vq_set_gmt_update()` 触发器自动维护。
