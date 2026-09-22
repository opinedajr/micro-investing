<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'

import DashboardCompositionCharts from '@/components/dashboard/DashboardCompositionCharts.vue'
import DashboardKpiCards from '@/components/dashboard/DashboardKpiCards.vue'
import { walletApi } from '@/lib/api'
import { useDashboardStore } from '@/stores/dashboard'

const store = useDashboardStore()
const { summary, allocation, risk } = storeToRefs(store)

onMounted(async () => {
  try {
    const wallets = await walletApi.list()
    const walletId = wallets[0]?.id
    if (walletId) {
      await store.fetchCurrentState(walletId)
    }
  } catch (error) {
    store.error = error instanceof Error ? error.message : 'Unexpected error'
  }
})
</script>

<template>
  <main class="dashboard">
    <header class="dashboard__header">
      <h1 class="dashboard__title">Dashboard</h1>
    </header>

    <div class="dashboard__grid">
      <DashboardKpiCards :summary="summary" />
      <DashboardCompositionCharts :allocation="allocation" :risk="risk" />
    </div>
  </main>
</template>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem;
}

.dashboard__header {
  margin-bottom: 1.5rem;
}

.dashboard__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.dashboard__grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 1rem;
}
</style>
