# FinvRegion — 区域字典映射

> 所属：FinvQuant 数据字典映射 · 存放：`Docs/DataDictMapping/FinvRegion.md`
> 数据表：`finv_region`（表结构：[`Deploy/Migrations/V6__finv_region.sql`](../../Deploy/Migrations/V6__finv_region.sql)）
> 用途：区域代码字典（idx 数字序号 + region 区域简写），供交易所/市场按区域归类与查询。

## 1. 表结构

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `idx` | INTEGER | PK，1~999999 | 区域序号（数字） |
| `region` | TEXT | NOT NULL, UNIQUE | 区域简写（如 `CN` / `HK` / `USA` / `JP`） |
| `gmt_create` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 首次插入时间 |
| `gmt_update` | TIMESTAMPTZ | NOT NULL DEFAULT now() | 最后更新时间（触发器维护） |

## 2. 数据清单

> 本表**暂无初始数据**，待后续补充；届时新增数据种子脚本（`Deploy/Migrations/V100000+` 段位，见 [Prompt.md](../../Prompt.md) 迁移分段约定），并同步更新本节表格。

| idx | region |
|----:|--------|
| （待补充） | |

## 3. 说明

- **关联参考**：[`FinvExchange.md`](FinvExchange.md) 中 `finv_exchange.region`（如 `CN` / `HK` / `USA` / `JP`）可对齐本表 `region` 建立区域维度归类（关联规则待定，不建物理外键，程序层控制）。
- **审计字段**：`gmt_create` / `gmt_update` 与既有表规范一致，`gmt_update` 由 `vq_set_gmt_update()` 触发器自动维护。
