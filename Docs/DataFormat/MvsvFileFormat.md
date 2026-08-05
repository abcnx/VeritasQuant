# MVSV-1 历史行情文件格式规范（MvsvFileFormat）

> 所属：FinvQuant 数据格式 · 存放：`Docs/DataFormat/`
> 适用范围：**所有向 FinvQuant 导入历史行情的导出程序**（脚本/服务/工具）必须按本规范生成 MVSV-1 文件。
> 服务端解析实现：`internal/mvsv/parser.go`（本规范与该实现严格一致，如有差异以代码为准）。
> 导入接口：`POST /API/V1/Quote/Import/Upload`（见 [ImportsUpload.md](../API/HistoryQuote/ImportsUpload.md)）。

---

## 1. 文件总览

MVSV-1（Minute Value Stream V1）是 FinvQuant 历史分钟行情的文本交换格式，一个文件恰好描述**一个证券**在**一段连续时间**内的分钟级 K 线。

```
┌─────────────────────────────────┐
│ 头部区：# Key : Value 行（若干） │  ← 声明元数据，含必填/可选键
│ 空行（1 个）                     │  ← 头部与数据区的分隔
│ 数据区：每行一条分钟 K 线        │  ← ts|...| 列布局由 # Field 决定
└─────────────────────────────────┘
```

### 1.1 编码与换行

| 项目 | 要求 | 强制/可选 |
|------|------|-----------|
| 字符编码 | **UTF-8**（无 BOM） | 强制 |
| 换行符 | `\n` 或 `\r\n`（解析时均兼容） | 强制 |
| 行尾 | 每行以换行符结尾；文件末尾最后一个数据行允许无换行 | 建议 |

### 1.2 头部行格式

- 每行头部必须以 **`# `**（井号 + 空格）开头；
- 键值之间以 **` : `**（空格-冒号-空格）分隔；
- 值**允许用双引号包裹**（推荐），解析器会自动去除首尾双引号；整数等值可不加引号；
- 头部键**不可重复**（重复会解析失败）；
- 头部区与数据区之间必须有一个**空行**分隔。

```
# Key : "Value"
```

> ⚠️ 注意：分隔符是 ` : `（含两侧空格），不是 `:`。`# Code:"NVDA"` 这类写法无法解析。

---

## 2. 头部键清单

### 2.1 必填头部键（10 个，缺一不可）

缺失任一必填键，文件整体解析失败（HTTP 422，`MVSV 解析失败: 缺少必填头部: <Key>`）。

| # | 键 | 值类型 | 示例 | 校验规则（不满足则报错） |
|---|-----|--------|------|--------------------------|
| 1 | `Format` | string | `"MVSV-1"` | **必须严格等于 `MVSV-1`** |
| 2 | `Field` | string | `"ts\|dt\|o\|c\|l\|h\|v\|t\|cp\|cr\|p"` | **必须匹配 §3 支持的一种列布局**，否则报「Field 布局不支持」 |
| 3 | `Count` | int | `15000` | **必须为非负整数**；且必须等于数据区实际记录行数（末尾不匹配报错） |
| 4 | `EffectiveTimeZone` | string | `"Asia/Shanghai"` / `"America/New_York"` | **必须为 IANA 合法时区名**（`time.LoadLocation` 可加载），用于 ts 一致性校验 |
| 5 | `Code` | string | `"NVDA"` / `"GCMain"` | 证券代码（通常与 FinvQuant 字典 usc 一致）；导入时作为 `secu_code` |
| 6 | `Market` | string | `"NSDQ"` / `"COMEX"` | 市场标识（描述性，不参与落表，仅头部保留） |
| 7 | `MarketCode` | int | `11` / `1320` | 市场数字代码（与 FinvQuant 市场编码体系对应；导入时参与一致性校验） |
| 8 | `CurrencyCode` | int | `55` | 货币代码（描述性） |
| 9 | `PriceAccuracy` | int | `3` / `1` | 价格精度（小数位；描述性，不参与计算） |
| 10 | `LotSize` | int | `1` / `100` | 每手股数/合约乘数（描述性） |

### 2.2 可选头部键（不影响解析，导出程序建议携带以便溯源）

以下键解析器**不强制**，但建议导出程序携带（内容会被保留在头部，部分用于后续扩展/审计）：

| 键 | 示例 | 说明 |
|----|------|------|
| `Title` | `"US_NVDA_Min_V4_2026"` | 文件标题/批次名 |
| `FieldType` | `"Int\|Long\|Decimal\|..."` | 各列类型声明（描述性） |
| `FieldName` | `"Ts\|DateTime\|Open\|..."` | 各列英文名（描述性） |
| `字段名称` | `"时间戳(UTC)\|日期时间\|..."` | 各列中文名（描述性；注意键可含中文） |
| `TimeZoneSource` | `"StockMapping.db secu_futu.tz ..."` | 时区来源说明 |
| `StockId` | `202597` | 行情源证券 ID |
| `FutuSymbol` | `"NVDA"` / `"GCmain"` | 富途符号（可能与 Code 大小写不同） |
| `Exchange` | `"US"` / `"COMEX"` | 交易所标识 |
| `InstrumentType` / `InstrumentTypeV2` | `3` / `10` | 证券类型编码 |
| `EngName` | `"NVIDIA"` | 英文名称 |
| `TimeZone` | `"America/New_York"` | 时区标识（与 EffectiveTimeZone 通常一致） |
| `DelistingFlag` | `0` | 退市标志 |
| `ListedExchange` / `ListedBoard` | `"NASDAQ"` / `""` | 上市交易所/板块 |
| `Region` | `"US"` | 地区 |
| `Name` | `"英伟达"` | 中文名称 |
| `Period` | `"Min"` | 周期（当前仅分钟） |
| `Start` / `End` | `"202601010500"` / `"202607210632"` | 数据起止时间（描述性） |
| `Size` | `195279` | 记录数（描述性，通常与 Count 一致） |
| `Dsv` | `3` | 数据源版本 |
| `Year` | `2026` | 数据年份 |

> 规则：**任何 `# Key : Value` 形式的额外头部键都会被接受并保留**，只要键不重复、格式合法。导出程序可自由扩展自定义键（如批次号、生成时间）。

---

## 3. 数据区列布局（# Field）

### 3.1 当前支持的两种布局（必须二选一）

#### 布局 A：11 列（典型文件：`US_NVDA_Min_V4_*`）

```
# Field : "ts|dt|o|c|l|h|v|t|cp|cr|p"
```

| 列序 | 列名 | 语义 | 类型 | 落表字段 |
|------|------|------|------|----------|
| 1 | `ts` | UTC 时间戳（秒） | Int | `ts`（主键） |
| 2 | `dt` | 本地日期时间，**14 位** `yyyyMMddHHmmss` | Long | `date` + `time`（拆分） |
| 3 | `o` | 开盘价 | Decimal | `open` |
| 4 | `c` | 收盘价 | Decimal | `close` |
| 5 | `l` | 最低价 | Decimal | `low` |
| 6 | `h` | 最高价 | Decimal | `high` |
| 7 | `v` | 成交量（股/手） | Long | `volume` |
| 8 | `t` | 成交额 | Decimal | `turnover` |
| 9 | `cp` | 涨跌值 | Decimal | **不落表**（忽略） |
| 10 | `cr` | 涨跌幅（%） | Decimal | **不落表**（忽略） |
| 11 | `p` | 前一收盘价 | Decimal | `prev_close` |

**数据行示例：**
```
1777405260|20260428154100|213.670966101|213.633309989|213.491175647|213.700931177|275700|58954901483|-0.0276|-0.012904|213.660877859
```

#### 布局 B：12 列（典型文件：`GCmain_Min_V3_*`；pc 列已过时移除）

```
# Field : "ts|d|t|o|c|l|h|v|a|cp|cr|p"
```

| 列序 | 列名 | 语义 | 类型 | 落表字段 |
|------|------|------|------|----------|
| 1 | `ts` | UTC 时间戳（秒） | Int | `ts`（主键） |
| 2 | `d` | 本地日期，**8 位** `yyyyMMdd` | Long | `date` |
| 3 | `t` | 本地时间，**6 位** `HHmmss` | Long | `time` |
| 4 | `o` | 开盘价 | Decimal | `open` |
| 5 | `c` | 收盘价 | Decimal | `close` |
| 6 | `l` | 最低价 | Decimal | `low` |
| 7 | `h` | 最高价 | Decimal | `high` |
| 8 | `v` | 成交量 | Long | `volume` |
| 9 | `a` | 成交额 | Decimal | `turnover` |
| 10 | `cp` | 涨跌值 | Decimal | **不落表**（忽略） |
| 11 | `cr` | 涨跌幅（%） | Decimal | **不落表**（忽略） |
| 12 | `p` | 前一收盘价 | Decimal | `prev_close` |

**数据行示例**（注意行尾允许残留旧 pc 空段 `\|`，解析器自动截断）：
```
1767243600|20260101|000000|4340|4907.5|4319.7|5626.8|5926343|0|575.4|13.282242|4332.1|
```

> 📌 **行尾空段容忍**：历史导出程序（GCmain 系列）曾在第 12 列后带过时的 `pc` 空列，数据行形如 `...|4332.1|`（末尾多一个 `\|`）。解析器会**自动截断行尾空段**，导出程序无需再输出该空段，输出也不报错。

> ⚠️ **布局 A 与布局 B 的 `t` 语义不同**：布局 A 的 `t` 是**成交额**（第 8 列），布局 B 的 `t` 是**时间**（第 3 列，配合第 2 列 `d`）。解析器按 `# Field` 声明识别，导出程序必须保证 `# Field` 与数据行列序严格一致。

### 3.2 数据行通用要求

| 项目 | 要求 | 强制/可选 |
|------|------|-----------|
| 列分隔符 | 竖线 `\|`，**每行列数必须与 `# Field` 声明完全一致** | 强制 |
| 行尾空段 | 布局 B 允许行尾残留旧 pc 空段（`...\|4332.1\|` 末尾多一个 `\|`），解析器自动截断；布局 A 不允许多余列 | 布局 B 可选 |
| 空值 | 数值列允许为空（解析为 NULL）；但 `ts`、`d`/`dt` 缺失会影响一致性校验（缺失时跳过校验） | 可选 |
| 数字格式 | 价格/成交额保留原始字符串精度（不做四舍五入），建议十进制原样输出 | 建议 |
| 成交量 | 非负整数；负数解析为 NULL | 建议 |
| 空行 | 数据区中的空行会被跳过（不计入 Count） | — |

### 3.3 ts 与本地时间一致性校验（强制）

解析器会对每行做校验：`ts` 换算到 `EffectiveTimeZone` 的本地时间，必须等于该行声明的本地时间（布局 A 的 `dt` 或布局 B 的 `d+t` 拼接的 14 位 `yyyyMMddHHmmss`）。**不一致则该行报错，整个文件导入失败。**

```
示例（布局 A）：
  ts=1785720600, EffectiveTimeZone=Asia/Shanghai
  → 本地时间 = 20260803093000 → 数据行 dt 必须为 20260803093000

示例（布局 B）：
  ts=1767337200, EffectiveTimeZone=America/New_York
  → 本地时间 = 20260102020000 → 数据行 d+t 必须为 20260102 + 020000
```

> 导出程序建议：**本地时间直接用 ts 换算得出**（`ts → EffectiveTimeZone → yyyyMMddHHmmss`），不要自行推导，避免时区/夏令时误差。`ts <= 0` 或本地时间字段缺失/不足 14 位时跳过该校验。

---

## 4. 完整文件示例

### 4.1 布局 A 完整示例（美股分钟线）

```
# Title : "US_NVDA_Min_V4_2026"
# Format : "MVSV-1"
# Field : "ts|dt|o|c|l|h|v|t|cp|cr|p"
# FieldType : "Int|Long|Decimal|Decimal|Decimal|Decimal|Decimal|Decimal|Decimal|Decimal|Decimal"
# FieldName : "Ts|DateTime|Open|Close|Low|High|Volume|Turnover|ChangePrice|ChangeRatio|PrevClose"
# 字段名称 : "时间戳(UTC)|日期时间|开盘价|收盘价|最低价|最高价|成交量|成交额|涨跌值|涨跌幅(%)|前一收盘价"
# Count : 3
# EffectiveTimeZone : "Asia/Shanghai"
# Code : "518880"
# Market : "SSE"
# MarketCode : 1
# CurrencyCode : 1
# PriceAccuracy : 3
# LotSize : 100

1785720600|20260803093000|7.001|7.002|6.999|7.003|100000|70010000000|0.001|0.000143|7.000
1785720660|20260803093100|7.002|7.001|7.000|7.003|80000|56010000000|0.000|-0.000143|7.002
1785720720|20260803093200|7.003|7.004|7.001|7.005|90000|63020000000|0.002|0.000286|7.001
```

### 4.2 布局 B 完整示例（期货分钟线）

```
# Title : "GCmain_Min_V3_2026_195279_2026072202"
# Format : "MVSV-1"
# Field : "ts|d|t|o|c|l|h|v|a|cp|cr|p"
# Count : 3
# EffectiveTimeZone : "America/New_York"
# TimeZoneSource : "JSON time_zone"
# StockId : 70000294
# FutuSymbol : "GCmain"
# Code : "GCMain"
# Market : "COMEX"
# MarketCode : 1320
# Exchange : "COMEX"
# PriceAccuracy : 1
# CurrencyCode : 55
# InstrumentType : 10
# InstrumentTypeV2 : 9
# LotSize : 100
# EngName : "Gold Futures (AUG6)"
# TimeZone : "America/New_York"
# DelistingFlag : 0
# ListedExchange : "COMEX"
# ListedBoard : ""
# Name : "黄金期货主连 (2608)"
# Period : "Min"
# Start : "202601010500"
# End : "202607210632"
# Size : 195279
# Dsv : 3
# Year : 2026

1767337200|20260102|020000|4340|4907.5|4319.7|5626.8|5926343|0|575.4|13.282242|4332.1|
1767337260|20260102|020100|4340|4343.9|4338.5|4349.5|314|0|11.8|0.272385|4332.1|
1767337320|20260102|020200|4342.7|4344|4342|4346.3|81|0|0.1|0.002302|4343.9|
```

---

## 5. 强制 / 可选要求汇总（导出程序速查）

### 🔴 强制要求（违反则整个文件导入失败）

| # | 要求 | 报错信息（节选） |
|---|------|------------------|
| 1 | 头部键必须以 `# ` 开头，键值用 ` : ` 分隔 | 头部格式必须为 '# Key : Value' |
| 2 | 头部键不可重复 | 存在重复头部键 |
| 3 | 头部区与数据区间必须有空行分隔 | 未找到头部与数据的空行分隔 |
| 4 | 头部区不能为空 | 头部为空 |
| 5 | `Format` 必须严格为 `MVSV-1` | Format 必须严格为 MVSV-1 |
| 6 | `Field` 必须是两种支持布局之一 | Field 布局不支持 |
| 7 | `Count` 必须为非负整数 | Count 必须为非负整数 |
| 8 | `Count` 必须等于数据区实际行数 | Count=…，实际记录数=… |
| 9 | `EffectiveTimeZone` 必须为 IANA 合法时区 | EffectiveTimeZone 非法 |
| 10 | 10 个必填头部键一个不能少 | 缺少必填头部 |
| 11 | 数据行列数必须与 `# Field` 完全一致 | 列数=…，期望 …（Field: …） |
| 12 | 每行 ts 与本地时间（dt 或 d+t）在 EffectiveTimeZone 下必须一致 | ts 与本地时间/EffectiveTimeZone 不一致 |

### 🟡 可选 / 建议要求（不满足不报错，但影响质量）

| # | 要求 | 说明 |
|---|------|------|
| 1 | 携带可选头部键（Title/StockId/FutuSymbol/Name 等） | 便于溯源与审计，见 §2.2 |
| 2 | 头部值统一用双引号包裹 | 与现有导出程序风格一致（解析器自动去引号） |
| 3 | 价格保留原始精度，不四舍五入 | 避免精度损失（落表为 NUMERIC(20,6) 以内的原始串） |
| 4 | 本地时间由 ts 换算得出，勿自行推导 | 避免时区/夏令时误差导致一致性校验失败 |
| 5 | 布局 B 数据行末尾可残留旧 pc 空段（行尾多一个 `\|`） | 与现网 GCmain 文件一致，解析器自动截断 |
| 6 | 使用 `\n` 换行 | `\r\n` 也兼容，但统一更佳 |

---

## 6. 与 FinvQuant 字典的关联（导出程序须知）

- **`Code`**：建议使用 FinvQuant `finv_security` 字典的 **usc**（如 `NVDA`、`GCMain`、`518880`），导入时可被 `Security/Lookup` 匹配；若用源证券代码（security_code），Lookup 也支持。
- **`MarketCode`**：建议与 FinvQuant `finv_market.market_code` 编码体系一致（如 11=上交所、1320=COMEX 期货等）；导入时会与表单/字典做一致性核对。
- 主表 `finv_quote_secu_kline_min` 不再存储 market_code（V21 起），市场信息由 `finv_security` 字典关联获取；文件头 `MarketCode` 仍参与导入一致性校验。

---

## 7. 常见失败场景（排障）

| 现象 | 根因 |
|------|------|
| `缺少必填头部: MarketCode` | 头部缺键（大小写敏感：`MarketCode` 不是 `marketcode`） |
| `Field 布局不支持: xxx` | `# Field` 拼写与两种布局不一致（含空格/大小写差异） |
| `Count=15000，实际记录数=14998` | 数据行数少 2 行（可能是文件截断或头部空行误算） |
| `ts 与本地时间/EffectiveTimeZone 不一致` | ts 与 dt/d+t 换算不一致（时区错/夏令时/本地时间手写错） |
| `第 N 行存在重复头部键` | 同一键出现两次 |
| `列数=12，期望 11` | 数据行多了一列（如布局 A 误带末尾 `\|`） |

---

## 8. 版本记录

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2026-08-06 | 初版：双布局（11 列/13 列）规范，必填/可选要求明细（对应 `internal/mvsv/parser.go` 双布局支持） |
| 1.1 | 2026-08-06 | 更正：布局 B 实为 12 列（`ts|d|t|o|c|l|h|v|a|cp|cr|p`，pc 已过时移除）；补充行尾空段容忍说明 |
