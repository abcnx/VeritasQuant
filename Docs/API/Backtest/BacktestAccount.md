# GET/POST /API/V1/Meta/FinvQuant/Backtest/Account/* — 回测账户管理

回测账户接口：分页查询、详情、新增/修改、回测开关、删除。

> 账户 = 回测运行的「初始资金 + 交易成本 + 保证金模式」基线配置；支持多用户（user_id）与
> 单用户多子账户（group_id 分组）。

## 1. 分页查询账户

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Account/List`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 可选 | 匹配 account_code / account_name / remark |
| `allow_backtest` | string | 可选 | 回测开关过滤 `0`/`1` |
| `page` / `page_size` | int | 可选 | 分页（默认 1 / 20，上限 500） |

### 响应（成功 HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "total": 1,
    "list": [
      {
        "account_id": "a0000000-0000-4000-8000-000000000001",
        "account_code": "ACCT-GOLD-001",
        "account_name": "黄金期货回测账户",
        "user_id": "default",
        "group_id": null,
        "env_id": null,
        "initial_capital": 100000,
        "currency_type": "USD",
        "commission_rate": 0.0003,
        "slippage_pct": 0.0001,
        "margin_mode": "FULL",
        "margin_rate": 1,
        "allow_backtest": "1",
        "status": "ENABLED",
        "remark": "默认回测账户（GCMain 黄金期货主连）",
        "created_by": "system"
      }
    ]
  }
}
```

## 2. 查询账户详情

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Account/Get?account_id=xxx`

## 3. 新增/修改账户

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Account/Save`

### 请求体（JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `account_id` | string | 可选 | 空则新增，非空则 UPDATE |
| `account_code` | string | ✅ | 账户编码（全局唯一） |
| `account_name` | string | ✅ | 账户名称 |
| `user_id` | string | 可选 | 所属用户（默认 `default`，多用户隔离） |
| `group_id` | string | 可选 | 子账户分组 / 主账户归属（单用户多子账户，NULL=主账户） |
| `env_id` | string | 可选 | 默认关联环境（缺省回测时取默认环境） |
| `initial_capital` | number | ✅ | 初始启动资金（>0） |
| `currency_type` | string | 可选 | 计价币种（默认 USD） |
| `commission_rate` | number | 可选 | 手续费率（按成交金额比例，默认 0） |
| `slippage_pct` | number | 可选 | 滑点（按成交价比例，默认 0） |
| `margin_mode` | string | 可选 | FULL 全额 / FUTURES 期货保证金（预留，默认 FULL） |
| `margin_rate` | number | 可选 | 保证金比例（默认 1） |
| `allow_backtest` | string | 可选 | 回测开关 `0`/`1` |
| `status` | string | 可选 | DRAFT/ENABLED/DISABLED |
| `remark` | string | 可选 | 备注 |

### 响应

```json
{ "code": 0, "message": "保存成功", "data": { "account_id": "a0000000-0000-4000-8000-000000000001" } }
```

## 4. 切换回测开关 / 5. 删除账户

- `POST /API/V1/Meta/FinvQuant/Backtest/Account/Toggle`：`{ "account_id": "xxx", "allow_backtest": "0" }`
- `POST /API/V1/Meta/FinvQuant/Backtest/Account/Delete`：`{ "account_id": "xxx" }`（已关联回测任务时禁止删除）

## 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 4001 | 请求体格式错误 |
| 2006 | 服务端处理失败（校验错误 / 数据库错误等） |

## 已使用位置（业务菜单）

| 业务菜单 | 菜单文档 | 使用接口 |
|----------|----------|----------|
| 账户管理 | [Docs/Menu/Backtest/AccountManage.md](../../Menu/Backtest/AccountManage.md) | List / Get / Save / Toggle / Delete |
| 黄金期货合约回测验证 | [Docs/Menu/Backtest/BacktestGoldFutures.md](../../Menu/Backtest/BacktestGoldFutures.md) | List（账户下拉） |
