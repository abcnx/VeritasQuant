# POST /API/V1/Quote/Import/Upload — 历史行情文件导入

上传 MVSV-1 分钟级历史行情文件，服务端解析后**字段级覆盖**导入 PostgreSQL `finv_quote_secu_kline_min` 表（主键 `ts + market_code + secu_code`）。

## 请求

- **方法**：`POST`
- **路径**：`/API/V1/Quote/Import/Upload`
- **内容类型**：`multipart/form-data`
- **文件大小上限**：50 MiB

### Form 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | file | ✅ | MVSV-1 行情文件（如 `US_NVDA_Min_V4_2026_2026072907_15000.mvsv`） |
| `market_code` | string | 可选 | 市场数字代码（登记字段；填写后与文件头部 `MarketCode` 校验一致性） |
| `secu_code` | string | 可选 | 证券代码（登记字段；填写后与文件头部 `Code` 校验一致性） |
| `source` | string | 可选 | 数据源名称（默认 `upload`） |
| `upsert_mode` | string | 可选 | 覆盖模式：`FIELD`（字段级，默认）/ `ROW`（整行覆盖） |
| `imported_by` | string | 可选 | 导入人标识（默认 `gui`） |

## 响应

### 成功（HTTP 200）

```json
{
  "code": 0,
  "message": "行情导入完成",
  "data": {
    "batch_id": "import_NVDA_20260804080500",
    "secu_code": "NVDA",
    "market_code": 11,
    "record_count": 15000,
    "inserted": 15000,
    "updated": 0,
    "mode": "FIELD"
  }
}
```

| data 字段 | 说明 |
|-----------|------|
| `batch_id` | 导入批次 ID（对应 `finv_quote_ingest_batches`） |
| `secu_code` / `market_code` | 证券代码 / 市场数字代码（来自文件头部 `Code` / `MarketCode`） |
| `record_count` | 本次导入记录数 |
| `inserted` | 新增行数 |
| `updated` | 覆盖行数（>0 时自动写入 `finv_quote_revision_log` 修正审计） |
| `mode` | 实际覆盖模式 |

### 失败（HTTP 422）

```json
{
  "code": 4001,
  "message": "MVSV 解析失败: TEST.mvsv: 缺少必填头部: MarketCode",
  "error": {
    "code": "DATA_IMPORT_CONTRACT_INVALID",
    "catalog_version": "1.0",
    "retryable": false
  }
}
```

| 场景 | HTTP | code |
|------|------|------|
| 缺文件字段 / 文件为空 / 超 50 MiB / 模式非法 / MVSV 解析失败 | 422 | 4001 |
| 服务端内部错误（读文件/导入失败） | 500 | 2006 |

## 示例

### cURL

```bash
curl -X POST http://localhost:16001/API/V1/Quote/Import/Upload \
  -F "file=@VeritasQuant/Data/US_NSDQ_NVDA/US_NVDA_Min_V4_2026_2026072907_15000.mvsv" \
  -F "source=cn-feed" \
  -F "upsert_mode=FIELD"
```

### 前端（Vue）

```typescript
const form = new FormData()
form.append('file', file)
form.append('source', 'cn-feed')
form.append('upsert_mode', 'FIELD')
const response = await fetch('/API/V1/Quote/Import/Upload', { method: 'POST', body: form })
const body = await response.json()
```

## 说明

- 导入按主键 `(ts, market_code, secu_code)` 覆盖同键数据；`FIELD` 模式只覆盖新数据有值的字段（NULL 保留旧值）。
- 每次导入自动登记批次（`finv_quote_ingest_batches`）；发生覆盖时写入修正审计（`finv_quote_revision_log`），可追溯"改了哪些行、为什么"。
- 支持分批上传多个文件：每文件一次上传即一个批次，重复上传同键数据按覆盖模式更新。
