<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { apiGet } from '../../../../../api'
import { fmtDate, fmtNum, fmtPct, fmtTime } from '../../../../../utils'

const route = useRoute()
const router = useRouter()

interface Report {
  secu_code: string
  period: string
  report_precision: string
  start_date: number
  end_date: number
  bar_count: number
  initial_capital: number
  final_equity: number
  total_profit: number
  total_return_pct: number
  annual_return_pct: number
  max_drawdown_pct: number
  sharpe_ratio: number
  volatility_pct: number
  trade_count: number
  buy_count: number
  sell_count: number
  win_count: number
  loss_count: number
  win_rate_pct: number
  profit_factor: number
  total_fee: number
  max_invested: number
  avg_invested: number
  invested_days: number
  best_day_pct: number
  worst_day_pct: number
  profit_days: number
  loss_days: number
  trade_signal_detail: Record<string, number>
  event_stats: EventStats | null
  generated_at: string
}

interface EventStats {
  trigger_count: number
  filled_count: number
  rejected_count: number
  expired_count: number
  pending_count: number
  avg_latency_bars: number
  avg_latency_sec: number
  reject_reasons: Record<string, number>
  trigger_reasons: Record<string, number>
}

interface EquityPoint {
  seq: number
  ts: number
  date: number
  time: number
  equity: number
  cash: number
  position_value: number
  position_qty: number
  profit: number
  roi: number
  drawdown: number
}

interface TradeRow {
  trade_id: number
  seq: number
  ts: number
  date: number
  time: number
  action: string
  price: number
  qty: number
  amount: number
  fee: number
  profit: number
  position_after: number
  cash_after: number
  signal: string
  remark?: string
}

interface CashflowRow {
  cashflow_id: number
  seq: number
  date: number
  time: number
  flow_type: string
  amount: number
  cash_before: number
  cash_after: number
  trade_id: number
  remark: string
}

interface PositionLogRow {
  log_id: number
  seq: number
  date: number
  time: number
  action: string
  price: number
  qty: number
  position_before: number
  position_after: number
  avg_cost_before: number
  avg_cost_after: number
  trade_id: number
  remark: string
}

interface EventTraceRow {
  event_id: number
  seq: number
  action: string
  trigger_reason: string
  trigger_date: number
  trigger_time: number
  order_date?: number
  order_time?: number
  exec_status: string
  exec_date: number
  exec_time: number
  latency_bars: number
  latency_sec: number
  alive_sec?: number
  reject_reason: string
  price: number
  qty: number
  trade_id: number
}

// ---------------------------------------------------------------------
// 回测背景信息类型（Run/Get 返回完整 Run 对象；strategy_snapshot 为 StrategyDefinition 序列化）
// ---------------------------------------------------------------------

interface StrategySnapshot {
  version?: string
  strategy_type?: string
  description?: string
  universe?: { securities?: string[] }
  data?: { period?: string; price_field?: string; warmup_bars?: number; fill_mode?: string }
  indicators?: IndicatorDef[]
  signals?: { buy?: string; sell?: string }
  rules?: { buy?: RuleDef; sell?: RuleDef }
  risk?: RiskDef
  cost?: CostDef
}

interface IndicatorDef {
  id?: string
  type?: string
  params?: Record<string, unknown>
}

interface RuleDef {
  action?: string
  quantity_type?: string // ALL_IN/ALL/FIXED/PERCENT/AMOUNT
  quantity?: number
  max_per_day?: number
  allowed_times?: string[] // hhmmss 字符串
  max_per_run?: number
  allow?: boolean
  max_per_week?: number
  max_per_month?: number
  max_fee_per_window?: number
  fee_window_days?: number
}

interface BuilderDef {
  enabled?: boolean
  target_position_pct?: number
  tranches?: number
  tranche_interval_bars?: number
}

interface RiskDef {
  stop_loss_pct?: number
  take_profit_pct?: number
  max_position_pct?: number
  max_positions?: number
  max_trades_per_day?: number
  min_interval_bars?: number
  builder?: BuilderDef | null
  reduce_tranches?: number
  max_trades_per_week?: number
  max_trades_per_month?: number
  max_fee_per_window?: number
  fee_window_days?: number
}

interface CostDef {
  commission_rate?: number // 比值（0.0003 = 0.03%）
  slippage_pct?: number
}

interface AccountSnapshot {
  account_id: string
  account_code: string
  account_name: string
  user_id: string
  group_id?: string | null
  initial_capital: number
  currency_type: string
  commission_rate: number
  slippage_pct: number
  margin_mode: string // FULL/FUTURES
  margin_rate: number
}

interface EnvironmentConfig {
  trading_sessions?: { start: string; end: string }[]
  trading_rules?: {
    t_plus?: number
    tick_size?: number
    contract_multiplier?: number
    limit_up_pct?: number
    limit_down_pct?: number
  }
  cost?: CostDef
  fill_mode?: string
  currency?: string
  preferences?: Record<string, unknown>
}

interface EnvironmentSnapshot {
  env_id: string
  env_code: string
  env_name: string
  env_type: string // BACKTEST/PAPER/SIMULATION/LIVE
  region: string
  market_code: number
  config: EnvironmentConfig
  user_id: string
  is_default: string
  allow_backtest: string
  status: string
  description: string
}

interface RunOptions {
  enable_backtest?: boolean
  report_precision?: string
  initial_capital?: number | null
  commission_rate?: number | null
  slippage_pct?: number | null
  max_trades_per_day?: number | null
  allowed_times?: string[]
}

interface RunDetail {
  run_id: string
  run_no: number
  user_id: string
  strategy_id: string
  strategy_code: string
  strategy_name: string
  strategy_snapshot: StrategySnapshot
  account_id: string
  account_code: string
  account_name: string
  account_snapshot: AccountSnapshot
  env_id: string
  environment_snapshot: EnvironmentSnapshot | null
  secu_code: string
  market_code: number
  period: string
  report_precision: string
  start_ts: number
  end_ts: number
  start_date: number
  end_date: number
  options: RunOptions
  status: string
  progress: number
  error_message: string
  started_at?: string | null
  finished_at?: string | null
  created_by: string
}

interface StrategyDetail {
  strategy_id: string
  strategy_code: string
  strategy_name: string
  strategy_type: string
  description: string
  definition: Record<string, unknown>
  definition_version: number
  data_period: string
  secu_code: string
  template_id?: string | null
  allow_backtest: string
  status: string
  created_by: string
}

const run = ref<RunDetail | null>(null)
const strategyDetail = ref<StrategyDetail | null>(null)
const report = ref<Report | null>(null)
const equity = ref<EquityPoint[]>([])
const trades = ref<TradeRow[]>([])
const tradeTotal = ref(0)
const tradePage = ref(1)
const tradePageSize = ref(20)
const cashflows = ref<CashflowRow[]>([])
const positionLogs = ref<PositionLogRow[]>([])
const eventTraces = ref<EventTraceRow[]>([])
const traceTab = ref('trades')
const loading = ref(false)
const error = ref('')

const charts: { chart: echarts.ECharts; key: string }[] = []
const chartEls: Record<string, HTMLDivElement | null> = {
  equity: null,
  roi: null,
  profit: null,
  position: null,
}
// 图表 DOM 绑定（Vue 函数 ref：:ref="el => setChartRef('equity', el)"）
function setChartRef(key: string, el: unknown) {
  chartEls[key] = el as HTMLDivElement | null
}
// 全屏放大
const fullscreenKey = ref('')
const fullscreenDialog = ref(false)
const fullscreenChartEl = ref<HTMLDivElement | null>(null)

function fmtDateTimeLocal(d: number | undefined, t: number | undefined): string {
  if (!d) return '-'
  return `${fmtDate(d)} ${fmtTime(t)}`
}

function flowTypeName(t: string): string {
  return {
    INITIAL_DEPOSIT: '初始资金', BUY_PAY: '买入付款', SELL_RECEIVE: '卖出收款',
    FEE: '手续费', MARGIN_HOLD: '保证金占用', MARGIN_RELEASE: '保证金释放',
  }[t] ?? t
}

function flowTypeColor(t: string): string {
  return { INITIAL_DEPOSIT: 'blue', BUY_PAY: 'red', SELL_RECEIVE: 'green', FEE: 'orange', MARGIN_HOLD: 'purple', MARGIN_RELEASE: 'teal' }[t] ?? 'grey'
}

function posActionName(a: string): string {
  return { OPEN: '开仓', ADD: '加仓', REDUCE: '减仓', CLOSE: '平仓' }[a] ?? a
}

function posActionColor(a: string): string {
  return { OPEN: 'green', ADD: 'teal', REDUCE: 'orange', CLOSE: 'red' }[a] ?? 'grey'
}

function execStatusName(s: string): string {
  return { FILLED: '成交', REJECTED: '拒绝', PENDING: '挂单', EXPIRED: '过期' }[s] ?? s
}

function execStatusColor(s: string): string {
  return { FILLED: 'green', REJECTED: 'red', PENDING: 'blue', EXPIRED: 'orange' }[s] ?? 'grey'
}

// 未成交原因 → 可读说明 / 解决建议（让用户看得懂“为何无法交易”）
function rejectReasonHint(reason: string): string {
  if (reason.includes('不在环境交易时段')) return '该 bar 时间点不在环境配置的交易时段内；可在「环境与模板管理」中调整 trading_sessions（当前 GCMain 已配置为全天时段）'
  if (reason.includes('已达目标持仓上限') || reason.includes('已达最大持仓')) return '当前持仓已达到策略目标仓位上限，买入信号不再追加；如需继续加仓请提高 target_position_pct / max_position_pct'
  if (reason.includes('无持仓')) return '当前处于空仓状态，卖出信号无法执行；需先有买入成交形成持仓'
  if (reason.includes('今日已达成交笔数上限')) return '当日成交笔数已达上限（max_trades_per_day）；可调高该值或改用周/月窗口限制'
  if (reason.includes('近7日') || reason.includes('近30日')) return '滚动窗口内交易频率已达上限；可调高 max_trades_per_week / max_trades_per_month 或手续费上限'
  if (reason.includes('涨停')) return '买入价格触及涨停价，无法成交；通常次日恢复'
  if (reason.includes('跌停')) return '卖出价格触及跌停价，无法成交；通常次日恢复'
  if (reason.includes('T+') && reason.includes('交收')) return 'T+N 交收限制：当日买入的持仓在交收期内不可卖出（环境 trading_rules.t_plus）'
  if (reason.includes('资金不足')) return '可用资金不足以覆盖委托金额（含手续费）；可降低单笔投入或调高初始资金'
  if (reason.includes('最小交易间隔')) return '相邻两次成交间隔未达到策略设置的 min_interval_bars'
  if (reason.includes('允许交易时间点')) return '当前时间点不在策略规则 allowed_times 限定范围内'
  if (reason.includes('回测结束')) return '回测区间结束时仍有未成交委托，后续无 K 线可撮合'
  if (reason.includes('分批建仓最小间隔')) return '相邻两批建仓间隔未达到 builder.tranche_interval_bars 设定值'
  return ''
}

// ---------------------------------------------------------------------
// 回测背景信息：可读化函数与计算属性
// ---------------------------------------------------------------------

interface KV {
  label: string
  value: string
}

/** 数量方式中文名 */
function quantityTypeName(t: string | undefined): string {
  return { ALL_IN: '全部可用资金', ALL: '全部清仓', FIXED: '固定数量', PERCENT: '可用资金百分比', AMOUNT: '固定金额' }[t ?? ''] ?? (t ?? '-')
}

/** 保证金模式中文名 */
function marginModeName(m: string | undefined): string {
  return { FULL: '全额保证金', FUTURES: '期货保证金' }[m ?? ''] ?? (m ?? '-')
}

/** 环境类型中文名 */
function envTypeName(t: string | undefined): string {
  return { BACKTEST: '回测', PAPER: '模拟盘', SIMULATION: '仿真', LIVE: '实盘' }[t ?? ''] ?? (t ?? '-')
}

/** 撮合模式中文名 */
function fillModeName(f: string | undefined): string {
  return { NEXT_BAR_OPEN: '下一Bar开盘价成交（无未来函数）', CURRENT_CLOSE: '当前Bar收盘价成交（近似）' }[f ?? ''] ?? (f ?? '-')
}

/** 数据周期中文名 */
function periodName(p: string | undefined): string {
  return { Min: '分钟', Hour: '小时', Day: '日线' }[p ?? ''] ?? (p ?? '-')
}

/** 策略类型中文名 */
function strategyTypeName(t: string | undefined): string {
  return { RULE_BASED: '规则驱动' }[t ?? ''] ?? (t ?? '-')
}

/** hhmmss 字符串 → HH:mm:ss（如 '093000' → '09:30:00'） */
function hhmmssFmt(s: string | undefined): string {
  if (!s) return '-'
  const t = s.padStart(6, '0')
  return `${t.slice(0, 2)}:${t.slice(2, 4)}:${t.slice(4, 6)}`
}

/** 单段交易时段 → '09:30-15:00' */
function sessionFmt(start: string, end: string): string {
  return `${(start || '').slice(0, 2)}:${(start || '').slice(2, 4)}-${(end || '').slice(0, 2)}:${(end || '').slice(2, 4)}`
}

/** 交易时段数组 → '09:30-15:00 / 21:00-02:30'；空=不限制 */
function sessionsFmt(sessions: { start: string; end: string }[] | undefined): string {
  if (!sessions?.length) return '不限制（全天）'
  return sessions.map((s) => sessionFmt(s.start, s.end)).join(' / ')
}

/** 比值 → 百分比（commission_rate/slippage/margin_rate，0.0003 → '0.03%'） */
function pctRatio(r: number | undefined): string {
  if (r === undefined || r === null || Number.isNaN(r)) return '-'
  return `${(r * 100).toFixed(2)}%`
}

/** 频次限制 → '3 次'；0 或空 → '不限制' */
function limitText(n: number | undefined): string {
  if (n === undefined || n === null || n === 0) return '不限制'
  return `${fmtNum(n, 0)} 次`
}

/** 数量方式完整描述（含数量值） */
function quantityText(r: RuleDef | undefined): string {
  const t = r?.quantity_type
  if (!t) return '-'
  switch (t) {
    case 'FIXED': return `${quantityTypeName(t)} ${fmtNum(r.quantity, 0)} 手`
    case 'PERCENT': return `${quantityTypeName(t)} ${fmtNum(r.quantity)}%`
    case 'AMOUNT': return `${quantityTypeName(t)} ${fmtNum(r.quantity)}`
    default: return quantityTypeName(t)
  }
}

/** 窗口手续费上限 → '55000（30 日窗口）'；0 或空 → '不限制' */
function feeWindowText(v: number | undefined, days: number | undefined): string {
  if (v === undefined || v === null || v === 0) return '不限制'
  return `${fmtNum(v)}（${days || 30} 日窗口）`
}

/** 单个指标概览：'MA(ma_fast, window=5, field=close)' */
function indText(ind: IndicatorDef): string {
  const p = ind.params ?? {}
  const pText = Object.entries(p).map(([k, v]) => `${k}=${v}`).join(', ')
  return `${ind.type ?? '?'}${ind.id ? `(${ind.id}${pText ? ', ' + pText : ''})` : pText ? `(${pText})` : ''}`
}

/** 买卖信号表达式一行 */
function signalText(signals: { buy?: string; sell?: string } | undefined): string {
  if (!signals) return '（未设置）'
  const parts: string[] = []
  if (signals.buy) parts.push(`买入: ${signals.buy}`)
  if (signals.sell) parts.push(`卖出: ${signals.sell}`)
  return parts.length ? parts.join(' | ') : '（未设置）'
}

/** ① 数据配置一行：'Min / close / 预热30bar / 下一Bar开盘价成交' */
function dataConfigText(d: StrategySnapshot['data'] | undefined): string {
  if (!d) return '-'
  return [
    d.period ? periodName(d.period) : '',
    d.price_field ? `字段 ${d.price_field}` : '',
    d.warmup_bars ? `预热 ${d.warmup_bars} bar` : '',
    d.fill_mode ? fillModeName(d.fill_mode) : '',
  ].filter(Boolean).join(' / ') || '-'
}

// ---- 计算属性（模板薄化） ----

const options = computed<RunOptions>(() => run.value?.options ?? {})
const env = computed<EnvironmentSnapshot | null>(() => run.value?.environment_snapshot ?? null)
const snapRules = computed<{ buy?: RuleDef; sell?: RuleDef } | undefined>(() => run.value?.strategy_snapshot?.rules)
const indicators = computed<IndicatorDef[]>(() => run.value?.strategy_snapshot?.indicators ?? [])
const hasStrategyDef = computed<boolean>(() => !!run.value?.strategy_snapshot && Object.keys(run.value.strategy_snapshot).length > 0)

// 策略描述：快照 description 为主，空则回退当前策略说明
const primaryStrategyDesc = computed<string>(() => {
  const snap = (run.value?.strategy_snapshot?.description ?? '').trim()
  if (snap) return snap
  return (strategyDetail.value?.description ?? '').trim() || '（未填写策略说明）'
})

// 当前说明与快照不一致时提示
const strategyDescNote = computed<{ label: string; text: string } | null>(() => {
  const snap = (run.value?.strategy_snapshot?.description ?? '').trim()
  const cur = (strategyDetail.value?.description ?? '').trim()
  if (!snap || !cur || snap === cur) return null
  return { label: '策略当前说明与回测快照不同', text: cur }
})

const accountText = computed<string>(() => {
  const a = run.value?.account_snapshot
  if (!a?.account_name) return '-'
  return `${a.account_name}${a.account_code ? `（${a.account_code}）` : ''}`
})

// 成本生效值（覆盖链：环境 > 任务 options > 策略 definition.cost > 账户）
const effectiveCommission = computed<number | undefined>(() => {
  const envCost = env.value?.config?.cost?.commission_rate
  if (envCost) return envCost
  if (options.value.commission_rate != null) return options.value.commission_rate
  const stratCost = run.value?.strategy_snapshot?.cost?.commission_rate
  if (stratCost) return stratCost
  return run.value?.account_snapshot?.commission_rate || undefined
})
const effectiveSlippage = computed<number | undefined>(() => {
  const envSlip = env.value?.config?.cost?.slippage_pct
  if (envSlip) return envSlip
  if (options.value.slippage_pct != null) return options.value.slippage_pct
  const stratSlip = run.value?.strategy_snapshot?.cost?.slippage_pct
  if (stratSlip) return stratSlip
  return run.value?.account_snapshot?.slippage_pct || undefined
})
const effectiveCapital = computed<number | undefined>(() => {
  if (options.value.initial_capital != null) return options.value.initial_capital
  return run.value?.account_snapshot?.initial_capital || undefined
})
const commissionOverridden = computed<boolean>(() => options.value.commission_rate != null)

const envRules = computed<KV[]>(() => envRulesKV(env.value?.config?.trading_rules))

// ---- 键值行生成（模板与 CSV 共用） ----

/** 买入/卖出规则块 → KV 行 */
function ruleKV(r: RuleDef | undefined): KV[] {
  const opts = options.value
  const rows: KV[] = []
  rows.push({ label: '方向开关', value: r?.allow === false ? '禁用' : '启用' })
  rows.push({ label: '数量方式', value: quantityText(r) })
  rows.push({ label: '每日最大触发', value: limitText(r?.max_per_day) })
  rows.push({ label: '整轮回测最大触发', value: limitText(r?.max_per_run) })
  // 限定交易时间点：任务级覆盖优先（引擎对 buy/sell 合并生效）
  const allowed = opts.allowed_times?.length ? opts.allowed_times : r?.allowed_times
  const allowedText = allowed?.length ? allowed.map(hhmmssFmt).join('、') : '不限制'
  rows.push({ label: '限定交易时间点', value: allowedText + (opts.allowed_times?.length ? '（任务覆盖）' : '') })
  rows.push({ label: '近7日最大触发', value: limitText(r?.max_per_week) })
  rows.push({ label: '近30日最大触发', value: limitText(r?.max_per_month) })
  rows.push({ label: '窗口手续费上限', value: feeWindowText(r?.max_fee_per_window, r?.fee_window_days) })
  return rows
}

/** 风控块 → KV 行 */
function riskKV(risk: RiskDef | undefined): KV[] {
  const opts = options.value
  const rows: KV[] = []
  rows.push({ label: '止损', value: risk?.stop_loss_pct ? fmtPct(risk.stop_loss_pct) : '未设置' })
  rows.push({ label: '止盈', value: risk?.take_profit_pct ? fmtPct(risk.take_profit_pct) : '未设置' })
  rows.push({ label: '单标的仓位上限', value: risk?.max_position_pct ? fmtPct(risk.max_position_pct) : '100%' })
  rows.push({ label: '最大持仓数', value: risk?.max_positions ? `${fmtNum(risk.max_positions, 0)} 个` : '不限制' })
  const maxTrades = opts.max_trades_per_day != null ? opts.max_trades_per_day : risk?.max_trades_per_day
  rows.push({ label: '每日最大成交笔数', value: limitText(maxTrades) + (opts.max_trades_per_day != null ? '（任务覆盖）' : '') })
  rows.push({ label: '最小交易间隔', value: risk?.min_interval_bars ? `${fmtNum(risk.min_interval_bars, 0)} bar` : '不限制' })
  const b = risk?.builder
  rows.push({
    label: '分批建仓',
    value: b?.enabled
      ? `启用：目标仓位 ${fmtPct(b.target_position_pct ?? 0)}，分 ${b.tranches ?? 1} 批，间隔 ${b.tranche_interval_bars ?? 0} bar`
      : '不启用',
  })
  rows.push({ label: '分批减仓份数', value: risk?.reduce_tranches ? `${fmtNum(risk.reduce_tranches, 0)} 份` : '1 份（一次性清仓）' })
  rows.push({ label: '近7日成交上限', value: limitText(risk?.max_trades_per_week) })
  rows.push({ label: '近30日成交上限', value: limitText(risk?.max_trades_per_month) })
  rows.push({ label: '窗口手续费上限', value: feeWindowText(risk?.max_fee_per_window, risk?.fee_window_days) })
  return rows
}

/** 环境交易规则 → KV 行 */
function envRulesKV(rules: EnvironmentConfig['trading_rules'] | undefined): KV[] {
  if (!rules) return []
  const rows: KV[] = []
  rows.push({ label: 'T+N 交收', value: rules.t_plus ? `T+${rules.t_plus}` : 'T+0' })
  rows.push({ label: '最小变动价位', value: rules.tick_size ? fmtNum(rules.tick_size) : '-' })
  rows.push({ label: '合约乘数', value: rules.contract_multiplier ? fmtNum(rules.contract_multiplier, 0) : '-' })
  rows.push({ label: '涨停幅度', value: rules.limit_up_pct ? fmtPct(rules.limit_up_pct) : '无限制' })
  rows.push({ label: '跌停幅度', value: rules.limit_down_pct ? fmtPct(rules.limit_down_pct) : '无限制' })
  return rows
}

/** 汇总全部背景信息为 KV 行（供 CSV 导出） */
function backgroundKV(): KV[] {
  const rows: KV[] = []
  const snap = run.value?.strategy_snapshot
  const acc = run.value?.account_snapshot
  const envSnap = env.value
  // ① 策略
  rows.push({ label: '策略', value: `${run.value?.strategy_name ?? ''}${run.value?.strategy_code ? `（${run.value.strategy_code}）` : ''}` })
  rows.push({ label: '策略类型', value: strategyTypeName(snap?.strategy_type) })
  rows.push({ label: '策略描述', value: primaryStrategyDesc.value })
  if (strategyDescNote.value?.text) rows.push({ label: '当前策略说明', value: strategyDescNote.value.text })
  rows.push({ label: '数据配置', value: dataConfigText(snap?.data) })
  rows.push({ label: '指标', value: snap?.indicators?.length ? snap.indicators.map(indText).join(' | ') : '未配置指标' })
  rows.push({ label: '买卖信号', value: signalText(snap?.signals) })
  // ② 标的与市场
  rows.push({ label: '标的代码', value: run.value?.secu_code ?? '-' })
  rows.push({ label: '市场代码', value: run.value?.market_code ? fmtNum(run.value.market_code, 0) : '-' })
  rows.push({ label: '数据周期', value: periodName(run.value?.period) })
  rows.push({ label: '报告精度', value: report.value?.report_precision ?? '-' })
  rows.push({ label: '回测区间', value: `${fmtDate(run.value?.start_date)} ~ ${fmtDate(run.value?.end_date)}` })
  rows.push({ label: 'K线数量', value: report.value?.bar_count ? `${fmtNum(report.value.bar_count, 0)} 根` : '-' })
  // ③ 账户
  rows.push({ label: '账户', value: accountText.value })
  rows.push({ label: '初始资金', value: fmtNum(effectiveCapital.value) + (options.value.initial_capital != null ? '（任务覆盖）' : '') })
  rows.push({ label: '手续费率', value: pctRatio(effectiveCommission.value) + (commissionOverridden.value ? '（任务覆盖）' : '') })
  rows.push({ label: '滑点', value: pctRatio(effectiveSlippage.value) })
  rows.push({ label: '保证金模式', value: marginModeName(acc?.margin_mode) + (acc?.margin_rate ? `（${pctRatio(acc.margin_rate)}）` : '') })
  rows.push({ label: '计价币种', value: acc?.currency_type ?? '-' })
  // ④ 环境
  if (envSnap) {
    rows.push({ label: '环境', value: `${envSnap.env_name}${envSnap.env_code ? `（${envSnap.env_code}）` : ''}` })
    rows.push({ label: '环境类型', value: envTypeName(envSnap.env_type) })
    rows.push({ label: '地区', value: envSnap.region || '-' })
    rows.push({ label: '交易时段', value: sessionsFmt(envSnap.config?.trading_sessions) })
    rows.push({ label: '撮合模式', value: fillModeName(envSnap.config?.fill_mode ?? snap?.data?.fill_mode) })
    rows.push({ label: '环境币种', value: envSnap.config?.currency ?? '-' })
    if (envSnap.description) rows.push({ label: '环境说明', value: envSnap.description })
  } else {
    rows.push({ label: '环境', value: '未启用环境（按策略/账户配置运行）' })
  }
  // ⑤ 交易规则与限制
  for (const dir of ['buy', 'sell'] as const) {
    const dn = dir === 'buy' ? '买入' : '卖出'
    for (const kv of ruleKV(snap?.rules?.[dir])) rows.push({ label: `${dn}·${kv.label}`, value: kv.value })
  }
  for (const kv of riskKV(snap?.risk)) rows.push({ label: `风控·${kv.label}`, value: kv.value })
  for (const kv of envRulesKV(envSnap?.config?.trading_rules)) rows.push({ label: `环境·${kv.label}`, value: kv.value })
  return rows
}

async function loadReport(runId: string) {
  loading.value = true
  error.value = ''
  disposeCharts()
  report.value = null
  strategyDetail.value = null
  equity.value = []
  trades.value = []
  cashflows.value = []
  positionLogs.value = []
  eventTraces.value = []
  try {
    // 先拿任务完整信息（含策略/账户/环境快照），再并行加载报告与明细
    const r = await apiGet<RunDetail>(
      `/Meta/Finv/Quant/Backtest/Run/Get?runId=${encodeURIComponent(runId)}`,
    )
    if (r.status !== 'SUCCEEDED') {
      error.value = `任务当前状态为 ${r.status ?? '未知'}，仅 SUCCEEDED 任务可查看报告`
      return
    }
    run.value = r
    const [rep, eq, tr, cf, pl, ev, st] = await Promise.all([
      apiGet<Report>(`/Meta/Finv/Quant/Backtest/Run/Report?runId=${runId}`),
      apiGet<{ list: EquityPoint[] }>(`/Meta/Finv/Quant/Backtest/Run/Equity?runId=${runId}&page=1&pageSize=5000`),
      apiGet<{ total: number; list: TradeRow[] }>(`/Meta/Finv/Quant/Backtest/Run/Trades?runId=${runId}&page=1&pageSize=${tradePageSize.value}`),
      apiGet<{ list: CashflowRow[] }>(`/Meta/Finv/Quant/Backtest/Run/Cashflows?runId=${runId}&page=1&pageSize=1000`),
      apiGet<{ list: PositionLogRow[] }>(`/Meta/Finv/Quant/Backtest/Run/PositionLogs?runId=${runId}&page=1&pageSize=1000`),
      apiGet<{ list: EventTraceRow[] }>(`/Meta/Finv/Quant/Backtest/Run/EventTraces?runId=${runId}&page=1&pageSize=1000`),
      // 策略当前说明（Strategy/Get，失败不阻塞报告）
      r.strategy_id
        ? apiGet<StrategyDetail>(`/Meta/Finv/Quant/Backtest/Strategy/Get?strategyId=${encodeURIComponent(r.strategy_id)}`).catch(() => null)
        : Promise.resolve(null),
    ])
    report.value = rep
    equity.value = eq.list ?? []
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
    cashflows.value = cf.list ?? []
    positionLogs.value = pl.list ?? []
    eventTraces.value = ev.list ?? []
    strategyDetail.value = st
    await nextTick()
    renderCharts()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function xLabels(): string[] {
  const precision = report.value?.report_precision ?? 'Day'
  return equity.value.map((p) => {
    if (precision === 'Min') {
      return `${fmtDate(p.date)} ${fmtTime(p.time)}`
    }
    if (precision === 'Hour') {
      const h = Math.floor((p.time || 0) / 10000)
      return `${fmtDate(p.date)} ${String(h).padStart(2, '0')}时`
    }
    return fmtDate(p.date)
  })
}

function mkChart(key: string, el: HTMLDivElement | null): echarts.ECharts | null {
  if (!el) return null
  const chart = echarts.init(el)
  charts.push({ chart, key })
  return chart
}

function renderCharts() {
  if (!equity.value.length) return
  const labels = xLabels()
  const base = (title: string, unit: string) => ({
    title: { text: title, left: 12, textStyle: { fontSize: 14, fontWeight: 600 } },
    tooltip: { trigger: 'axis' },
    grid: { left: 70, right: 20, top: 48, bottom: 28 },
    xAxis: { type: 'category', data: labels, axisLabel: { fontSize: 10 } },
    yAxis: { type: 'value', scale: true, name: unit },
  })

  // ① 账户余额曲线（总资产 + 现金）
  const c1 = mkChart('equity', chartEls.equity)
  c1?.setOption({
    ...base('① 账户余额（总资产=现金+持仓市值）', report.value?.report_precision ?? ''),
    legend: { data: ['总资产', '现金'], top: 4, right: 12 },
    series: [
      { name: '总资产', type: 'line', data: equity.value.map((p) => p.equity), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#1565c0' } },
      { name: '现金', type: 'line', data: equity.value.map((p) => p.cash), showSymbol: false, lineStyle: { width: 1, type: 'dashed' }, itemStyle: { color: '#90a4ae' } },
    ],
  })

  // ② 投资收益率曲线
  const c2 = mkChart('roi', chartEls.roi)
  c2?.setOption({
    ...base('② 投资收益率（%）', '%'),
    series: [
      { name: '收益率', type: 'line', data: equity.value.map((p) => p.roi), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#2e7d32' }, areaStyle: { opacity: 0.08 } },
    ],
  })

  // ③ 投资收益额曲线
  const c3 = mkChart('profit', chartEls.profit)
  c3?.setOption({
    ...base('③ 累计收益额', ''),
    series: [
      { name: '收益额', type: 'line', data: equity.value.map((p) => p.profit), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#ef6c00' }, areaStyle: { opacity: 0.08 } },
    ],
  })

  // ⑦ 持仓金额曲线
  const c4 = mkChart('position', chartEls.position)
  c4?.setOption({
    ...base('⑦ 账户持仓金额', ''),
    series: [
      { name: '持仓市值', type: 'line', data: equity.value.map((p) => p.position_value), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#6a1b9a' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

// 全屏放大：复用主图表配置渲染到 dialog 内
function openFullscreen(key: string) {
  fullscreenKey.value = key
  fullscreenDialog.value = true
  nextTick(() => {
    const el = fullscreenChartEl.value
    if (!el) return
    const src = charts.find((c) => c.key === key)?.chart
    const f = echarts.init(el)
    if (src) {
      f.setOption(JSON.parse(JSON.stringify(src.getOption())))
    }
    f.resize()
    // 记录到独立变量，关闭时 dispose
    fullscreenChart = f
  })
}

let fullscreenChart: echarts.ECharts | null = null

function closeFullscreen() {
  if (fullscreenChart) {
    fullscreenChart.dispose()
    fullscreenChart = null
  }
  fullscreenDialog.value = false
  fullscreenKey.value = ''
  // 重新触发主图表 resize（布局恢复后）
  setTimeout(() => charts.forEach((c) => c.chart.resize()), 50)
}

function disposeCharts() {
  charts.forEach((c) => c.chart.dispose())
  charts.length = 0
}

function resizeCharts() {
  charts.forEach((c) => c.chart.resize())
}

async function loadTradesPage() {
  if (!run.value) return
  try {
    const tr = await apiGet<{ total: number; list: TradeRow[] }>(
      `/Meta/Finv/Quant/Backtest/Run/Trades?runId=${run.value.run_id}&page=${tradePage.value}&pageSize=${tradePageSize.value}`,
    )
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
  } catch (e) {
    error.value = (e as Error).message
  }
}

// ---------------------------------------------------------------------
// 导出（需求：报告底部结构化数据导出 JSON / CSV）
// ---------------------------------------------------------------------
function downloadBlob(filename: string, content: string, mime: string) {
  const blob = new Blob(['﻿' + content], { type: mime + ';charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function exportJSON() {
  const payload = {
    run: run.value,
    strategy_detail: strategyDetail.value,
    report: report.value,
    equity_points: equity.value,
    trades: trades.value,
    cashflows: cashflows.value,
    position_logs: positionLogs.value,
    event_traces: eventTraces.value,
  }
  downloadBlob(`finvquant_report_${run.value?.run_no ?? 'run'}_${Date.now()}.json`, JSON.stringify(payload, null, 2), 'application/json')
}

function csvEscape(v: unknown): string {
  const s = String(v ?? '')
  if (s.includes(',') || s.includes('"') || s.includes('\n')) {
    return '"' + s.replaceAll('"', '""') + '"'
  }
  return s
}

function exportCSV() {
  const lines: string[] = []
  const hr = () => lines.push('')
  // 1) 任务与报告主指标（key,value）
  lines.push('== 任务与汇总指标 ==')
  lines.push('字段,值')
  if (run.value) {
    lines.push(csvEscape('任务号'), csvEscape(run.value.run_no))
    lines.push(csvEscape('任务ID'), csvEscape(run.value.run_id))
    lines.push(csvEscape('策略'), csvEscape(run.value.strategy_name))
  }
  if (report.value) {
    for (const [k, v] of Object.entries(report.value)) {
      if (k === 'event_stats' || k === 'trade_signal_detail') continue
      lines.push(`${csvEscape(k)},${csvEscape(v)}`)
    }
  }
  // 2) 回测背景信息（字段,值）
  hr()
  lines.push('== 回测背景 ==')
  lines.push('字段,值')
  for (const kv of backgroundKV()) {
    lines.push(`${csvEscape(kv.label)},${csvEscape(kv.value)}`)
  }
  // 3) 未成交原因分布
  if (report.value?.event_stats?.reject_reasons) {
    hr()
    lines.push('== 未成交原因分布 ==')
    lines.push('原因,计数,说明')
    for (const [reason, cnt] of Object.entries(report.value.event_stats.reject_reasons)) {
      lines.push(`${csvEscape(reason)},${cnt},${csvEscape(rejectReasonHint(reason))}`)
    }
  }
  // 4) 成交记录
  hr()
  lines.push('== 成交记录 ==')
  lines.push('序号,时间,方向,价格,数量,金额,手续费,盈亏,持仓后,信号')
  for (const t of trades.value) {
    lines.push([t.seq, `${fmtDate(t.date)} ${fmtTime(t.time)}`, t.action, t.price, t.qty, t.amount, t.fee, t.profit, t.position_after, t.signal].map(csvEscape).join(','))
  }
  // 5) 持仓变化
  hr()
  lines.push('== 持仓变化明细 ==')
  lines.push('序号,时间,动作,价格,数量,持仓前,持仓后,成本前,成本后,备注')
  for (const p of positionLogs.value) {
    lines.push([p.seq, `${fmtDate(p.date)} ${fmtTime(p.time)}`, p.action, p.price, p.qty, p.position_before, p.position_after, p.avg_cost_before, p.avg_cost_after, p.remark].map(csvEscape).join(','))
  }
  // 6) 事件追踪
  hr()
  lines.push('== 事件追踪 ==')
  lines.push('序号,方向,触发原因,触发时间,结果,未成交原因')
  for (const ev of eventTraces.value) {
    lines.push([ev.seq, ev.action, ev.trigger_reason, `${fmtDate(ev.trigger_date)} ${fmtTime(ev.trigger_time)}`, ev.exec_status, ev.reject_reason || ''].map(csvEscape).join(','))
  }
  downloadBlob(`finvquant_report_${run.value?.run_no ?? 'run'}_${Date.now()}.csv`, lines.join('\n'), 'text/csv')
}

onMounted(async () => {
  window.addEventListener('resize', resizeCharts)
  const runId = route.query.runId as string | undefined
  if (runId) {
    await loadReport(runId)
  } else {
    error.value = '缺少 runId 参数，请从「回测分析」任务列表选择任务查看报告'
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  disposeCharts()
  if (fullscreenChart) {
    fullscreenChart.dispose()
    fullscreenChart = null
  }
})
</script>

<template>
  <v-container fluid>
    <v-alert v-if="error" type="error" dismissible class="mb-3">
      {{ error }}
      <template v-if="route.query.runId">
        <v-btn size="small" variant="text" color="error" class="ml-2" @click="router.push('/Meta/Finv/Quant/Backtest/Analysis')">
          返回任务列表
        </v-btn>
      </template>
    </v-alert>

    <v-card :loading="loading" v-if="!report">
      <v-card-text class="text-center py-12 text-medium-emphasis">
        <v-icon icon="mdi-chart-line" size="56" class="mb-3" />
        <p class="text-h6">正在加载投资策略回测收益分析报告…</p>
      </v-card-text>
    </v-card>

    <template v-if="report">
      <!-- 回测背景信息（平铺展开，5 小节） -->
      <v-card class="mb-3">
        <v-card-title class="pb-0 d-flex align-center flex-wrap">
          <v-icon icon="mdi-clipboard-text-clock-outline" class="mr-2" color="primary" />
          回测背景信息
          <v-chip v-if="strategyDescNote" size="small" color="warning" variant="tonal" class="ml-2"
            prepend-icon="mdi-alert">{{ strategyDescNote.label }}</v-chip>
        </v-card-title>
        <v-card-text>
          <!-- ① 策略信息 -->
          <div class="text-subtitle-2 font-weight-bold mb-1">
            <v-icon icon="mdi-account-cog-outline" size="small" class="mr-1" />① 策略信息
          </div>
          <template v-if="hasStrategyDef">
            <div class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">策略</span>
              <span class="text-body-2">{{ run?.strategy_name }}{{ run?.strategy_code ? `（${run?.strategy_code}）` : '' }}</span>
            </div>
            <div class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">策略类型</span>
              <span class="text-body-2">{{ strategyTypeName(run?.strategy_snapshot?.strategy_type) }}</span>
            </div>
            <div class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">策略描述</span>
              <span class="text-body-2">{{ primaryStrategyDesc }}</span>
            </div>
            <div v-if="strategyDescNote" class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">当前策略说明</span>
              <span class="text-body-2">{{ strategyDescNote.text }}</span>
            </div>
            <div class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">数据配置</span>
              <span class="text-body-2">{{ dataConfigText(run?.strategy_snapshot?.data) }}</span>
            </div>
            <div v-if="indicators.length" class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">指标</span>
              <span class="text-body-2 d-flex flex-wrap ga-1">
                <v-chip v-for="ind in indicators" :key="ind.id ?? ind.type ?? indText(ind)" size="x-small" variant="outlined">
                  {{ indText(ind) }}
                </v-chip>
              </span>
            </div>
            <div class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">买卖信号</span>
              <span class="text-body-2">{{ signalText(run?.strategy_snapshot?.signals) }}</span>
            </div>
          </template>
          <v-alert v-else density="compact" type="info" variant="tonal">
            本次回测未记录策略定义快照（策略可能已删除或定义为空），仅展示任务与报告信息。
          </v-alert>

          <v-row class="mt-2">
            <!-- ② 标的与市场 -->
            <v-col cols="12">
              <div class="text-subtitle-2 font-weight-bold mb-1">
                <v-icon icon="mdi-chart-box-outline" size="small" class="mr-1" />② 标的与市场
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">标的代码</span>
                <span class="text-body-2">{{ run?.secu_code || '-' }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">市场代码</span>
                <span class="text-body-2">{{ run?.market_code ? fmtNum(run?.market_code, 0) : '-' }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">数据周期</span>
                <span class="text-body-2">{{ periodName(run?.period) }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">报告精度</span>
                <span class="text-body-2">{{ report.report_precision || '-' }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">回测区间</span>
                <span class="text-body-2">{{ fmtDate(run?.start_date) }} ~ {{ fmtDate(run?.end_date) }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">K线数量</span>
                <span class="text-body-2">{{ fmtNum(report.bar_count, 0) }} 根</span>
              </div>
            </v-col>

            <!-- ③ 账户配置 -->
            <v-col cols="12">
              <div class="text-subtitle-2 font-weight-bold mb-1">
                <v-icon icon="mdi-bank-outline" size="small" class="mr-1" />③ 账户配置
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">账户</span>
                <span class="text-body-2">{{ accountText }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">初始资金</span>
                <span class="text-body-2 d-flex align-center">
                  {{ fmtNum(effectiveCapital) }}
                  <v-chip v-if="options.initial_capital != null" size="x-small" color="primary" variant="tonal" class="ml-1">任务覆盖</v-chip>
                </span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">手续费率</span>
                <span class="text-body-2 d-flex align-center">
                  {{ pctRatio(effectiveCommission) }}
                  <v-chip v-if="commissionOverridden" size="x-small" color="primary" variant="tonal" class="ml-1">任务覆盖</v-chip>
                </span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">滑点</span>
                <span class="text-body-2">{{ pctRatio(effectiveSlippage) }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">保证金模式</span>
                <span class="text-body-2">{{ marginModeName(run?.account_snapshot?.margin_mode) }}{{ run?.account_snapshot?.margin_rate ? `（${pctRatio(run?.account_snapshot.margin_rate)}）` : '' }}</span>
              </div>
              <div class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">计价币种</span>
                <span class="text-body-2">{{ run?.account_snapshot?.currency_type || '-' }}</span>
              </div>
            </v-col>
          </v-row>

          <v-row>
            <!-- ④ 环境配置 -->
            <v-col cols="12">
              <div class="text-subtitle-2 font-weight-bold mb-1">
                <v-icon icon="mdi-server-outline" size="small" class="mr-1" />④ 环境配置
              </div>
              <template v-if="env">
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">环境</span>
                  <span class="text-body-2">{{ env.env_name }}{{ env.env_code ? `（${env.env_code}）` : '' }}</span>
                </div>
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">环境类型</span>
                  <span class="text-body-2">{{ envTypeName(env.env_type) }}</span>
                </div>
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">地区</span>
                  <span class="text-body-2">{{ env.region || '-' }}</span>
                </div>
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">交易时段</span>
                  <span class="text-body-2">{{ sessionsFmt(env.config?.trading_sessions) }}</span>
                </div>
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">撮合模式</span>
                  <span class="text-body-2">{{ fillModeName(env.config?.fill_mode ?? run?.strategy_snapshot?.data?.fill_mode) }}</span>
                </div>
                <div class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">币种</span>
                  <span class="text-body-2">{{ env.config?.currency || '-' }}</span>
                </div>
                <div v-if="env.description" class="d-flex align-center py-1">
                  <span class="text-caption text-medium-emphasis kv-label">环境说明</span>
                  <span class="text-body-2">{{ env.description }}</span>
                </div>
              </template>
              <v-alert v-else density="compact" type="info" variant="tonal">
                本次回测未记录环境快照，按策略/账户配置运行。
              </v-alert>
            </v-col>
          </v-row>

          <!-- ⑤ 交易规则与限制 -->
          <div class="text-subtitle-2 font-weight-bold mt-1 mb-1">
            <v-icon icon="mdi-list-status" size="small" class="mr-1" />⑤ 交易规则与限制
          </div>
          <template v-if="hasStrategyDef">
            <div class="text-caption font-weight-bold mb-1">买入规则</div>
            <div v-for="(kv, i) in ruleKV(snapRules?.buy)" :key="'buy-' + i" class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">{{ kv.label }}</span>
              <span class="text-body-2">{{ kv.value }}</span>
            </div>
            <div class="text-caption font-weight-bold mb-1 mt-2">卖出规则</div>
            <div v-for="(kv, i) in ruleKV(snapRules?.sell)" :key="'sell-' + i" class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">{{ kv.label }}</span>
              <span class="text-body-2">{{ kv.value }}</span>
            </div>
            <div class="text-caption font-weight-bold mb-1 mt-2">风控与频率限制</div>
            <div v-for="(kv, i) in riskKV(run?.strategy_snapshot?.risk)" :key="'risk-' + i" class="d-flex align-center py-1">
              <span class="text-caption text-medium-emphasis kv-label">{{ kv.label }}</span>
              <span class="text-body-2">{{ kv.value }}</span>
            </div>
            <div v-if="envRules.length" class="mt-2">
              <div class="text-caption font-weight-bold mb-1">环境交易规则（T+N / 涨跌停 / 精度）</div>
              <div v-for="(kv, i) in envRules" :key="'envrule-' + i" class="d-flex align-center py-1">
                <span class="text-caption text-medium-emphasis kv-label">{{ kv.label }}</span>
                <span class="text-body-2">{{ kv.value }}</span>
              </div>
            </div>
          </template>
          <v-alert v-else density="compact" type="info" variant="tonal">
            无策略定义快照，无法展示交易规则与限制。
          </v-alert>
        </v-card-text>
      </v-card>

      <!-- 返回 + 报告头 -->
      <v-card class="mb-3">
        <v-card-title class="pb-0 d-flex align-center flex-wrap">
          <v-btn size="small" variant="text" prepend-icon="mdi-arrow-left" class="mr-2"
            @click="router.push('/Meta/Finv/Quant/Backtest/Analysis')">任务列表</v-btn>
          <v-icon icon="mdi-file-chart" class="mr-2" color="primary" />
          投资策略回测收益分析报告
          <v-chip size="small" class="ml-2">{{ report.secu_code }}</v-chip>
          <v-chip size="small" class="ml-1">{{ report.period }} / {{ report.report_precision }}</v-chip>
          <v-chip size="small" class="ml-1">{{ fmtDate(report.start_date) }} ~ {{ fmtDate(report.end_date) }}</v-chip>
          <v-chip size="small" color="grey" class="ml-1">共 {{ fmtNum(report.bar_count, 0) }} 根K线</v-chip>
          <v-spacer />
          <v-btn size="small" color="primary" variant="tonal" prepend-icon="mdi-export" class="ml-2"
            @click="exportJSON">导出 JSON</v-btn>
          <v-btn size="small" color="teal" variant="tonal" prepend-icon="mdi-file-delimited" class="ml-2"
            @click="exportCSV">导出 CSV</v-btn>
        </v-card-title>
        <v-card-text>
          <!-- 核心指标 -->
          <v-row>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="blue">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">期末总资产</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.final_equity) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" :color="report.total_profit >= 0 ? 'green' : 'red'">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">总收益额</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.total_profit) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" :color="report.total_return_pct >= 0 ? 'green' : 'red'">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">⑥ 到期收益率</div>
                  <div class="text-h6 font-weight-bold">{{ fmtPct(report.total_return_pct) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="teal">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">年化收益率</div>
                  <div class="text-h6 font-weight-bold">{{ fmtPct(report.annual_return_pct) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="orange">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">最大回撤</div>
                  <div class="text-h6 font-weight-bold">{{ fmtPct(report.max_drawdown_pct) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="purple">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">夏普比率</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.sharpe_ratio) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="indigo">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">胜率</div>
                  <div class="text-h6 font-weight-bold">{{ fmtPct(report.win_rate_pct) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="brown">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">盈亏比</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.profit_factor) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="cyan">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">④ 最大投入</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.max_invested) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="light-blue">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">⑤ 平均投入</div>
                  <div class="text-h6 font-weight-bold">{{ fmtNum(report.avg_invested) }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="grey">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">持仓天数</div>
                  <div class="text-h6 font-weight-bold">{{ report.invested_days }}</div>
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="6" sm="4" md="3" lg="2">
              <v-card variant="tonal" color="blue-grey">
                <v-card-text class="pa-2 text-center">
                  <div class="text-caption">交易笔数</div>
                  <div class="text-h6 font-weight-bold">{{ report.trade_count }}</div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-row class="mt-1">
            <v-col cols="12" md="6">
              <v-table density="compact">
                <tbody>
                  <tr><td class="text-body-2">初始启动资金</td><td class="text-right">{{ fmtNum(report.initial_capital) }}</td></tr>
                  <tr><td class="text-body-2">手续费总额</td><td class="text-right">{{ fmtNum(report.total_fee) }}</td></tr>
                  <tr><td class="text-body-2">买入 / 卖出笔数</td><td class="text-right">{{ report.buy_count }} / {{ report.sell_count }}</td></tr>
                  <tr><td class="text-body-2">盈利 / 亏损平仓</td><td class="text-right">{{ report.win_count }} / {{ report.loss_count }}</td></tr>
                </tbody>
              </v-table>
            </v-col>
            <v-col cols="12" md="6">
              <v-table density="compact">
                <tbody>
                  <tr><td class="text-body-2">年化波动率</td><td class="text-right">{{ fmtPct(report.volatility_pct) }}</td></tr>
                  <tr><td class="text-body-2">最佳 / 最差单期收益</td><td class="text-right">{{ fmtPct(report.best_day_pct) }} / {{ fmtPct(report.worst_day_pct) }}</td></tr>
                  <tr><td class="text-body-2">盈利 / 亏损期数</td><td class="text-right">{{ report.profit_days }} / {{ report.loss_days }}</td></tr>
                  <tr><td class="text-body-2">报告生成时间</td><td class="text-right">{{ report.generated_at?.replace('T', ' ').slice(0, 19) }}</td></tr>
                </tbody>
              </v-table>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- 曲线图（全宽 + 放大/全屏） -->
      <v-row>
        <v-col cols="12">
          <v-card class="mb-3">
            <v-card-title class="text-subtitle-1 d-flex align-center">
              <span class="flex-grow-1">① 账户余额曲线</span>
              <v-btn size="small" variant="text" icon="mdi-fullscreen" title="放大/全屏" @click="openFullscreen('equity')" />
            </v-card-title>
            <v-card-text class="pt-0">
              <div :ref="el => setChartRef('equity', el)" style="height: 320px" />
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12">
          <v-card class="mb-3">
            <v-card-title class="text-subtitle-1 d-flex align-center">
              <span class="flex-grow-1">② 投资收益率曲线</span>
              <v-btn size="small" variant="text" icon="mdi-fullscreen" title="放大/全屏" @click="openFullscreen('roi')" />
            </v-card-title>
            <v-card-text class="pt-0">
              <div :ref="el => setChartRef('roi', el)" style="height: 320px" />
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12">
          <v-card class="mb-3">
            <v-card-title class="text-subtitle-1 d-flex align-center">
              <span class="flex-grow-1">③ 累计收益额曲线</span>
              <v-btn size="small" variant="text" icon="mdi-fullscreen" title="放大/全屏" @click="openFullscreen('profit')" />
            </v-card-title>
            <v-card-text class="pt-0">
              <div :ref="el => setChartRef('profit', el)" style="height: 320px" />
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12">
          <v-card class="mb-3">
            <v-card-title class="text-subtitle-1 d-flex align-center">
              <span class="flex-grow-1">⑦ 账户持仓金额曲线</span>
              <v-btn size="small" variant="text" icon="mdi-fullscreen" title="放大/全屏" @click="openFullscreen('position')" />
            </v-card-title>
            <v-card-text class="pt-0">
              <div :ref="el => setChartRef('position', el)" style="height: 320px" />
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- 信号归因 + ⑨ 链路追踪统计 -->
      <v-card class="mb-3">
        <v-card-title class="text-subtitle-1">
          <v-icon icon="mdi-sigma" class="mr-2" color="primary" />信号归因
        </v-card-title>
        <v-card-text v-if="report.trade_signal_detail && Object.keys(report.trade_signal_detail).length">
          <v-chip v-for="(cnt, sig) in report.trade_signal_detail" :key="sig" class="mr-2 mb-1">
            {{ sig }}：{{ cnt }} 笔
          </v-chip>
        </v-card-text>
      </v-card>

      <!-- ⑨ 链路追踪统计（事件触发/成交/拒绝/委托耗时） -->
      <v-card v-if="report.event_stats" class="mb-3">
        <v-card-title class="text-subtitle-1">
          <v-icon icon="mdi-routes" class="mr-2" color="primary" />⑨ 链路追踪统计（事件触发 → 委托 → 成交/拒绝）
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="blue"><v-card-text class="pa-2 text-center"><div class="text-caption">事件触发总数</div><div class="text-h6">{{ report.event_stats.trigger_count }}</div></v-card-text></v-card></v-col>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="green"><v-card-text class="pa-2 text-center"><div class="text-caption">成交事件</div><div class="text-h6">{{ report.event_stats.filled_count }}</div></v-card-text></v-card></v-col>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="red"><v-card-text class="pa-2 text-center"><div class="text-caption">拒绝事件</div><div class="text-h6">{{ report.event_stats.rejected_count }}</div></v-card-text></v-card></v-col>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="orange"><v-card-text class="pa-2 text-center"><div class="text-caption">过期事件</div><div class="text-h6">{{ report.event_stats.expired_count }}</div></v-card-text></v-card></v-col>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="teal"><v-card-text class="pa-2 text-center"><div class="text-caption">平均委托耗时</div><div class="text-h6">{{ fmtNum(report.event_stats.avg_latency_bars, 2) }} bar</div></v-card-text></v-card></v-col>
            <v-col cols="6" sm="3" lg="2"><v-card variant="tonal" color="teal"><v-card-text class="pa-2 text-center"><div class="text-caption">平均耗时（秒）</div><div class="text-h6">{{ fmtNum(report.event_stats.avg_latency_sec, 2) }} s</div></v-card-text></v-card></v-col>
          </v-row>

          <!-- 未成交原因分布：可读化（原因 + 计数 + 说明/建议） -->
          <v-row v-if="report.event_stats.reject_reasons && Object.keys(report.event_stats.reject_reasons).length" class="mt-2">
            <v-col cols="12">
              <div class="text-subtitle-2 text-medium-emphasis mb-1">未成交原因分布（含解决建议）：</div>
              <v-table density="compact">
                <thead>
                  <tr><th>未成交原因</th><th class="text-right" style="width: 110px">次数</th><th>说明 / 解决建议</th></tr>
                </thead>
                <tbody>
                  <tr v-for="(cnt, reason) in report.event_stats.reject_reasons" :key="reason">
                    <td class="text-error text-body-2">{{ reason }}</td>
                    <td class="text-right text-body-2 font-weight-bold">{{ cnt }}</td>
                    <td class="text-caption text-medium-emphasis">{{ rejectReasonHint(reason) || '-' }}</td>
                  </tr>
                </tbody>
              </v-table>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- ⑨ 链路追踪明细：成交记录 / 资金流水 / 持仓变化 / 事件追踪 -->
      <v-card>
        <v-tabs v-model="traceTab" color="primary">
          <v-tab value="trades"><v-icon icon="mdi-swap-horizontal" class="mr-1" />成交记录（{{ report.trade_count }}）</v-tab>
          <v-tab value="cashflows"><v-icon icon="mdi-cash-multiple" class="mr-1" />资金流水明细（{{ cashflows.length }}）</v-tab>
          <v-tab value="positionLogs"><v-icon icon="mdi-briefcase-variant-outline" class="mr-1" />持仓变化明细（{{ positionLogs.length }}）</v-tab>
          <v-tab value="eventTraces"><v-icon icon="mdi-routes" class="mr-1" />事件追踪（{{ eventTraces.length }}）</v-tab>
        </v-tabs>
        <v-window v-model="traceTab">
          <!-- 成交记录 -->
          <v-window-item value="trades">
            <v-data-table-server v-model:page="tradePage" v-model:items-per-page="tradePageSize"
              :headers="[
                { title: '时间', key: 'timeText', width: 170 },
                { title: '方向', key: 'action', width: 80 },
                { title: '价格', key: 'price', width: 100 },
                { title: '数量', key: 'qty', width: 100 },
                { title: '金额', key: 'amount', width: 120 },
                { title: '手续费', key: 'fee', width: 100 },
                { title: '盈亏', key: 'profit', width: 110 },
                { title: '持仓后', key: 'position_after', width: 100 },
                { title: '信号', key: 'signal' },
                { title: '备注', key: 'remark', width: 110 },
              ]" :items="trades" :items-length="tradeTotal" item-value="trade_id"
              @update:options="loadTradesPage">
              <template #item.timeText="{ item }">
                {{ fmtDate(item.date) }} {{ String(item.time).padStart(6, '0').slice(0, 2) }}:{{ String(item.time).padStart(6, '0').slice(2, 4) }}
              </template>
              <template #item.action="{ item }">
                <v-chip size="small" :color="item.action === 'BUY' ? 'red' : 'green'">
                  {{ item.action === 'BUY' ? '买入' : '卖出' }}
                </v-chip>
              </template>
              <template #item.profit="{ item }">
                <span :class="item.profit > 0 ? 'text-success' : item.profit < 0 ? 'text-error' : ''">
                  {{ fmtNum(item.profit) }}
                </span>
              </template>
              <template #item.remark="{ item }">
                <span class="text-caption">{{ item.remark || '-' }}</span>
              </template>
            </v-data-table-server>
          </v-window-item>

          <!-- 资金流水明细 -->
          <v-window-item value="cashflows">
            <v-table density="compact">
              <thead>
                <tr><th>序号</th><th>时间</th><th>类型</th><th>金额</th><th>变动前现金</th><th>变动后现金</th><th>关联成交</th><th>备注</th></tr>
              </thead>
              <tbody>
                <tr v-for="cf in cashflows" :key="cf.cashflow_id">
                  <td>{{ cf.seq }}</td>
                  <td>{{ fmtDate(cf.date) }} {{ String(cf.time).padStart(6, '0').slice(0, 2) }}:{{ String(cf.time).padStart(6, '0').slice(2, 4) }}</td>
                  <td>
                    <v-chip size="x-small" :color="flowTypeColor(cf.flow_type)">{{ flowTypeName(cf.flow_type) }}</v-chip>
                  </td>
                  <td :class="cf.amount < 0 ? 'text-error' : 'text-success'">{{ fmtNum(cf.amount) }}</td>
                  <td>{{ fmtNum(cf.cash_before) }}</td>
                  <td>{{ fmtNum(cf.cash_after) }}</td>
                  <td>{{ cf.trade_id ? '#' + cf.trade_id : '-' }}</td>
                  <td class="text-caption">{{ cf.remark }}</td>
                </tr>
                <tr v-if="!cashflows.length"><td colspan="8" class="text-center text-medium-emphasis py-4">暂无资金流水</td></tr>
              </tbody>
            </v-table>
          </v-window-item>

          <!-- 持仓变化明细 -->
          <v-window-item value="positionLogs">
            <v-table density="compact">
              <thead>
                <tr><th>序号</th><th>时间</th><th>动作</th><th>价格</th><th>变动数量</th><th>持仓前</th><th>持仓后</th><th>成本前</th><th>成本后</th><th>关联成交</th><th>备注</th></tr>
              </thead>
              <tbody>
                <tr v-for="pl in positionLogs" :key="pl.log_id">
                  <td>{{ pl.seq }}</td>
                  <td>{{ fmtDate(pl.date) }} {{ String(pl.time).padStart(6, '0').slice(0, 2) }}:{{ String(pl.time).padStart(6, '0').slice(2, 4) }}</td>
                  <td>
                    <v-chip size="x-small" :color="posActionColor(pl.action)">{{ posActionName(pl.action) }}</v-chip>
                  </td>
                  <td>{{ fmtNum(pl.price) }}</td>
                  <td :class="pl.qty < 0 ? 'text-error' : 'text-success'">{{ fmtNum(pl.qty, 4) }}</td>
                  <td>{{ fmtNum(pl.position_before, 4) }}</td>
                  <td>{{ fmtNum(pl.position_after, 4) }}</td>
                  <td>{{ fmtNum(pl.avg_cost_before) }}</td>
                  <td>{{ fmtNum(pl.avg_cost_after) }}</td>
                  <td>{{ pl.trade_id ? '#' + pl.trade_id : '-' }}</td>
                  <td class="text-caption">{{ pl.remark }}</td>
                </tr>
                <tr v-if="!positionLogs.length"><td colspan="11" class="text-center text-medium-emphasis py-4">暂无持仓变化</td></tr>
              </tbody>
            </v-table>
          </v-window-item>

          <!-- 事件追踪 -->
          <v-window-item value="eventTraces">
            <v-table density="compact">
              <thead>
                <tr><th>序号</th><th>方向</th><th>触发原因</th><th>触发时间</th><th>委托下单时间</th><th>结果</th><th>成交时间</th><th>委托耗时</th><th>存活时间</th><th>未成交原因</th><th>关联成交</th></tr>
              </thead>
              <tbody>
                <tr v-for="ev in eventTraces" :key="ev.event_id">
                  <td>{{ ev.seq }}</td>
                  <td>
                    <v-chip size="x-small" :color="ev.action === 'BUY' ? 'red' : 'green'">{{ ev.action === 'BUY' ? '买入' : '卖出' }}</v-chip>
                  </td>
                  <td>{{ ev.trigger_reason }}</td>
                  <td>{{ fmtDateTimeLocal(ev.trigger_date, ev.trigger_time) }}</td>
                  <td>{{ fmtDateTimeLocal(ev.order_date, ev.order_time) }}</td>
                  <td>
                    <v-chip size="x-small" :color="execStatusColor(ev.exec_status)">{{ execStatusName(ev.exec_status) }}</v-chip>
                  </td>
                  <td>{{ ev.exec_status === 'FILLED' ? fmtDateTimeLocal(ev.exec_date, ev.exec_time) : '-' }}</td>
                  <td>{{ ev.exec_status === 'FILLED' ? ev.latency_bars + ' bar / ' + ev.latency_sec + 's' : '-' }}</td>
                  <td>{{ ev.exec_status !== 'PENDING' ? ev.alive_sec + 's' : '-' }}</td>
                  <td class="text-error text-caption">{{ ev.reject_reason || '-' }}</td>
                  <td>{{ ev.trade_id ? '#' + ev.trade_id : '-' }}</td>
                </tr>
                <tr v-if="!eventTraces.length"><td colspan="11" class="text-center text-medium-emphasis py-4">暂无事件追踪记录</td></tr>
              </tbody>
            </v-table>
          </v-window-item>
        </v-window>
      </v-card>
    </template>

    <!-- 图表全屏放大 Dialog -->
    <v-dialog v-model="fullscreenDialog" fullscreen transition="fade-transition">
      <v-card>
        <v-toolbar color="primary" density="comfortable">
          <v-toolbar-title>图表放大：{{ fullscreenKey }}</v-toolbar-title>
          <v-spacer />
          <v-btn icon="mdi-close" @click="closeFullscreen" title="关闭" />
        </v-toolbar>
        <v-card-text class="pa-2">
          <div ref="fullscreenChartEl" style="height: calc(100vh - 72px); width: 100%" />
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<style scoped>
/* 回测背景信息卡片：label 固定宽度，value 紧跟左对齐 */
.kv-label {
  min-width: 96px;
  flex-shrink: 0;
  padding-right: 8px;
}
</style>
