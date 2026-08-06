<script setup lang="ts">
import * as echarts from 'echarts'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { apiGet } from '../../../../../api'

interface QuoteBar {
  date: number
  time: number
  open: string | null
  high: string | null
  low: string | null
  close: string | null
  volume: number | null
  turnover: string | null
  change: string | null
  change_pct: string | null
  remark: string | null
}

// 证券下拉选项（usc 为 key，security_name_cn 为展示值，来自 finv_security 字典）
interface SecurityOption {
  usc: string
  security_name_cn: string
}

// 富途牛牛风格配色：红涨绿跌 + 均线黄/蓝/紫/青
const COLOR_UP = '#ef232a'
const COLOR_DOWN = '#14b143'
const MA_COLORS = ['#f6c343', '#3b8ff7', '#c56cf0', '#2fb28a'] // MA5 / MA10 / MA20 / MA30
const BOLL_COLORS = ['#ff8c69', '#ffffff', '#7ec8e3'] // 上轨 / 中轨 / 下轨（布林带）

const PAGE_SIZE = 240 // 每页条数（约全天 4 小时交易时段的分钟线数）

const secuCode = ref('NVDA') // 证券代码（usc key），默认 NVDA
const secuName = ref('') // 证券名称（security_name_cn），选中字典项后回填，便于确认
const dateInput = ref<string | null>(null) // 日历选择的日期（yyyy-MM-dd）
const dateMenu = ref(false)
const period = ref('Min')
const days = ref(1) // 1 日 / 5 日
const page = ref(1)
const loading = ref(false)
const error = ref('')
const total = ref(0)
const chartEl = ref<HTMLDivElement | null>(null)

// 查询结果提示（snackbar 吐司，3 秒自动消失，替代常驻 alert）
const toast = ref<{ text: string; visible: boolean }>({ text: '', visible: false })
function showToast(text: string) {
  toast.value = { text, visible: true }
}

let chart: echarts.ECharts | null = null

// 本地时区当天（不使用 toISOString 避免 UTC 偏移跨日）
function localTodayISO(): string {
  const d = new Date()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}
const todayISO = localTodayISO()

// Vuetify 4 的 v-date-picker 选中后 model 值可能是 Date 对象（随 locale 显示为英文），
// 统一规范化为 yyyy-MM-dd 纯日期字符串（无时间），保证展示与传参格式稳定
function normalizeDate(v: unknown): string | null {
  if (v === null || v === undefined || v === '') return null
  if (typeof v === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(v)) return v
  const d = new Date(v as string)
  if (Number.isNaN(d.getTime())) return null
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

function onDatePicked(v: unknown) {
  dateInput.value = normalizeDate(v)
  dateMenu.value = false
}

// 提交给后端的交易日期（yyyymmdd 纯数字，8 位）
const date = computed(() => (dateInput.value ? dateInput.value.replace(/-/g, '') : ''))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

// 证券下拉选项（usc:security_name_cn 作为展示文本，value 为 usc）
const securityOptions = ref<SecurityOption[]>([])
const secuItems = computed(() =>
  securityOptions.value.map((o) => ({
    title: `${o.usc}:${o.security_name_cn}`,
    value: o.usc,
  })),
)

// 证券代码输入变化：若命中字典项则回填证券名称，否则清空
function onSecuCodeInput(val: unknown) {
  // v-combobox 可能返回字符串或对象 {title,value}，统一取字符串值对比
  const s = typeof val === 'string' ? val : (val as { value?: string } | null | undefined)?.value ?? ''
  const hit = securityOptions.value.find((o) => o.usc === s)
  secuName.value = hit ? hit.security_name_cn : ''
}

// 查询条件变化时回到第 1 页（不自动触发查询，由用户点击“查询”）
watch([secuCode, dateInput, days, period], () => {
  page.value = 1
})

onMounted(async () => {
  // 加载证券下拉字典（finv_security.usc + security_name_cn）
  try {
    const data = await apiGet<{ list: SecurityOption[] }>('/Meta/FinvQuant/Metadata/Security/Options')
    securityOptions.value = data.list ?? []
  } catch {
    // 字典加载失败不阻塞查询（仍可手动输入证券代码）
    securityOptions.value = []
  }
})

onBeforeUnmount(() => {
  chart?.dispose()
  chart = null
})

function formatTime(time: number): string {
  const s = String(time).padStart(6, '0')
  return `${s.slice(0, 2)}:${s.slice(2, 4)}:${s.slice(4, 6)}`
}

function formatAxisLabel(bar: QuoteBar): string {
  const hm = formatTime(bar.time).slice(0, 5)
  if (days.value > 1 && bar.date) {
    const d = String(bar.date)
    return `${d.slice(4, 6)}/${d.slice(6, 8)} ${hm}`
  }
  return hm
}

function toNum(v: string | null | undefined): number | null {
  if (v === null || v === undefined || v === '') return null
  const n = Number(v)
  return Number.isNaN(n) ? null : n
}

// 滑动窗口均线（窗口内存在缺失值则输出 null，避免断点被错误填充）
function computeMA(closes: (number | null)[], n: number): (number | null)[] {
  const out: (number | null)[] = []
  let sum = 0
  let valid = 0
  for (let i = 0; i < closes.length; i++) {
    const v = closes[i]
    if (v !== null) {
      sum += v
      valid++
    }
    if (i >= n) {
      const drop = closes[i - n]
      if (drop !== null) {
        sum -= drop
        valid--
      }
    }
    out.push(valid >= n ? sum / n : null)
  }
  return out
}

// 布林带（BOLL）：中轨 = MA(n)，上/下轨 = 中轨 ± k × 标准差（滑动窗口）
// 返回 [上轨, 中轨, 下轨] 三条序列；窗口内存在缺失值则输出 null
function computeBOLL(closes: (number | null)[], n: number, k: number): [(number | null)[], (number | null)[], (number | null)[]] {
  const upper: (number | null)[] = []
  const mid: (number | null)[] = []
  const lower: (number | null)[] = []
  for (let i = 0; i < closes.length; i++) {
    // 取窗口内有效值
    const start = Math.max(0, i - n + 1)
    const window: number[] = []
    for (let j = start; j <= i; j++) {
      const v = closes[j]
      if (v !== null) window.push(v)
    }
    if (window.length < n) {
      upper.push(null)
      mid.push(null)
      lower.push(null)
      continue
    }
    let sum = 0
    for (const v of window) sum += v
    const mean = sum / window.length
    let variance = 0
    for (const v of window) variance += (v - mean) * (v - mean)
    const std = Math.sqrt(variance / window.length)
    mid.push(mean)
    upper.push(mean + k * std)
    lower.push(mean - k * std)
  }
  return [upper, mid, lower]
}

async function query() {
  error.value = ''
  // v-combobox 可能返回字符串（自由输入）或对象（从字典下拉选中 {title,value}），
  // 统一归一化为纯字符串 usc 代码，避免 .trim() 对非字符串报错
  const secuRaw = secuCode.value
  const secuStr = typeof secuRaw === 'string' ? secuRaw : (secuRaw as { value?: string } | null | undefined)?.value ?? ''
  if (!secuStr.trim()) {
    error.value = '证券代码不能为空'
    return
  }
  if (!/^\d{8}$/.test(date.value)) {
    error.value = '请通过日历选择交易日期'
    return
  }

  loading.value = true
  try {
    const params = new URLSearchParams({
      secu_code: secuStr.trim(),
      date: date.value,
      period: period.value,
      days: String(days.value),
      page: String(page.value),
      page_size: String(PAGE_SIZE),
    })
    // 选中字典项时回传证券名称（security_name_cn），便于服务端与前端确认证券
    if (secuName.value.trim()) {
      params.set('secu_name', secuName.value.trim())
    }
    const response = await fetch(`/API/V1/Quote/Query?${params}`)
    const body = await response.json()
    if (body.code !== 0) {
      error.value = body.message || '查询失败'
      return
    }
    const data = body.data ?? {}
    const bars: QuoteBar[] = data.bars ?? []
    total.value = data.total ?? bars.length
    const rangeText =
      days.value > 1 && bars.length
        ? `${String(bars[0].date).slice(4)}-${String(bars[bars.length - 1].date).slice(4)}`
        : date.value
    const namePart = secuName.value.trim() ? `（${secuName.value.trim()}）` : ''
    showToast(`${data.secu_code ?? '?'}${namePart} ${rangeText} ${data.period ?? 'Min'} · ${days.value}日 · 共 ${total.value} 根 · 第 ${page.value}/${totalPages.value} 页`)
    await nextTick()
    if (bars.length) {
      renderChart(bars)
    } else {
      chart?.clear()
    }
  } catch {
    error.value = '网络错误：无法连接服务端'
  } finally {
    loading.value = false
  }
}

function renderChart(bars: QuoteBar[]) {
  if (!chartEl.value) return
  if (!chart) {
    chart = echarts.init(chartEl.value)
  }

  const labels = bars.map((b) => formatAxisLabel(b))
  const candleData = bars.map((b) => [
    toNum(b.open) ?? 0,
    toNum(b.close) ?? 0,
    toNum(b.low) ?? 0,
    toNum(b.high) ?? 0,
  ])
  const volumes = bars.map((b) => b.volume ?? 0)
  const closes = bars.map((b) => toNum(b.close))
  const maSeries = [5, 10, 20, 30].map((n, idx) => ({
    name: `MA${n}`,
    type: 'line' as const,
    data: computeMA(closes, n),
    smooth: true,
    showSymbol: false,
    connectNulls: false,
    lineStyle: { width: 1.1, color: MA_COLORS[idx] },
    emphasis: { disabled: true },
    z: 3,
  }))
  // 布林带三条线（20 周期，2 倍标准差）：上轨 / 中轨 / 下轨
  const [bollUp, bollMid, bollLow] = computeBOLL(closes, 20, 2)
  const bollSeries = [
    { name: 'BOLL上轨', type: 'line' as const, data: bollUp, smooth: true, showSymbol: false, connectNulls: false, lineStyle: { width: 1, color: BOLL_COLORS[0] }, emphasis: { disabled: true }, z: 2 },
    { name: 'BOLL中轨', type: 'line' as const, data: bollMid, smooth: true, showSymbol: false, connectNulls: false, lineStyle: { width: 1, color: BOLL_COLORS[1] }, emphasis: { disabled: true }, z: 2 },
    { name: 'BOLL下轨', type: 'line' as const, data: bollLow, smooth: true, showSymbol: false, connectNulls: false, lineStyle: { width: 1, color: BOLL_COLORS[2] }, emphasis: { disabled: true }, z: 2 },
  ]

  chart.setOption(
    {
      animation: false,
      backgroundColor: '#1e1e1e',
      axisPointer: { link: [{ xAxisIndex: 'all' }] },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross', label: { backgroundColor: '#3a3a3a' } },
        backgroundColor: '#2d2d2d',
        borderColor: '#444444',
        textStyle: { color: '#eee', fontSize: 12 },
        formatter: (params: unknown) => {
          const list = params as Array<{ dataIndex: number; seriesName: string; value: number | number[] | null }>
          if (!list?.length) return ''
          const index = list[0].dataIndex
          const bar = bars[index]
          if (!bar) return ''
          const price = (v: string | null) => (v !== null && v !== undefined ? Number(v).toFixed(4) : '—')
          const up = (toNum(bar.close) ?? 0) >= (toNum(bar.open) ?? 0)
          const changeText =
            bar.change && bar.change_pct
              ? `<span style="color:${up ? COLOR_UP : COLOR_DOWN}">${bar.change}（${Number(bar.change_pct).toFixed(2)}%）</span>`
              : '—'
          const rows: string[] = [
            `<b>${bar.date} ${formatTime(bar.time)}</b>`,
            `开盘 ${price(bar.open)}&nbsp;&nbsp;&nbsp;收盘 <span style="color:${up ? COLOR_UP : COLOR_DOWN}">${price(bar.close)}</span>`,
            `最高 ${price(bar.high)}&nbsp;&nbsp;&nbsp;最低 ${price(bar.low)}`,
            `涨跌 ${changeText}`,
          ]
          if (bar.volume !== null && bar.volume !== undefined) rows.push(`成交量 ${Number(bar.volume).toLocaleString()}`)
          if (bar.turnover) rows.push(`成交额 ${Number(bar.turnover).toLocaleString()}`)
          // 均线：每个指标单独一行展示
          const maVals = [5, 10, 20, 30]
            .map((n) => {
              const v = computeMA(closes, n)[index]
              return v === null ? null : `MA${n} ${v.toFixed(4)}`
            })
            .filter((x): x is string => x !== null)
          maVals.forEach((t) => rows.push(t))
          // 布林带三条线值：每个指标单独一行展示
          const bollVals = [
            { label: 'BOLL上轨', v: bollUp[index] },
            { label: 'BOLL中轨', v: bollMid[index] },
            { label: 'BOLL下轨', v: bollLow[index] },
          ]
            .map((x) => (x.v === null ? null : `${x.label} ${x.v.toFixed(4)}`))
            .filter((x): x is string => x !== null)
          bollVals.forEach((t) => rows.push(t))
          if (bar.remark) rows.push(`备注 ${bar.remark}`)
          return rows.join('<br/>')
        },
      },
      legend: [
        {
          // 第一行：K 线 + 均线
          top: 4,
          left: 8,
          itemWidth: 14,
          itemHeight: 8,
          icon: 'roundRect',
          textStyle: { fontSize: 11, color: '#ccc' },
          data: ['K线', 'MA5', 'MA10', 'MA20', 'MA30'],
        },
        {
          // 第二行：布林带（上/中/下轨），支持点击显隐
          top: 20,
          left: 8,
          itemWidth: 14,
          itemHeight: 8,
          icon: 'roundRect',
          textStyle: { fontSize: 11, color: '#ccc' },
          data: ['BOLL上轨', 'BOLL中轨', 'BOLL下轨'],
        },
      ],
      grid: [
        { left: 8, right: 56, top: 32, height: '58%' },
        { left: 8, right: 56, top: '72%', height: '17%' },
      ],
      xAxis: [
        {
          type: 'category',
          data: labels,
          boundaryGap: true,
          axisLine: { lineStyle: { color: '#3a3a3a' } },
          axisTick: { show: false },
          axisLabel: { color: '#aaa', fontSize: 11, hideOverlap: true },
        },
        {
          type: 'category',
          data: labels,
          gridIndex: 1,
          axisLine: { lineStyle: { color: '#3a3a3a' } },
          axisTick: { show: false },
          axisLabel: { show: false },
        },
      ],
      yAxis: [
        {
          scale: true,
          position: 'right',
          axisLabel: { color: '#aaa', fontSize: 11 },
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: { lineStyle: { color: '#333333' } },
        },
        {
          gridIndex: 1,
          position: 'right',
          axisLabel: { show: false },
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: { show: false },
        },
      ],
      dataZoom: [{ type: 'inside', xAxisIndex: [0, 1], start: 0, end: 100 }],
      series: [
        {
          name: 'K线',
          type: 'candlestick',
          data: candleData,
          itemStyle: {
            color: COLOR_UP,
            color0: COLOR_DOWN,
            borderColor: COLOR_UP,
            borderColor0: COLOR_DOWN,
          },
        },
        ...maSeries,
        ...bollSeries,
        {
          name: '成交量',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: volumes,
          barWidth: '60%',
          itemStyle: {
            color: (param: unknown) => {
              const p = param as { dataIndex: number }
              const bar = bars[p.dataIndex]
              const up = (toNum(bar?.close) ?? 0) >= (toNum(bar?.open) ?? 0)
              return up ? COLOR_UP : COLOR_DOWN
            },
          },
        },
      ],
    },
    true,
  )
}
</script>

<template>
  <v-card>
    <v-card-title>
      <v-icon icon="mdi-chart-line" class="mr-2" />
      历史行情查询
    </v-card-title>
    <v-card-subtitle>
      按证券代码 + 交易日查询分钟级 K 线；支持 1 日 / 5 日切换与分页翻页；红涨绿跌，悬停查看详情
    </v-card-subtitle>

    <v-card-text>
      <v-form @submit.prevent="query">
        <v-row align="center" dense>
          <v-col cols="12" sm="3">
            <!-- 证券代码：下拉字典（usc:security_name_cn）+ 手动输入，默认 NVDA -->
            <v-combobox
              v-model="secuCode"
              :items="secuItems"
              label="证券代码"
              placeholder="如: NVDA"
              density="compact"
              hide-details
              clearable
              @update:model-value="onSecuCodeInput"
            />
          </v-col>
          <v-col cols="12" sm="3">
            <v-menu v-model="dateMenu" :close-on-content-click="false" transition="scale-transition">
              <template #activator="{ props }">
                <v-text-field
                  v-model="dateInput"
                  label="交易日期"
                  placeholder="点击日历选择"
                  readonly
                  density="compact"
                  hide-details
                  prepend-inner-icon="mdi-calendar"
                  v-bind="props"
                />
              </template>
              <v-date-picker
                v-model="dateInput"
                :max="todayISO"
                show-adjacent-months
                @update:model-value="onDatePicked"
              />
            </v-menu>
          </v-col>
          <v-col cols="12" sm="2">
            <v-select
              v-model="period"
              label="周期"
              :items="[{ title: '1分（Min）', value: 'Min' }]"
              density="compact"
              hide-details
            />
          </v-col>
          <v-col cols="12" sm="2">
            <v-btn-toggle v-model="days" mandatory density="comfortable" variant="outlined" divided>
              <v-btn :value="1" size="small">1日</v-btn>
              <v-btn :value="5" size="small">5日</v-btn>
            </v-btn-toggle>
          </v-col>
          <v-col cols="12" sm="2" class="text-right">
            <v-btn type="submit" color="primary" :loading="loading" prepend-icon="mdi-magnify">
              查询
            </v-btn>
          </v-col>
        </v-row>
      </v-form>

      <v-alert v-if="error" type="error" class="mt-3" density="compact">{{ error }}</v-alert>
    </v-card-text>

    <!-- 图表占满卡片可用宽度，左右不留空隙 -->
    <div ref="chartEl" class="w-100" style="width: 100%; height: 560px" />

    <v-card-text class="pt-0">
      <v-row align="center" justify="space-between" dense>
        <v-col cols="auto" class="text-body-2 text-medium-emphasis">
          每页 {{ PAGE_SIZE }} 根 · 共 {{ total }} 根
        </v-col>
        <v-col cols="auto">
          <v-pagination
            v-model="page"
            :length="totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="query"
          />
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>

  <!-- 查询结果提示：3 秒自动消失 -->
  <v-snackbar v-model="toast.visible" :timeout="3000" color="info" location="top" rounded="lg">
    <v-icon icon="mdi-information-outline" class="mr-2" />
    {{ toast.text }}
    <template #actions>
      <v-btn variant="text" icon="mdi-close" @click="toast.visible = false" />
    </template>
  </v-snackbar>
</template>
