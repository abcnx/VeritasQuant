# FinvMarket — 交易市场字典映射

> 所属：FinvQuant 数据字典映射 · 存放：`Docs/DataDictMapping/FinvMarket.md`
> 数据表：`finv_market`（表结构：[`Deploy/Migrations/V3__finv_market.sql`](../../Deploy/Migrations/V3__finv_market.sql)；初始数据：[`Deploy/Migrations/V100015__finv_market_seed.sql`](../../Deploy/Migrations/V100015__finv_market_seed.sql)；二次定版：[`Deploy/Migrations/V100017__finv_market_seed_full.sql`](../../Deploy/Migrations/V100017__finv_market_seed_full.sql)）
> 用途：交易所下属交易市场代码表（如 上交所股票/基金/债券等细分市场），与 [`FinvExchange.md`](FinvExchange.md) 交易所字典互补。

## 1. 表结构

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `market_code` | INTEGER | PK，1~999999 | 市场数字代码 |
| `market_flag` | TEXT | NOT NULL, UNIQUE | 市场标识（如 `SH_A` / `SZ_STAR` 等，编码待定） |
| `market_abbr` | TEXT | NOT NULL | 交易所简码 |
| `market_name` | TEXT | NOT NULL | 交易所名称 |
| `en_security_type` | TEXT | NOT NULL | 允许证券类型（如 `STOCK` / `FUND` / `BOND` 等） |
| `base_currency` | TEXT | NOT NULL | 基础计价货币（如 `CNY` / `USD` / `HKD`） |
| `gmt_create` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 首次插入时间 |
| `gmt_update` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 最后更新时间（触发器维护） |

索引：

- `idx_finv_market_security_type`：`(en_security_type, market_code)`
- `idx_finv_market_base_currency`：`(base_currency, market_code)`

## 2. 数据清单（54 条）

> V100017 二次定版（ACANX 2026-08-05 再次发放），与 V100015 逐条核对 **54/54 一致**。市场代码采用**重新编码体系**（非富途原始值），如 `1100`=外汇、`1110`=上交所 A 股、`1120`=深交所 A 股、`1210`=港股主板、`1310`=美股指数、`2000`=加拿大 TSX 主板、`8300`=日本、`10001`=加密货币；`market_flag` = exchange + '_' + market_code；`en_security_type` 暂填 market_code 占位，`base_currency` 暂空串（待后续按市场补齐）。完整清单见迁移文件 [`V100017__finv_market_seed_full.sql`](../../Deploy/Migrations/V100017__finv_market_seed_full.sql)。

## 3. 说明

- **对齐参考**：MySQL `finv_market` 表结构；PG 侧按项目建表规范加固（主键、NOT NULL、审计字段、触发器）。
- **与 finv_exchange 关系**：`finv_exchange` 为交易所/市场整体字典（code 10/11/12...），`finv_market` 为交易所下属细分交易市场代码表，两者通过市场编码语义关联（映射规则待定）。
- **⚠️ 与富途映射表编码不一致**：`finv_futu_mapping_market_code.finv_market_code`（暂=futu 原始值，如 1/10/30/60/200...）与 finv_market 重新编码（1100/1110/1310/2000...）**不一致**，需后续迁移按 market_abbr + market_name 语义对齐映射表。
- **审计字段**：`gmt_create` / `gmt_update` 与既有表规范一致，`gmt_update` 由 `vq_set_gmt_update()` 触发器自动维护。
