# FinvFutuMappingCSMarket — 富途 CS 市场映射

> 所属：FinvQuant 数据字典映射 · 存放：`Docs/DataDictMapping/FinvFutuMappingCSMarket.md`
> 数据表：`finv_futu_mapping_cs_market`（表结构：[`Deploy/Migrations/V9__finv_futu_mapping_cs_market.sql`](../../Deploy/Migrations/V9__finv_futu_mapping_cs_market.sql)）
> 用途：FT（富途/moomoo）行情源 CS 市场标识与交易所代码（finv_exchange.exchange_code）的字段映射表。

## 1. 表结构

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `futu_cs_market` | TEXT | PK | 富途行情源 CS 市场标识（如 `HK` / `US` / `CN` 等，具体取值以行情源为准） |
| `finv_exchange_code` | INTEGER | NOT NULL，1~999999 | 交易所代码（关联 [FinvExchange](FinvExchange.md) `exchange_code`） |
| `gmt_create` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 首次插入时间 |
| `gmt_update` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 最后更新时间（触发器维护） |

索引：

- `idx_finv_futu_mapping_cs_market_finv`：`(finv_exchange_code, futu_cs_market)`（按交易所代码反查富途 CS 市场）

## 2. 数据清单

> 本表**暂无初始数据**，待确认富途 CS 市场标识与交易所代码的完整映射后补充；届时新增数据种子脚本（`Deploy/Migrations/V100000+` 段位），并同步更新本节表格。

| futu_cs_market | finv_exchange_code |
|----------------|-------------------:|
| （待补充） | |

## 3. 说明

- **映射方向**：`futu_cs_market`（富途行情源 CS 市场标识）→ `finv_exchange_code`（交易所代码）；富途侧标识按 TEXT 存储，兼容字符串取值。
- **关联字典**：`finv_exchange_code` → [FinvExchange.md](FinvExchange.md)（交易所/市场）；关联不建物理外键，由程序层控制（项目惯例）。
- **审计字段**：`gmt_create` / `gmt_update` 与既有表规范一致，`gmt_update` 由 `vq_set_gmt_update()` 触发器自动维护。
