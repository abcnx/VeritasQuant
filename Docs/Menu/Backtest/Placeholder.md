# 仿真数据验证 / 模拟盘验证 / 实盘仿真验证 / 实盘交易（占位）— 菜单业务文档

> 所属：FinvQuant 前端菜单 · 存放：`Docs/Menu/Backtest/Placeholder.md`
> 对应视图：`Web/src/views/PlaceholderView.vue`（四个菜单共用）

## 1. 菜单入口

| 菜单 | key | 图标 | 路由 | 状态 |
|------|-----|------|------|------|
| 仿真数据验证 | `simulation-data` | `mdi-database-sync-outline` | `/meta/finvquant/simulation/data` | 🚧 规划中 |
| 模拟盘验证 | `simulation-paper` | `mdi-account-cash-outline` | `/meta/finvquant/simulation/paper` | 🚧 规划中 |
| 实盘仿真验证 | `simulation-live-sim` | `mdi-robot-outline` | `/meta/finvquant/simulation/live-sim` | 🚧 规划中 |
| 实盘交易 | `live-trading` | `mdi-cash-register` | `/meta/finvquant/live-trading` | 🚧 规划中 |

## 2. 业务功能概述（规划）

| 菜单 | 规划功能 |
|------|----------|
| 仿真数据验证 | 仿真数据源配置与接入、行情数据质量校验（缺失/异常/跳变）、真实 vs 仿真数据比对 |
| 模拟盘验证 | 模拟盘账户与虚拟资金、策略信号实时推送与自动撮合、模拟盘收益与回撤跟踪 |
| 实盘仿真验证 | 实盘行情流接入与重放、仿真撮合与订单生命周期、策略实时执行链路验证 |
| 实盘交易 | 经纪商/交易所接入与账户绑定、实盘下单/撤单与订单管理、实盘风控（限额/熔断/人工确认） |

> 上述能力与量化回测共用一套**环境模型**（`finv_quant_environment`，env_type 已包含
> PAPER / SIMULATION / LIVE）与策略模型，后续版本在现有架构上扩展。

## 3. 当前页面逻辑

占位页根据路由名展示对应模块的规划说明与功能清单（`PlaceholderView.vue` 内置 plans 映射），
不发起后端请求；页面含「规划中」标识，提示当前已落地能力与后续迭代方向。

## 4. 注意事项

- 实盘交易涉及真实资金，将在仿真验证全部通过后开放；
- 环境模型中的 PAPER / SIMULATION / LIVE 类型已可在「环境与模板管理」中先行维护配置。

## 5. 使用的后端接口

当前占位页不调用后端接口；后续实现时按 [ApiSpec.md](../../DevSpec/ApiSpec.md) 补充文档并登记。
