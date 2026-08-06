// FinvQuant 前端路由：每个业务菜单对应独立 URL 路径（多级菜单支持）
import { createRouter, createWebHistory } from 'vue-router'

import DashboardView from './views/Dashboard/DashboardView.vue'
import QuoteQueryView from './views/HistoryQuote/QuoteQueryView.vue'
import QuoteImportView from './views/HistoryQuote/QuoteImportView.vue'
import MetaExchangeView from './views/Metadata/MetaExchangeView.vue'
import MetaMarketView from './views/Metadata/MetaMarketView.vue'
import MetaSecurityView from './views/Metadata/MetaSecurityView.vue'
// 量化回测（通用量化策略验证）
import BacktestGoldFuturesView from './views/FinvQuant/Backtest/BacktestGoldFuturesView.vue'
import AccountManageView from './views/FinvQuant/Account/AccountManageView.vue'
import FundManageView from './views/FinvQuant/Fund/FundManageView.vue'
import PositionManageView from './views/FinvQuant/Position/PositionManageView.vue'
import StrategyManageView from './views/FinvQuant/Strategy/StrategyManageView.vue'
import BacktestAnalysisView from './views/FinvQuant/Backtest/BacktestAnalysisView.vue'
import BacktestReportView from './views/FinvQuant/Backtest/BacktestReportView.vue'
import EnvironmentTemplateView from './views/FinvQuant/Backtest/EnvironmentTemplateView.vue'
import PlaceholderView from './views/Common/PlaceholderView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', name: 'dashboard', component: DashboardView, meta: { title: '仪表盘' } },
    // 历史行情 → 历史行情查询
    { path: '/quote/query', name: 'quote-query', component: QuoteQueryView, meta: { title: '历史行情查询' } },
    // 元数据管理 → 业务元数据维护 → 四个子菜单
    { path: '/meta/exchange', name: 'meta-exchange', component: MetaExchangeView, meta: { title: '交易所信息维护' } },
    { path: '/meta/market', name: 'meta-market', component: MetaMarketView, meta: { title: '交易所下设市场信息维护' } },
    { path: '/meta/security', name: 'meta-security', component: MetaSecurityView, meta: { title: '规范证券信息维护' } },
    { path: '/meta/import', name: 'meta-import', component: QuoteImportView, meta: { title: '历史行情数据导入' } },
    // 量化策略验证 → 黄金期货合约回测验证
    {
      path: '/Meta/FinvQuant/Backtest/GoldFutures',
      name: 'backtest-gold-futures',
      component: BacktestGoldFuturesView,
      meta: { title: '黄金期货合约回测验证' },
    },
    // 环境与模板管理
    {
      path: '/Meta/FinvQuant/EnvTemplate',
      name: 'env-template',
      component: EnvironmentTemplateView,
      meta: { title: '环境与模板管理' },
    },
    // 账户管理 / 资金管理 / 持仓管理 / 策略管理 / 回测分析
    { path: '/Meta/FinvQuant/Account', name: 'account', component: AccountManageView, meta: { title: '账户管理' } },
    { path: '/Meta/FinvQuant/Fund', name: 'fund', component: FundManageView, meta: { title: '资金管理' } },
    { path: '/Meta/FinvQuant/Position', name: 'position', component: PositionManageView, meta: { title: '持仓管理' } },
    { path: '/Meta/FinvQuant/Strategy', name: 'strategy', component: StrategyManageView, meta: { title: '策略管理' } },
    { path: '/Meta/FinvQuant/Backtest/Analysis', name: 'backtest-analysis', component: BacktestAnalysisView, meta: { title: '回测分析' } },
    // 投资策略回测收益分析报告（独立页，任务列表跳转进入，?runId= 深链）
    {
      path: '/Meta/FinvQuant/Backtest/Analysis/Report',
      name: 'backtest-report',
      component: BacktestReportView,
      meta: { title: '投资策略回测收益分析报告' },
    },
    // 仿真数据验证 / 模拟盘验证 / 实盘仿真验证 / 实盘交易（规划占位）
    { path: '/Meta/FinvQuant/Simulation/Data', name: 'simulation-data', component: PlaceholderView, meta: { title: '仿真数据验证' } },
    { path: '/Meta/FinvQuant/Simulation/Paper', name: 'simulation-paper', component: PlaceholderView, meta: { title: '模拟盘验证' } },
    { path: '/Meta/FinvQuant/Simulation/LiveSim', name: 'simulation-live-sim', component: PlaceholderView, meta: { title: '实盘仿真验证' } },
    { path: '/Meta/FinvQuant/LiveTrading', name: 'live-trading', component: PlaceholderView, meta: { title: '实盘交易' } },
  ],
})

router.afterEach((to) => {
  const title = (to.meta.title as string) ?? ''
  document.title = title ? `${title} · FinvQuant` : 'FinvQuant 量化策略交易平台'
})

export default router
