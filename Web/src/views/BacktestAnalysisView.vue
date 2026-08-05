<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { apiGet } from '../api'

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
  generated_at: string
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

const currentRun = ref<RunRow | null>(null)
const report = ref<Report | null>(null)
const equity = ref<EquityPoint[]>([])
const trades = ref<TradeRow[]>([])
const tradeTotal = ref(0)
const tradePage = ref(1)
const tradePageSize = ref(20)
const reportLoading = ref(false)

const charts: echarts.ECharts[] = []
const equityChartEl = ref<HTMLDivElement | null>(null)
const roiChartEl = ref<HTMLDivElement | null>(null)
const profitChartEl = ref<HTMLDivElement | null>(null)
const positionChartEl = ref<HTMLDivElement | null>(null)

function fmtDate(d: number): string {
  if (!d) return '-'
  const s = String(d)
  return `${s.slice(0, 4)}-${s.slice(4, 6)}-${s.slice(6, 8)}`
}

function fmtNum(v: number | undefined | null, digits = 2): string {
  if (v === undefined || v === null || Number.isNaN(v)) return '-'
  return Number(v).toLocaleString('en-US', { maximumFractionDigits: digits })
}

function fmtPct(v: number | undefined | null): string {
  if (v === undefined || v === null || Number.isNaN(v)) return '-'
  return `${Number(v).toFixed(2)}%`
}

function statusColor(s: string): string {
  return { PENDING: 'grey', RUNNING: 'primary', SUCCEEDED: 'success', FAILED: 'error', CANCELLED: 'warning' }[s] ?? 'grey'
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
    const data = await apiGet<{ total: number; list: RunRow[] }>(`/Backtest/Run/List?${params.toString()}`)
    runs.value = data.list ?? []
    runTotal.value = data.total ?? 0
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
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
  try {
    const [rep, eq, tr] = await Promise.all([
      apiGet<Report>(`/Backtest/Run/Report?run_id=${run.run_id}`),
      apiGet<{ list: EquityPoint[] }>(`/Backtest/Run/Equity?run_id=${run.run_id}&page=1&page_size=5000`),
      apiGet<{ total: number; list: TradeRow[] }>(`/Backtest/Run/Trades?run_id=${run.run_id}&page=1&page_size=${tradePageSize.value}`),
    ])
    report.value = rep
    equity.value = eq.list ?? []
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
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
      const t = String(p.time).padStart(6, '0')
      return `${fmtDate(p.date)} ${t.slice(0, 2)}:${t.slice(2, 4)}`
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
      `/Backtest/Run/Trades?run_id=${currentRun.value.run_id}&page=${tradePage.value}&page_size=${tradePageSize.value}`,
    )
    trades.value = tr.list ?? []
    tradeTotal.value = tr.total ?? 0
  } catch (e) {
    error.value = (e as Error).message
  }
}

watch(tradePage, loadTradesPage)
watch(tradePageSize, loadTradesPage)

onMounted(async () => {
  await loadRuns()
  const q = route.query.run_id as string | undefined
  if (q) {
    const target = runs.value.find((r) => r.run_id === q)
    if (target) await openRun(target)
    else {
      // 翻页查找（简化：仅当第一页找到；否则提示从列表选择）
      error.value = '目标任务不在当前列表，请从列表中选择（可按状态/标的筛选）'
    }
  }
  window.addEventListener('resize', resizeCharts)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  disposeCharts()
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

          <!-- 信号归因 + 成交记录 -->
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

          <v-card>
            <v-card-title class="text-subtitle-1">
              <v-icon icon="mdi-swap-horizontal" class="mr-2" color="primary" />成交记录
            </v-card-title>
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
            </v-data-table-server>
          </v-card>
        </template>
      </v-col>
    </v-row>
  </v-container>
</template>
