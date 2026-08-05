# GET/POST /API/V1/Meta/FinvQuant/Backtest/Template/* — 回测模板管理

模板接口：分页查询、详情、新增/修改、删除。

> 模板 = 策略/账户/环境三类可复用配置（STRATEGY/ACCOUNT/ENVIRONMENT），
> 相同部分（环境、约束、规则、限制、策略）复用、差异部分自定义；
> 内置模板（is_builtin=1，user_id='system'）全局可见且禁止删除。

## 1. 分页查询模板

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Template/List`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | string | 可选 | 所属用户（默认 `default`；`system` 内置模板全局可见） |
| `template_type` | string | 可选 | STRATEGY / ACCOUNT / ENVIRONMENT |
| `keyword` | string | 可选 | 匹配 template_code / template_name / description |
| `page` / `page_size` | int | 可选 | 分页 |

### 响应（成功 HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "total": 3,
    "list": [
      {
        "template_id": "t0000000-0000-4000-8000-000000000002",
        "template_code": "TPL-STRAT-DUALMA",
        "template_name": "双均线交叉策略模板",
        "template_type": "STRATEGY",
        "content": { "version": "1", "universe": { "securities": ["GCMain"] }, "indicators": [] },
        "user_id": "system",
        "is_builtin": "1",
        "status": "ENABLED",
        "description": "内置：双均线交叉策略模板（GCMain 示例）",
        "created_by": "system"
      }
    ]
  }
}
```

## 2. 查询模板详情

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Template/Get?template_id=xxx`

## 3. 新增/修改模板

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Template/Save`

### 请求体（JSON）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `template_id` | string | 可选 | 空则新增，非空则 UPDATE |
| `template_code` | string | ✅ | 模板编码（全局唯一） |
| `template_name` | string | ✅ | 模板名称 |
| `template_type` | string | ✅ | STRATEGY / ACCOUNT / ENVIRONMENT |
| `content` | object | ✅ | 模板内容（STRATEGY=策略定义 / ACCOUNT=账户配置 / ENVIRONMENT=环境配置） |
| `user_id` | string | 可选 | 所属用户（默认 `default`） |
| `is_builtin` | string | 可选 | 是否内置 `0`/`1` |
| `status` | string | 可选 | DRAFT/ENABLED/DISABLED |
| `description` | string | 可选 | 说明 |

### 响应

```json
{ "code": 0, "message": "保存成功", "data": { "template_id": "t0000000-0000-4000-8000-000000000002" } }
```

## 4. 删除模板

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Template/Delete`
- **请求体**：`{ "template_id": "xxx" }`
- **说明**：内置模板（is_builtin='1'）禁止删除。

## 错误码

| code | HTTP | 说明 |
|------|------|------|
| 0 | 200 | 成功 |
| 4001 | 400 | 请求体格式错误 / 参数校验错误（必填缺失、表达式错误、枚举越界等） |
| 4004 | 404 | 资源不存在 |
| 4009 | 409 | 状态冲突 / 禁止操作（已关联任务禁止删除、内置模板禁止修改、无权访问他人数据等） |
| 2006 | 500 | 其他服务端错误（message 给出具体原因） |

## 已使用位置（业务菜单）

| 业务菜单 | 菜单文档 | 使用接口 |
|----------|----------|----------|
| 环境与模板管理 | [Docs/Menu/Backtest/EnvironmentTemplate.md](../../Menu/Backtest/EnvironmentTemplate.md) | List / Get / Save / Delete |
