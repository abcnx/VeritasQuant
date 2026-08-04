<script setup lang="ts">
import { ref } from 'vue'

const file = ref<File | null>(null)
const source = ref('')
const upsertMode = ref('FIELD')
const confirmChecked = ref(false)
const importing = ref(false)
const result = ref<string>('')
const error = ref<string>('')

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

async function submitImport() {
  error.value = ''
  result.value = ''
  if (!file.value) {
    error.value = '请选择要上传的 MVSV 文件'
    return
  }
  if (!source.value.trim()) {
    error.value = '数据源不能为空'
    return
  }
  if (!confirmChecked.value) {
    error.value = '请确认导入将覆盖同时刻同证券的对应字段值'
    return
  }

  const form = new FormData()
  form.append('file', file.value)
  form.append('source', source.value.trim())
  form.append('upsert_mode', upsertMode.value)
  form.append('imported_by', 'gui')

  importing.value = true
  try {
    const response = await fetch('/API/V1/imports/upload', { method: 'POST', body: form })
    const body = await response.json()
    if (body.code !== 0) {
      error.value = body.message || '导入失败'
      return
    }
    const data = body.data ?? {}
    result.value = `导入完成：${data.secu_code ?? '?'}（market=${data.market_code ?? '?'}）` +
      ` ${data.record_count ?? 0} 条` +
      `（新增 ${data.inserted ?? 0} / 覆盖 ${data.updated ?? 0}）` +
      `｜批次 ${data.batch_id ?? '-'}｜模式 ${data.mode ?? '-'}`
  } catch {
    error.value = '网络错误：无法连接服务端'
  } finally {
    importing.value = false
  }
}
</script>

<template>
  <v-card max-width="720" class="mx-auto">
    <v-card-title>
      <v-icon icon="mdi-database-import" class="mr-2" />
      历史行情数据导入
    </v-card-title>
    <v-card-subtitle>
      上传 MVSV-1 分钟级历史行情文件，服务端解析后字段级覆盖导入 PostgreSQL（finv_quote_secu_kline_min）
    </v-card-subtitle>
    <v-card-text>
      <v-form @submit.prevent="submitImport">
        <v-file-input
          label="选择 MVSV 行情文件"
          accept=".mvsv,.txt"
          prepend-icon="mdi-file"
          @change="onFileChange"
        />
        <v-text-field v-model="source" label="数据源" placeholder="如: cn-feed" />
        <v-select
          v-model="upsertMode"
          label="覆盖模式"
          :items="[
            { title: 'FIELD（只覆盖有值的字段，推荐）', value: 'FIELD' },
            { title: 'ROW（整行覆盖）', value: 'ROW' },
          ]"
        />
        <v-checkbox v-model="confirmChecked" label="我确认导入将覆盖同时刻同证券的对应字段值" />
        <v-btn
          type="submit"
          color="primary"
          :loading="importing"
          prepend-icon="mdi-upload"
        >
          上传并导入
        </v-btn>
      </v-form>

      <v-alert v-if="error" type="error" class="mt-4" density="compact">
        {{ error }}
      </v-alert>
      <v-alert v-if="result" type="success" class="mt-4" density="compact">
        {{ result }}
      </v-alert>
    </v-card-text>
  </v-card>
</template>
