<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import DashboardCompositionCharts from '@/components/dashboard/DashboardCompositionCharts.vue'
import DashboardEvolution from '@/components/dashboard/DashboardEvolution.vue'
import DashboardKpiCards from '@/components/dashboard/DashboardKpiCards.vue'
import { walletApi } from '@/lib/api'
import { useDashboardStore } from '@/stores/dashboard'

const store = useDashboardStore()
const { summary, allocation, risk, evolution } = storeToRefs(store)

const walletId = ref<string | null>(null)
const selectedYear = ref<number | null>(null)
const selectedQuarter = ref<number | null>(null)

onMounted(async () => {
  try {
    const wallets = await walletApi.list()
    const id = wallets[0]?.id
    if (id) {
      walletId.value = id
      await Promise.all([store.fetchCurrentState(id), refetchEvolution()])
    }
  } catch (error) {
    store.error = error instanceof Error ? error.message : 'Unexpected error'
  }
})

async function refetchEvolution() {
  if (!walletId.value) {
    return
  }
  await store.fetchEvolution(
    walletId.value,
    selectedYear.value ?? undefined,
    selectedQuarter.value ?? undefined,
  )
}

function onYearChange(value: number | null) {
  selectedYear.value = value
  if (value === null) {
    selectedQuarter.value = null
  }
  void refetchEvolution()
}

function onQuarterChange(value: number | null) {
  selectedQuarter.value = value
  void refetchEvolution()
}
</script>

<template>
  <main class="dashboard">
    <header class="dashboard__header">
      <h1 class="dashboard__title">Dashboard</h1>
    </header>

    <div class="dashboard__grid">
      <DashboardKpiCards :summary="summary" />
      <DashboardCompositionCharts :allocation="allocation" :risk="risk" />
      <DashboardEvolution
        :evolution="evolution"
        :year="selectedYear"
        :quarter="selectedQuarter"
        @change-year="onYearChange"
        @change-quarter="onQuarterChange"
      />
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
