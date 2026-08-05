# GET/POST /API/V1/Meta/FinvQuant/Backtest/Environment/* — 回测环境管理

环境接口：分页查询、详情、新增/修改、回测开关、删除。

> 环境 = 回测/模拟盘/仿真/实盘（BACKTEST/PAPER/SIMULATION/LIVE）× 地区/市场的运行配置，
> 驱动引擎自适应不同市场的交易约束与规则（交易时段、tick_size、成本基准、撮合模式、币种、偏好）。
> 环境配置结构见 [Docs/DevSpec/BacktestStrategySpec.md](../../DevSpec/BacktestStrategySpec.md) 第 10 章。

## 1. 分页查询环境

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Environment/List`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | string | 可选 | 所属用户（默认 `default`；`system` 内置环境全局可见，列表返回 system + 当前用户） |
| `env_type` | string | 可选 | BACKTEST / PAPER / SIMULATION / LIVE |
| `keyword` | string | 可选 | 匹配 env_code / env_name / description / region |
| `page` / `page_size` | int | 可选 | 分页 |

### 响应（成功 HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "total": 2,
    "list": [
      {
        "env_id": "e0000000-0000-4000-8000-000000000001",
        "env_code": "ENV-BT-COMEX-GC",
        "env_name": "COMEX 黄金期货回测环境",
        "env_type": "BACKTEST",
        "region": "US",
        "market_code": 0,
        "config": {
          "trading_sessions": [{ "start": "082000", "end": "133000" }],
          "trading_rules": { "t_plus": 0, "tick_size": 0.1, "contract_multiplier": 100, "limit_up_pct": 0, "limit_down_pct": 0 },
          "cost": { "commission_rate": 0.0003, "slippage_pct": 0.0001 },
          "fill_mode": "NEXT_BAR_OPEN",
          "currency": "USD",
          "preferences": { "date_format": "YYYY-MM-DD", "quote_direction": "RED_UP" }
        },
        "user_id": "system",
        "is_default": "1",
        "allow_backtest": "1",
        "status": "ENABLED",
        "description": "COMEX 黄金期货（GCMain 主连）回测默认环境",
        "created_by": "system"
      }
    ]
  }
}
```

## 2. 查询环境详情

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Environment/Get?env_id=xxx`

## 3. 新增/修改环境

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Environment/Save`

### 请求体（JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `env_id` | string | 可选 | 空则新增，非空则 UPDATE |
| `env_code` | string | ✅ | 环境编码（全局唯一） |
| `env_name` | string | ✅ | 环境名称 |
| `env_type` | string | 可选 | BACKTEST/PAPER/SIMULATION/LIVE（默认 BACKTEST） |
| `region` | string | 可选 | 地区（CN/US/HK...） |
| `market_code` | int | 可选 | 关联市场（默认 0=通用） |
| `config` | object | 可选 | 环境配置（trading_sessions 起止需 hhmmss，tick_size ≥ 0） |
| `user_id` | string | 可选 | 所属用户（默认 `default`） |
| `is_default` | string | 可选 | 是否默认环境 `0`/`1` |
| `allow_backtest` | string | 可选 | 回测开关 `0`/`1` |
| `status` | string | 可选 | DRAFT/ENABLED/DISABLED |
| `description` | string | 可选 | 说明 |

### 响应

```json
{ "code": 0, "message": "保存成功", "data": { "env_id": "e0000000-0000-4000-8000-000000000001" } }
```

## 4. 切换回测开关 / 5. 删除环境

- `POST /API/V1/Meta/FinvQuant/Backtest/Environment/Toggle`：`{ "env_id": "xxx", "allow_backtest": "0" }`
- `POST /API/V1/Meta/FinvQuant/Backtest/Environment/Delete`：`{ "env_id": "xxx" }`（已关联回测任务时禁止删除）

## 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 4001 | 请求体格式错误 |
| 2006 | 服务端处理失败（校验错误 / 数据库错误等） |

## 已使用位置（业务菜单）

| 业务菜单 | 菜单文档 | 使用接口 |
|----------|----------|----------|
| 环境与模板管理 | [Docs/Menu/Backtest/EnvironmentTemplate.md](../../Menu/Backtest/EnvironmentTemplate.md) | List / Get / Save / Toggle / Delete |
| 黄金期货合约回测验证 | [Docs/Menu/Backtest/BacktestGoldFutures.md](../../Menu/Backtest/BacktestGoldFutures.md) | List（环境下拉） |
