# GET/POST /API/V1/Meta/FinvQuant/Backtest/Strategy/* — 回测策略管理

结构化回测策略定义接口：分页查询、详情、新增/修改、回测开关、删除。

> 策略定义采用通用可扩展 JSON 模型（universe/data/indicators/signals/rules/risk/cost），
> 模型规范见 [Docs/DevSpec/BacktestStrategySpec.md](../../DevSpec/BacktestStrategySpec.md)。
> 保存时服务端对定义做结构校验与信号表达式编译校验。

## 1. 分页查询策略

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Strategy/List`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 可选 | 关键字，匹配 strategy_code / strategy_name / description / secu_code |
| `allow_backtest` | string | 可选 | 按回测开关过滤：`0`（关闭）/ `1`（开启） |
| `page` | int | 可选 | 页码（默认 1） |
| `page_size` | int | 可选 | 每页条数（默认 20，上限 500） |

### 响应（成功 HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "total": 2,
    "list": [
      {
        "strategy_id": "b0000000-0000-4000-8000-000000000001",
        "strategy_code": "STRAT-DUALMA-GC",
        "strategy_name": "GCMain 双均线交叉策略",
        "strategy_type": "RULE_BASED",
        "description": "双均线交叉策略（示例）",
        "definition": { "version": "1", "universe": { "securities": ["GCMain"] }, "indicators": [] },
        "definition_version": 1,
        "data_period": "Min",
        "secu_code": "GCMain",
        "user_id": "default",
        "template_id": null,
        "allow_backtest": "1",
        "status": "ENABLED",
        "created_by": "system"
      }
    ]
  }
}
```

## 2. 查询策略详情

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Strategy/Get?strategy_id=xxx`

## 3. 新增/修改策略

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Strategy/Save`

### 请求体（JSON，字段与 List 响应一致）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `strategy_id` | string | 可选 | 空则新增（自动生成 UUID），非空则 UPDATE |
| `strategy_code` | string | ✅ | 策略编码（全局唯一） |
| `strategy_name` | string | ✅ | 策略名称 |
| `strategy_type` | string | 可选 | 默认 `RULE_BASED`（当前版本仅实现该类型） |
| `description` | string | 可选 | 策略说明 |
| `definition` | object | ✅ | 结构化策略定义（JSON 模型 v1，保存时编译校验） |
| `definition_version` | int | 可选 | 定义版本（默认 1） |
| `data_period` | string | 可选 | 默认周期 Min/Hour/Day |
| `secu_code` | string | 可选 | 默认标的（缺省取 definition.universe.securities[0]） |
| `user_id` | string | 可选 | 所属用户（默认 `default`，多用户隔离） |
| `template_id` | string | 可选 | 来源模板（finv_quant_template） |
| `allow_backtest` | string | 可选 | 回测开关 `0`/`1`（默认 `1`） |
| `status` | string | 可选 | DRAFT/ENABLED/DISABLED（默认 ENABLED） |

### 响应（成功 HTTP 200）

```json
{ "code": 0, "message": "保存成功", "data": { "strategy_id": "b0000000-0000-4000-8000-000000000001" } }
```

## 4. 切换回测开关

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Strategy/Toggle`

### 请求体

```json
{ "strategy_id": "xxx", "allow_backtest": "0" }
```

## 5. 删除策略

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Strategy/Delete`

### 请求体

```json
{ "strategy_id": "xxx" }
```

> 策略已关联回测任务时禁止删除（返回业务错误），可改为禁用。

## 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 4001 | 请求体格式错误 |
| 2006 | 服务端处理失败（校验错误 / 数据库错误等，message 给出具体原因） |

## 已使用位置（业务菜单）

| 业务菜单 | 菜单文档 | 使用接口 |
|----------|----------|----------|
| 策略管理 | [Docs/Menu/Backtest/StrategyManage.md](../../Menu/Backtest/StrategyManage.md) | List / Get / Save / Toggle / Delete |
| 黄金期货合约回测验证 | [Docs/Menu/Backtest/BacktestGoldFutures.md](../../Menu/Backtest/BacktestGoldFutures.md) | List（策略下拉） |
