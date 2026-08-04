<script setup lang="ts">
import { markRaw, ref, type Component } from 'vue'

import DashboardView from './views/DashboardView.vue'
import QuoteImportView from './views/QuoteImportView.vue'
import QuoteQueryView from './views/QuoteQueryView.vue'

type ViewName = 'dashboard' | 'quote-import' | 'quote-query'

const currentView = ref<ViewName>('dashboard')

// 直接存组件对象（markRaw 避免响应式代理）；:is 需要组件而非 Ref
const views: Record<ViewName, Component> = {
  'dashboard': markRaw(DashboardView),
  'quote-import': markRaw(QuoteImportView),
  'quote-query': markRaw(QuoteQueryView),
}

const menuItems = [
  { key: 'dashboard' as ViewName, title: '仪表盘', icon: 'mdi-view-dashboard' },
  { key: 'quote-query' as ViewName, title: '历史行情查询', icon: 'mdi-chart-candlestick' },
  { key: 'quote-import' as ViewName, title: '历史行情数据导入', icon: 'mdi-database-import' },
]

const drawer = ref(true)
</script>

<template>
  <v-app>
    <v-app-bar color="primary" density="comfortable">
      <v-app-bar-title>
        <v-icon icon="mdi-finance" class="mr-2" />
        FinvQuant 量化策略交易平台
      </v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer v-model="drawer">
      <v-list density="comfortable" nav>
        <v-list-item
          v-for="item in menuItems"
          :key="item.key"
          :prepend-icon="item.icon"
          :title="item.title"
          :active="currentView === item.key"
          @click="currentView = item.key"
        />
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <v-container fluid class="mt-4">
        <component :is="views[currentView]" />
      </v-container>
    </v-main>
  </v-app>
</template>
