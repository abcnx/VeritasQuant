# GET/POST /API/V1/Meta/FinvQuant/Backtest/Run/* — 回测任务查询与数据接口

回测任务生命周期接口：任务列表 / 详情 / 取消，以及报告、曲线、成交、链路追踪数据查询。

## 1. 分页查询任务

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/List`

### Query 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `status` | string | 可选 | PENDING / RUNNING / SUCCEEDED / FAILED / CANCELLED |
| `secu_code` | string | 可选 | 按标的过滤 |
| `strategy_id` | string | 可选 | 按策略过滤 |
| `keyword` | string | 可选 | 匹配 run_no / strategy_name / account_name / secu_code |
| `page` / `page_size` | int | 可选 | 分页（默认 1 / 20） |

### 响应（成功 HTTP 200）

```json
{
  "code": 0,
  "message": "查询完成",
  "data": {
    "total": 3,
    "list": [
      {
        "run_id": "xxxxx", "run_no": 1, "user_id": "default",
        "strategy_id": "xxx", "strategy_code": "STRAT-DUALMA-GC", "strategy_name": "GCMain 双均线交叉策略",
        "account_id": "xxx", "account_code": "ACCT-GOLD-001", "account_name": "黄金期货回测账户",
        "env_id": "xxx", "environment_snapshot": { "env_code": "ENV-BT-COMEX-GC", "config": {} },
        "secu_code": "GCMain", "market_code": 33, "period": "Min", "report_precision": "Day",
        "start_ts": 1514764800, "end_ts": 1753920000, "start_date": 20180101, "end_date": 20260731,
        "status": "SUCCEEDED", "progress": 100, "error_message": "",
        "report": { "total_return_pct": 12.34, "final_equity": 112340, "event_stats": {} },
        "started_at": "2026-08-06T04:00:00Z", "finished_at": "2026-08-06T04:01:23Z"
      }
    ]
  }
}
```

## 2. 查询任务详情

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/Get?run_id=xxx`

## 3. 取消任务

- **方法**：`POST`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/Cancel`
- **请求体**：`{ "run_id": "xxx" }`
- **说明**：仅可取消运行中的任务（PENDING/RUNNING），取消后状态置为 CANCELLED。

## 4. 查询回测报告（汇总指标）

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/Report?run_id=xxx`
- **说明**：仅 SUCCEEDED 任务可查。返回 `RunReport`（初始资金/期末总资产/总收益额/总收益率/年化收益率/
  最大回撤（含区间）/夏普比率/年化波动率/最大投入/平均投入/持仓天数/胜率/盈亏比/手续费/单期收益分布/
  信号归因/事件统计 event_stats 等）。

## 5. 净值曲线数据（余额/收益率/收益额/持仓金额/回撤）

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/Equity?run_id=xxx&page=&page_size=`
- **说明**：按报告精度（Day/Hour/Min）逐点返回；`equity`（总资产=现金+持仓市值）、`cash`、`position_value`、
  `position_qty`、`profit`（累计收益额）、`roi`（累计收益率 %）、`drawdown`（回撤 %）。

## 6. 成交记录

- **方法**：`GET`
- **路径**：`/API/V1/Meta/FinvQuant/Backtest/Run/Trades?run_id=xxx&page=&page_size=`
- **说明**：时间/方向/价格/数量/金额/手续费/平仓盈亏/成交后持仓与现金/触发信号/引擎内序号 seq。

## 7. 链路追踪数据（需求⑨）

| 接口 | 路径 | 说明 |
|------|------|------|
| 资金流水明细 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/Cashflows?run_id=xxx` | INITIAL_DEPOSIT/BUY_PAY/SELL_RECEIVE/FEE/MARGIN_HOLD/MARGIN_RELEASE，含变动前后现金 |
| 持仓变化明细 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/PositionLogs?run_id=xxx` | OPEN/ADD/REDUCE/CLOSE，含变动前后数量与加权成本 |
| 事件追踪 | `GET /API/V1/Meta/FinvQuant/Backtest/Run/EventTraces?run_id=xxx` | 触发原因/成交结果（FILLED/REJECTED/EXPIRED/PENDING）/委托耗时（bar·秒）/未成交原因 |

## 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 2006 | 查询失败 / 任务不存在 / 任务未完成时查报告等 |

## 已使用位置（业务菜单）

| 业务菜单 | 菜单文档 | 使用接口 |
|----------|----------|----------|
| 回测分析 | [Docs/Menu/Backtest/BacktestAnalysis.md](../../Menu/Backtest/BacktestAnalysis.md) | List / Get / Report / Equity / Trades / Cashflows / PositionLogs / EventTraces |
| 黄金期货合约回测验证 | [Docs/Menu/Backtest/BacktestGoldFutures.md](../../Menu/Backtest/BacktestGoldFutures.md) | List / Get |
| 资金管理 | [Docs/Menu/Backtest/FundManage.md](../../Menu/Backtest/FundManage.md) | List / Equity |
| 持仓管理 | [Docs/Menu/Backtest/PositionManage.md](../../Menu/Backtest/PositionManage.md) | List / Equity / Trades |
