# FinvQuant 服务端 API 文档

> 基路径：`/API/V1`（统一大写，见 [Docs/DevSpec/ApiSpec.md](../DevSpec/ApiSpec.md)）
> 服务端端口：**16001**；默认地址：`http://localhost:16001`

## 通用约定

### 响应信封

所有 REST JSON 响应顶层固定输出：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `code` | int | ✅ | 业务码；成功码集合 `{0, 1, 200, 202}` |
| `message` | string | ✅ | 文本消息 |
| `data` | object | 可选 | 业务数据 |
| `error` | object | 可选 | 错误时必填；含 `code`、`catalog_version`、`retryable` |
| `details` | object | 可选 | 补充详情 |
| `request_id` | string | 可选 | 请求追踪 ID |
| `trace_id` | string | 可选 | 链路追踪 ID |

所有 wire 字段使用 **snake_case**。

### 错误码

| 号段 | 含义 |
|------|------|
| `{0, 1, 200, 202}` | 成功 |
| `1000-2999` | 平台、安全和依赖错误 |
| `≥3000` | 项目自定义业务错误码 |

（详见 [Docs/DevSpec/ErrorCodeSpec.md](../DevSpec/ErrorCodeSpec.md)）

## 端点索引

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/API/V1/health/live` | 存活探针 |
| GET | `/API/V1/health/ready` | 就绪探针（PG + Redis） |
| GET | `/API/V1/version` | 服务端版本信息 |
| POST | `/API/V1/Quote/Import/Upload` | 上传 MVSV 历史行情文件并导入 PG（[详细文档](ImportsUpload.md)） |
| GET | `/API/V1/Quote/Query` | 按证券代码+日期查询分钟级 K 线（周期 Min）（[详细文档](HistoryQuote/HistoryQuote.md)） |
