# GET /API/V1/Quote/Query — 历史行情查询

按**证券代码 + 交易日期**查询分钟级 K 线数据（周期目前仅支持 1 分钟 `Min`，其他周期暂不支持）。

## 请求

- **方法**：`GET`
- **路径**：`/API/V1/Quote/Query`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `secu_code` | string | ✅ | 证券代码（如 `NVDA` / `518880`） |
| `date` | int | ✅ | 交易日期（`yyyymmdd`，如 `20260803`） |
| `period` | string | 可选 | 周期，目前仅支持 `Min`（1 分钟，默认） |

## 响应

### 成功（HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "secu_code": "NVDA",
    "date": 20260803,
    "period": "Min",
    "count": 390,
    "bars": [
      {
        "time": 93000,
        "open": "213.6700",
        "high": "213.8000",
        "low": "213.5000",
        "close": "213.7500",
        "volume": 275700,
        "turnover": "70010000000",
        "change": "0.0890",
        "change_pct": "0.0417",
        "remark": null
      }
    ]
  }
}
```

| bars 字段 | 说明 |
|-----------|------|
| `time` | 交易时间（hhmmss） |
| `open` / `high` / `low` / `close` | 开 / 高 / 低 / 收（数值字符串；缺失为 null） |
| `volume` | 成交量（缺失为 null，前端有值才展示） |
| `turnover` | 成交额（缺失为 null，前端有值才展示） |
| `change` | 涨跌额 = close - prev_close（缺失为 null） |
| `change_pct` | 涨跌幅 %（缺失为 null） |
| `remark` | 导入备注（可为 null） |

### 失败

| 场景 | HTTP | code |
|------|------|------|
| `secu_code` 为空 / `date` 非法 | 400 | 4001 |
| `period` 暂不支持 | 500 | 2006 |
| 服务端内部错误 | 500 | 2006 |

## 示例

```bash
curl "http://localhost:16001/API/V1/Quote/Query?secu_code=NVDA&date=20260803&period=Min"
```

## 说明

- 涨跌额/涨跌幅由 `close` 与 `prev_close` 计算得出（`change = close - prev_close`）。
- 前端 K 线展示约定：红涨绿跌；悬停显示时间点 OHLC 及有值的成交量/成交额/涨跌额/涨跌幅。
