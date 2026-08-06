<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { apiGet } from '../../../api'
import { fmtDate, fmtNum, fmtPct, fmtTime } from '../../../utils'

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

const run = ref<{ run_id: string; run_no: number; strategy_name: string } | null>(null)
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

async function loadReport(runId: string) {
  loading.value = true
  error.value = ''
  disposeCharts()
  report.value = null
  equity.value = []
  trades.value = []
  cashflows.value = []
  positionLogs.value = []
  eventTraces.value = []
  try {
    // 先拿任务信息（run_no / strategy_name），再并行加载报告与明细
    const r = await apiGet<{ run_id: string; run_no: number; strategy_name: string; status: string }>(
      `/Meta/FinvQuant/Backtest/Run/Get?runId=${encodeURIComponent(runId)}`,
    )
    if (r.status !== 'SUCCEEDED') {
      error.value = `任务当前状态为 ${r.status ?? '未知'}，仅 SUCCEEDED 任务可查看报告`
      return
    }
    run.value = r
    const [rep, eq, tr, cf, pl, ev] = await Promise.all([
      apiGet<Report>(`/Meta/FinvQuant/Backtest/Run/Report?runId=${runId}`),
      apiGet<{ list: EquityPoint[] }>(`/Meta/FinvQuant/Backtest/Run/Equity?runId=${runId}&page=1&pageSize=5000`),
      apiGet<{ total: number; list: TradeRow[] }>(`/Meta/FinvQuant/Backtest/Run/Trades?runId=${runId}&page=1&pageSize=${tradePageSize.value}`),
      apiGet<{ list: CashflowRow[] }>(`/Meta/FinvQuant/Backtest/Run/Cashflows?runId=${runId}&page=1&pageSize=1000`),
      apiGet<{ list: PositionLogRow[] }>(`/Meta/FinvQuant/Backtest/Run/PositionLogs?runId=${runId}&page=1&pageSize=1000`),
      apiGet<{ list: EventTraceRow[] }>(`/Meta/FinvQuant/Backtest/Run/EventTraces?runId=${runId}&page=1&pageSize=1000`),
    ])
    report.value = rep
    equity.value = eq.list ?? []
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
    cashflows.value = cf.list ?? []
    positionLogs.value = pl.list ?? []
    eventTraces.value = ev.list ?? []
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
      `/Meta/FinvQuant/Backtest/Run/Trades?runId=${run.value.run_id}&page=${tradePage.value}&pageSize=${tradePageSize.value}`,
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
  // 2) 未成交原因分布
  if (report.value?.event_stats?.reject_reasons) {
    hr()
    lines.push('== 未成交原因分布 ==')
    lines.push('原因,计数,说明')
    for (const [reason, cnt] of Object.entries(report.value.event_stats.reject_reasons)) {
      lines.push(`${csvEscape(reason)},${cnt},${csvEscape(rejectReasonHint(reason))}`)
    }
  }
  // 3) 成交记录
  hr()
  lines.push('== 成交记录 ==')
  lines.push('序号,时间,方向,价格,数量,金额,手续费,盈亏,持仓后,信号')
  for (const t of trades.value) {
    lines.push([t.seq, `${fmtDate(t.date)} ${fmtTime(t.time)}`, t.action, t.price, t.qty, t.amount, t.fee, t.profit, t.position_after, t.signal].map(csvEscape).join(','))
  }
  // 4) 持仓变化
  hr()
  lines.push('== 持仓变化明细 ==')
  lines.push('序号,时间,动作,价格,数量,持仓前,持仓后,成本前,成本后,备注')
  for (const p of positionLogs.value) {
    lines.push([p.seq, `${fmtDate(p.date)} ${fmtTime(p.time)}`, p.action, p.price, p.qty, p.position_before, p.position_after, p.avg_cost_before, p.avg_cost_after, p.remark].map(csvEscape).join(','))
  }
  // 5) 事件追踪
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
        <v-btn size="small" variant="text" color="error" class="ml-2" @click="router.push('/Meta/FinvQuant/Backtest/Analysis')">
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
      <!-- 返回 + 报告头 -->
      <v-card class="mb-3">
        <v-card-title class="pb-0 d-flex align-center flex-wrap">
          <v-btn size="small" variant="text" prepend-icon="mdi-arrow-left" class="mr-2"
            @click="router.push('/Meta/FinvQuant/Backtest/Analysis')">任务列表</v-btn>
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
