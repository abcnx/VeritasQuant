<script setup lang="ts">
import * as echarts from 'echarts'
import { nextTick, onBeforeUnmount, ref } from 'vue'

interface QuoteBar {
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

const secuCode = ref('')
const date = ref('')
const period = ref('Min')
const loading = ref(false)
const error = ref('')
const summary = ref('')
const chartEl = ref<HTMLDivElement | null>(null)

let chart: echarts.ECharts | null = null

onBeforeUnmount(() => {
  chart?.dispose()
  chart = null
})

function formatTime(time: number): string {
  const s = String(time).padStart(6, '0')
  return `${s.slice(0, 2)}:${s.slice(2, 4)}:${s.slice(4, 6)}`
}

async function query() {
  error.value = ''
  summary.value = ''
  if (!secuCode.value.trim()) {
    error.value = '证券代码不能为空'
    return
  }
  if (!/^\d{8}$/.test(date.value)) {
    error.value = '日期必须为 yyyymmdd 格式（如 20260803）'
    return
  }

  loading.value = true
  try {
    const params = new URLSearchParams({
      secu_code: secuCode.value.trim(),
      date: date.value,
      period: period.value,
    })
    const response = await fetch(`/API/V1/Quote/Query?${params}`)
    const body = await response.json()
    if (body.code !== 0) {
      error.value = body.message || '查询失败'
      return
    }
    const data = body.data ?? {}
    const bars: QuoteBar[] = data.bars ?? []
    summary.value = `${data.secu_code ?? '?'} ${data.date ?? '?'} ${data.period ?? 'Min'}：共 ${data.count ?? 0} 根分钟 K 线`
    await nextTick()
    renderChart(bars)
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
  const times = bars.map((b) => formatTime(b.time))
  const candleData = bars.map((b) => [
    b.open ?? 0,
    b.close ?? 0,
    b.low ?? 0,
    b.high ?? 0,
  ])
  const volumes = bars.map((b) => b.volume ?? null)

  chart.setOption({
    animation: false,
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: unknown) => {
        const list = params as Array<{ axisValue: string; dataIndex: number; seriesName: string; value: number | number[] }>
        if (!list?.length) return ''
        const index = list[0].dataIndex
        const bar = bars[index]
        const rows: string[] = [`<b>${bar ? formatTime(bar.time) : list[0].axisValue}</b>`]
        const price = (v: string | null) => (v !== null && v !== undefined ? Number(v).toFixed(4) : '—')
        rows.push(`开盘价：${price(bar?.open ?? null)}`)
        rows.push(`收盘价：${price(bar?.close ?? null)}`)
        rows.push(`最高价：${price(bar?.high ?? null)}`)
        rows.push(`最低价：${price(bar?.low ?? null)}`)
        // 有值才展示：成交量 / 成交额 / 涨跌额 / 涨跌幅
        if (bar?.volume !== null && bar?.volume !== undefined) rows.push(`成交量：${bar.volume}`)
        if (bar?.turnover) rows.push(`成交额：${bar.turnover}`)
        if (bar?.change) rows.push(`涨跌额：${bar.change}`)
        if (bar?.change_pct) rows.push(`涨跌幅：${Number(bar.change_pct).toFixed(4)}%`)
        if (bar?.remark) rows.push(`备注：${bar.remark}`)
        return rows.join('<br/>')
      },
    },
    legend: { data: ['K线', '成交量'] },
    grid: [
      { left: 60, right: 20, top: 20, height: '55%' },
      { left: 60, right: 20, top: '72%', height: '18%' },
    ],
    xAxis: [
      { type: 'category', data: times, boundaryGap: true, axisLabel: { rotate: 45 } },
      { type: 'category', data: times, gridIndex: 1, axisLabel: { show: false } },
    ],
    yAxis: [
      { scale: true, splitLine: { show: true } },
      { gridIndex: 1, splitLine: { show: false } },
    ],
    dataZoom: [{ type: 'inside', xAxisIndex: [0, 1], start: 0, end: 100 }],
    series: [
      {
        name: 'K线',
        type: 'candlestick',
        data: candleData,
        itemStyle: {
          // 红涨绿跌（中国习惯）
          color: '#ef232a',   // 阳线（涨）
          color0: '#14b143',  // 阴线（跌）
          borderColor: '#ef232a',
          borderColor0: '#14b143',
        },
      },
      {
        name: '成交量',
        type: 'bar',
        xAxisIndex: 1,
        yAxisIndex: 1,
        data: volumes,
        itemStyle: {
          color: (param: unknown) => {
            const p = param as { dataIndex: number }
            const bar = bars[p.dataIndex]
            const up = Number(bar?.close ?? 0) >= Number(bar?.open ?? 0)
            return up ? '#ef232a' : '#14b143'
          },
        },
      },
    ],
  }, true)
}
</script>

<template>
  <v-card max-width="960" class="mx-auto">
    <v-card-title>
      <v-icon icon="mdi-chart-candlestick" class="mr-2" />
      历史行情查询
    </v-card-title>
    <v-card-subtitle>
      按证券代码 + 交易日期查询分钟级 K 线（周期目前仅支持 1 分钟 Min）；红涨绿跌，悬停查看详情
    </v-card-subtitle>
    <v-card-text>
      <v-form @submit.prevent="query">
        <v-row align="center">
          <v-col cols="3">
            <v-text-field v-model="secuCode" label="证券代码" placeholder="如: NVDA" />
          </v-col>
          <v-col cols="3">
            <v-text-field v-model="date" label="交易日期" placeholder="如: 20260803" hint="yyyymmdd" />
          </v-col>
          <v-col cols="3">
            <v-select
              v-model="period"
              label="周期"
              :items="[{ title: '1分（Min）', value: 'Min' }]"
            />
          </v-col>
          <v-col cols="3">
            <v-btn type="submit" color="primary" :loading="loading" prepend-icon="mdi-magnify">
              查询
            </v-btn>
          </v-col>
        </v-row>
      </v-form>

      <v-alert v-if="error" type="error" class="mt-2" density="compact">{{ error }}</v-alert>
      <v-alert v-if="summary" type="info" class="mt-2" density="compact">{{ summary }}</v-alert>

      <div
        ref="chartEl"
        class="mt-2"
        style="width: 100%; height: 520px"
      />
    </v-card-text>
  </v-card>
</template>
