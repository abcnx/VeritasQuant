<script setup lang="ts">
import { ref, shallowRef } from 'vue'

import DashboardView from './views/DashboardView.vue'
import QuoteImportView from './views/QuoteImportView.vue'

type ViewName = 'dashboard' | 'quote-import'

const currentView = ref<ViewName>('dashboard')

const views: Record<ViewName, unknown> = {
  'dashboard': shallowRef(DashboardView),
  'quote-import': shallowRef(QuoteImportView),
}

const menuItems = [
  { key: 'dashboard' as ViewName, title: '仪表盘', icon: 'mdi-view-dashboard' },
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
      <v-container class="mt-4">
        <component :is="views[currentView]" />
      </v-container>
    </v-main>
  </v-app>
</template>
