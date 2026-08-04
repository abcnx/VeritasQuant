# FinvExchange — 交易所/市场字典映射

> 所属：FinvQuant 数据字典映射 · 存放：`Docs/DataDictMapping/FinvExchange.md`
> 数据表：`finv_exchange`（表结构：[`Deploy/Migrations/V2__finv_exchange.sql`](../../Deploy/Migrations/V2__finv_exchange.sql)；初始数据：[`Deploy/Migrations/V100000__finv_exchange_seed.sql`](../../Deploy/Migrations/V100000__finv_exchange_seed.sql)）
> 用途：交易所/市场编码字典（单一事实来源），覆盖证券 / 期货 / 黄金及贵金属 / 场外 / 期权 / 外汇市场。

## 1. 表结构

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `exchange_code` | INTEGER | PK，1~999999 | 市场数字代码 |
| `exchange_flag` | TEXT | NOT NULL, UNIQUE | 交易所标志（如 `CN` / `SH` / `NSDQ` / `FX`） |
| `exchange_abbr` | TEXT | NOT NULL | 交易所英文缩写（如 `SSE` / `SZSE` / `HKEX`） |
| `exchange_name` | TEXT | NOT NULL | 交易所英文全称 |
| `exchange_abbr_cn` | TEXT | NOT NULL | 交易所中文缩写/名称（如 `上交所` / `纳斯达克`） |
| `en_market_type` | TEXT | NOT NULL | 市场类型（证券 / 期货 / 黄金及贵金属 / 场外 / 期权 / 外汇） |
| `region` | TEXT | NOT NULL | 地区编码（如 `CN` / `HK` / `USA` / `JP`） |
| `base_currency` | TEXT | NOT NULL | 基础计价货币（如 `CNY` / `USD` / `HKD`） |
| `ft_list_exchange_code` | TEXT | 可空 | FT 行情源列表交易所编码（映射预留，暂空，后续补充） |
| `gmt_create` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 首次插入时间 |
| `gmt_update` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 最后更新时间（触发器维护） |

索引：

- `idx_finv_exchange_region`：`(region, exchange_code)`
- `idx_finv_exchange_market_type`：`(en_market_type, exchange_code)`

## 2. 数据清单（22 条）

> `ft_list_exchange_code` 列暂为空，后续按 FT 行情源列表编码映射补充。

| exchange_code | exchange_flag | exchange_abbr | exchange_name | exchange_abbr_cn | en_market_type | region | base_currency | ft_list_exchange_code |
|--------------:|---------------|---------------|---------------|------------------|----------------|--------|---------------|-----------------------|
| 10 | CN | CN | 中国证券市场 | 中国证券市场(全体) | 证券 | CN | CNY | |
| 11 | SH | SSE | 上海证券交易所 | 上交所 | 证券 | CN | CNY | |
| 12 | SZ | SZSE | 深圳证券交易所 | 深交所 | 证券 | CN | CNY | |
| 13 | BJ | BSE | 北京证券交易所 | 北交所 | 证券 | CN | CNY | |
| 14 | SHFE | SHFE | 上海期货交易所 | 上海期货交易所 | 期货 | CN | CNY | |
| 15 | SGE | SGE | 上海黄金交易所 | 上海黄金交易所 | 黄金及贵金属 | CN | CNY | |
| 19 | CNOTC | CNOTC | 中国场外交易 | 中国场外交易 | 场外 | CN | CNY | |
| 21 | HK | HKEX | 香港联合证券交易所 | 港交所 | 证券 | HK | HKD | |
| 29 | TW | TWSE | 台湾证券交易所 | 台交所 | 证券 | TW | NTD | |
| 30 | US | US | 美国证券市场 | 美国证券市场(全体) | 证券 | USA | USD | |
| 31 | NSDQ | NSDQ | 纳斯达克证券交易所 | 纳斯达克 | 证券 | USA | USD | |
| 32 | NYSE | NYSE | 纽约证券交易所 | 纽交所 | 证券 | USA | USD | |
| 33 | COMEX | COMEX | 芝加哥商品期货交易所 | 芝商所 | 期货 | USA | USD | |
| 34 | COBE | COBE | 芝加哥期权交易所 | 芝加哥期权交易所 | 期权 | USA | USD | |
| 35 | PINK | PINK | 粉红单交易市场 | 粉单市场 | 场外 | USA | USD | |
| 51 | OSE | OSE | 大阪证券交易所 | 大阪证券交易所 | 证券 | JP | JPY | |
| 52 | TSE | TSE | 东京证券交易所 | 东京证券交易所 | 证券 | JP | JPY | |
| 53 | SGX | SGX | 新加坡证券交易所 | 新交所 | 证券 | SG | SGD | |
| 54 | INBSE | INBSE | 孟买证券交易所 | 孟买交易所 | 证券 | IN | INR | |
| 55 | NSE | NSE | 印度国家证券交易所 | 印度国家证券交易所 | 证券 | IN | INR | |
| 100 | FX | FX | 外汇市场 | 外汇交易市场 | 外汇 | FX | CNY | |
| 101 | FX-CFD | FX-CFD | 外汇-差价合约 | FX-CFD | 外汇 | FX | USD | |

## 3. 说明

- **迁移拆分**：表结构（`V2__finv_exchange.sql`）与初始数据（`V100000__finv_exchange_seed.sql`）分文件存放；数据种子统一使用 **V100000+ 段位**，确保在所有表结构脚本（V1~V99999）之后执行，后续新增表结构/变更脚本不受影响。
- **数据来源**：FT 交易所清单；原始数据中 `exchange_code=19` 存在 `OTC` / `CNOTC` 两条冲突记录，经 ACANX 确认**仅保留 `19 CNOTC 中国场外交易`**，共 22 条。
- **映射预留**：`ft_list_exchange_code` 字段已建列、当前留空，待 FT 行情源列表编码映射确定后补充（`UPDATE finv_exchange SET ft_list_exchange_code = ... WHERE exchange_code = ...`）。
- **幂等插入**：初始数据使用 `INSERT ... ON CONFLICT (exchange_code) DO NOTHING`，重复执行不产生重复记录。
- **审计字段**：`gmt_create` / `gmt_update` 与既有表规范一致，`gmt_update` 由 `vq_set_gmt_update()` 触发器自动维护。
