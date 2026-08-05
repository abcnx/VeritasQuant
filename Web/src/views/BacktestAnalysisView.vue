<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { apiGet, apiPost } from '../api'
import { fmtDate, fmtNum, fmtPct, fmtTime, statusColor } from '../utils'

const route = useRoute()

interface RunRow {
  run_id: string
  run_no: number
  strategy_name: string
  account_name: string
  secu_code: string
  period: string
  report_precision: string
  start_date: number
  end_date: number
  status: string
  progress: number
  error_message: string
  report: Report | null
}

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

const runs = ref<RunRow[]>([])
const runTotal = ref(0)
const page = ref(1)
const pageSize = ref(10)
const statusFilter = ref('')
const secuFilter = ref('')
const keyword = ref('')
const loading = ref(false)
const error = ref('')
const cancelling = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const currentRun = ref<RunRow | null>(null)
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
const reportLoading = ref(false)

const charts: echarts.ECharts[] = []
const equityChartEl = ref<HTMLDivElement | null>(null)
const roiChartEl = ref<HTMLDivElement | null>(null)
const profitChartEl = ref<HTMLDivElement | null>(null)
const positionChartEl = ref<HTMLDivElement | null>(null)

function fmtDateTimeLocal(d: number | undefined, t: number | undefined): string {
  if (!d) return '-'
  return `${fmtDate(d)} ${fmtTime(t)}`
}

// ⑨ 链路追踪展示辅助
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

async function loadRuns() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      page_size: String(pageSize.value),
    })
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (secuFilter.value) params.set('secu_code', secuFilter.value)
    if (keyword.value) params.set('keyword', keyword.value)
    const data = await apiGet<{ total: number; list: RunRow[] }>(`/Meta/FinvQuant/Backtest/Run/List?${params.toString()}`)
    runs.value = data.list ?? []
    runTotal.value = data.total ?? 0
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
  syncPolling()
}

// 轮询：存在执行中/待执行任务时每 5s 自动刷新列表与进度（评审：原实现仅手动刷新）
function syncPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (runs.value.some((r) => r.status === 'PENDING' || r.status === 'RUNNING')) {
    pollTimer = setInterval(async () => {
      try {
        const params = new URLSearchParams({ page: String(page.value), page_size: String(pageSize.value) })
        if (statusFilter.value) params.set('status', statusFilter.value)
        if (secuFilter.value) params.set('secu_code', secuFilter.value)
        if (keyword.value) params.set('keyword', keyword.value)
        const data = await apiGet<{ total: number; list: RunRow[] }>(`/Meta/FinvQuant/Backtest/Run/List?${params.toString()}`)
        runs.value = data.list ?? []
        runTotal.value = data.total ?? 0
        if (!runs.value.some((r) => r.status === 'PENDING' || r.status === 'RUNNING')) syncPolling()
      } catch {
        // 轮询失败静默，等待下一次
      }
    }, 5000)
  }
}

async function cancelRun(run: RunRow) {
  if (!confirm(`确认取消回测任务 #${run.run_no}？`)) return
  cancelling.value = run.run_id
  try {
    await apiPost('/Meta/FinvQuant/Backtest/Run/Cancel', { run_id: run.run_id })
    await loadRuns()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    cancelling.value = ''
  }
}

async function openRun(run: RunRow) {
  if (run.status !== 'SUCCEEDED') return
  currentRun.value = run
  reportLoading.value = true
  error.value = ''
  disposeCharts()
  report.value = null
  equity.value = []
  trades.value = []
  cashflows.value = []
  positionLogs.value = []
  eventTraces.value = []
  try {
    const [rep, eq, tr, cf, pl, ev] = await Promise.all([
      apiGet<Report>(`/Meta/FinvQuant/Backtest/Run/Report?run_id=${run.run_id}`),
      apiGet<{ list: EquityPoint[] }>(`/Meta/FinvQuant/Backtest/Run/Equity?run_id=${run.run_id}&page=1&page_size=5000`),
      apiGet<{ total: number; list: TradeRow[] }>(`/Meta/FinvQuant/Backtest/Run/Trades?run_id=${run.run_id}&page=1&page_size=${tradePageSize.value}`),
      apiGet<{ list: CashflowRow[] }>(`/Meta/FinvQuant/Backtest/Run/Cashflows?run_id=${run.run_id}&page=1&page_size=1000`),
      apiGet<{ list: PositionLogRow[] }>(`/Meta/FinvQuant/Backtest/Run/PositionLogs?run_id=${run.run_id}&page=1&page_size=1000`),
      apiGet<{ list: EventTraceRow[] }>(`/Meta/FinvQuant/Backtest/Run/EventTraces?run_id=${run.run_id}&page=1&page_size=1000`),
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
    reportLoading.value = false
  }
}

function xLabels(): string[] {
  const precision = report.value?.report_precision ?? 'Day'
  return equity.value.map((p) => {
    if (precision === 'Min') {
      return `${fmtDate(p.date)} ${fmtTime(p.time)}`
    }
    if (precision === 'Hour') {
      // 小时精度显示 日期 + 小时（评审：原实现只显示日期）
      const h = Math.floor((p.time || 0) / 10000)
      return `${fmtDate(p.date)} ${String(h).padStart(2, '0')}时`
    }
    return fmtDate(p.date)
  })
}

function renderCharts() {
  if (!equity.value.length) return
  const labels = xLabels()
  const mk = (el: HTMLDivElement | null): echarts.ECharts | null => {
    if (!el) return null
    const chart = echarts.init(el)
    charts.push(chart)
    return chart
  }
  const base = (title: string, unit: string) => ({
    title: { text: title, left: 12, textStyle: { fontSize: 13, fontWeight: 600 } },
    tooltip: { trigger: 'axis' },
    grid: { left: 70, right: 20, top: 44, bottom: 28 },
    xAxis: { type: 'category', data: labels, axisLabel: { fontSize: 10 } },
    yAxis: { type: 'value', scale: true, name: unit },
  })

  // ① 账户余额曲线（总资产 + 现金）
  const c1 = mk(equityChartEl.value)
  c1?.setOption({
    ...base('① 账户余额（总资产=现金+持仓市值）', report.value?.report_precision ?? ''),
    legend: { data: ['总资产', '现金'], top: 4, right: 12 },
    series: [
      { name: '总资产', type: 'line', data: equity.value.map((p) => p.equity), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#1565c0' } },
      { name: '现金', type: 'line', data: equity.value.map((p) => p.cash), showSymbol: false, lineStyle: { width: 1, type: 'dashed' }, itemStyle: { color: '#90a4ae' } },
    ],
  })

  // ② 投资收益率曲线
  const c2 = mk(roiChartEl.value)
  c2?.setOption({
    ...base('② 投资收益率（%）', '%'),
    series: [
      { name: '收益率', type: 'line', data: equity.value.map((p) => p.roi), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#2e7d32' }, areaStyle: { opacity: 0.08 } },
    ],
  })

  // ③ 投资收益额曲线
  const c3 = mk(profitChartEl.value)
  c3?.setOption({
    ...base('③ 累计收益额', ''),
    series: [
      { name: '收益额', type: 'line', data: equity.value.map((p) => p.profit), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#ef6c00' }, areaStyle: { opacity: 0.08 } },
    ],
  })

  // ⑦ 持仓金额曲线
  const c4 = mk(positionChartEl.value)
  c4?.setOption({
    ...base('⑦ 账户持仓金额', ''),
    series: [
      { name: '持仓市值', type: 'line', data: equity.value.map((p) => p.position_value), showSymbol: false, lineStyle: { width: 2 }, itemStyle: { color: '#6a1b9a' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

function disposeCharts() {
  charts.forEach((c) => c.dispose())
  charts.length = 0
}

function resizeCharts() {
  charts.forEach((c) => c.resize())
}

async function loadTradesPage() {
  if (!currentRun.value) return
  try {
    const tr = await apiGet<{ total: number; list: TradeRow[] }>(
      `/Meta/FinvQuant/Backtest/Run/Trades?run_id=${currentRun.value.run_id}&page=${tradePage.value}&page_size=${tradePageSize.value}`,
    )
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
  } catch (e) {
    error.value = (e as Error).message
  }
}

// 成交表翻页只由 v-data-table-server 的 @update:options 触发一次（评审：原实现 watch 与
// update:options 双触发导致重复请求），此处不再挂 watch。

onMounted(async () => {
  await loadRuns()
  const q = route.query.run_id as string | undefined
  if (q) {
    // 深链：优先用 Run/Get 直接加载任务（评审：原实现只查第一页，深链经常失败）
    try {
      const target = await apiGet<RunRow>(`/Meta/FinvQuant/Backtest/Run/Get?run_id=${encodeURIComponent(q)}`)
      if (target?.status === 'SUCCEEDED') {
        await openRun(target)
      } else {
        error.value = `任务当前状态为 ${target?.status ?? '未知'}，仅 SUCCEEDED 任务可查看报告`
      }
    } catch (e) {
      const listTarget = runs.value.find((r) => r.run_id === q)
      if (listTarget) await openRun(listTarget)
      else error.value = '目标任务不在当前列表，请从列表中选择（可按状态/标的/关键字过滤）'
    }
  }
  window.addEventListener('resize', resizeCharts)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  disposeCharts()
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>

<template>
  <v-container fluid>
    <v-row>
      <!-- 任务列表 -->
      <v-col cols="12" md="4" lg="3">
        <v-card>
          <v-card-title class="d-flex align-center">
            <v-icon icon="mdi-chart-timeline-variant" class="mr-2" color="primary" />回测任务
            <v-spacer />
            <v-btn size="small" variant="tonal" prepend-icon="mdi-refresh" @click="loadRuns">刷新</v-btn>
          </v-card-title>
          <v-card-text class="pt-0">
            <v-select v-model="statusFilter" :items="[
              { title: '全部状态', value: '' },
              { title: '待执行', value: 'PENDING' },
              { title: '执行中', value: 'RUNNING' },
              { title: '成功', value: 'SUCCEEDED' },
              { title: '失败', value: 'FAILED' },
              { title: '已取消', value: 'CANCELLED' },
            ]" density="compact" hide-details class="mb-2" @update:model-value="loadRuns" />
            <v-text-field v-model="secuFilter" label="标的过滤" density="compact" hide-details class="mb-2"
              @keyup.enter="loadRuns" @blur="loadRuns" />
            <v-text-field v-model="keyword" label="关键字过滤（策略/账户/任务号）" density="compact" hide-details
              class="mb-2" clearable @keyup.enter="loadRuns" @blur="loadRuns" />
          </v-card-text>
          <v-list v-if="runs.length" max-height="70vh" class="overflow-y-auto">
            <v-list-item v-for="r in runs" :key="r.run_id" :active="currentRun?.run_id === r.run_id"
              @click="openRun(r)">
              <template #prepend>
                <v-avatar size="32" :color="statusColor(r.status)" variant="tonal">
                  <v-icon size="18">{{ r.status === 'SUCCEEDED' ? 'mdi-check' : r.status === 'FAILED' ? 'mdi-close' : 'mdi-progress-clock' }}</v-icon>
                </v-avatar>
              </template>
              <v-list-item-title class="text-body-2">
                #{{ r.run_no }} {{ r.strategy_name }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ r.secu_code }} · {{ r.period }} · {{ fmtDate(r.start_date) }}~{{ fmtDate(r.end_date) }}
              </v-list-item-subtitle>
              <template #append>
                <v-chip size="x-small" :color="statusColor(r.status)">{{ r.status }}</v-chip>
                <v-btn v-if="r.status === 'PENDING' || r.status === 'RUNNING'" size="x-small" variant="text"
                  color="warning" icon="mdi-stop-circle-outline" :loading="cancelling === r.run_id"
                  title="取消任务" @click.stop="cancelRun(r)" />
              </template>
            </v-list-item>
          </v-list>
          <v-pagination v-model="page" :length="Math.max(1, Math.ceil(runTotal / pageSize))" density="compact"
            class="mt-2" @update:model-value="loadRuns" />
        </v-card>
      </v-col>

      <!-- 报告区 -->
      <v-col cols="12" md="8" lg="9">
        <v-alert v-if="error" type="error" dismissible class="mb-3">{{ error }}</v-alert>

        <v-card v-if="!report" :loading="reportLoading">
          <v-card-text class="text-center py-12 text-medium-emphasis">
            <v-icon icon="mdi-chart-line" size="56" class="mb-3" />
            <p class="text-h6">选择左侧已成功（SUCCEEDED）的回测任务查看收益分析报告</p>
            <p class="text-body-2">报告包含：账户余额/收益率/收益额/持仓金额曲线，以及最大投入、平均投入、
              到期收益率、最大回撤、夏普比率、胜率、盈亏比等技术指标</p>
          </v-card-text>
        </v-card>

        <template v-else>
          <!-- 报告头 -->
          <v-card class="mb-3">
            <v-card-title class="pb-0">
              <v-icon icon="mdi-file-chart" class="mr-2" color="primary" />
              投资策略回测收益分析报告
              <v-chip size="small" class="ml-2">{{ report.secu_code }}</v-chip>
              <v-chip size="small" class="ml-1">{{ report.period }} / {{ report.report_precision }}</v-chip>
              <v-chip size="small" class="ml-1">{{ fmtDate(report.start_date) }} ~ {{ fmtDate(report.end_date) }}</v-chip>
              <v-chip size="small" color="grey" class="ml-1">共 {{ fmtNum(report.bar_count, 0) }} 根K线</v-chip>
            </v-card-title>
            <v-card-text>
              <v-row>
                <!-- 核心指标 -->
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

          <!-- 曲线图 -->
          <v-row>
            <v-col cols="12" lg="6"><v-card class="mb-3"><div ref="equityChartEl" style="height: 300px" /></v-card></v-col>
            <v-col cols="12" lg="6"><v-card class="mb-3"><div ref="roiChartEl" style="height: 300px" /></v-card></v-col>
            <v-col cols="12" lg="6"><v-card class="mb-3"><div ref="profitChartEl" style="height: 300px" /></v-card></v-col>
            <v-col cols="12" lg="6"><v-card class="mb-3"><div ref="positionChartEl" style="height: 300px" /></v-card></v-col>
          </v-row>

          <!-- 信号归因 + ⑨ 链路追踪明细 -->
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
              <v-row v-if="report.event_stats.reject_reasons && Object.keys(report.event_stats.reject_reasons).length" class="mt-2">
                <v-col cols="12">
                  <span class="text-caption text-medium-emphasis mr-2">未成交原因分布：</span>
                  <v-chip v-for="(cnt, reason) in report.event_stats.reject_reasons" :key="reason" size="small" color="error" variant="tonal" class="mr-2 mb-1">
                    {{ reason }}：{{ cnt }}
                  </v-chip>
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
      </v-col>
    </v-row>
  </v-container>
</template>
