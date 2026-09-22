<script setup lang="ts">
import { computed } from 'vue'
import Chart from 'primevue/chart'

import type { DashboardEvolution } from '@/lib/api'
import { formatCurrencyBRL } from '@/lib/utils'

const props = defineProps<{
  evolution: DashboardEvolution | null
  year: number | null
  quarter: number | null
}>()

const emit = defineEmits<{
  (event: 'change-year', value: number | null): void
  (event: 'change-quarter', value: number | null): void
}>()

interface CategoryDefinition {
  key: 'fixed_income' | 'stocks' | 'emergency_reserve'
  label: string
  color: string
}

interface TooltipContext {
  label: string
  parsed: { x: number; y: number }
}

const CATEGORIES: CategoryDefinition[] = [
  { key: 'fixed_income', label: 'Renda Fixa', color: '#22c55e' },
  { key: 'stocks', label: 'Ações', color: '#0ea5e9' },
  { key: 'emergency_reserve', label: 'Reserva de Emergência', color: '#f59e0b' },
]

const QUARTER_ORDINALS = ['1º', '2º', '3º', '4º']
const YEAR_OPTIONS_COUNT = 5
const TOTAL_COLOR = '#0ea5e9'

const MONTH_FORMATTER = new Intl.DateTimeFormat('pt-BR', { month: 'short' })

function monthLabel(year: number, month: number): string {
  const shortMonth = MONTH_FORMATTER.format(new Date(year, month - 1, 1)).replace('.', '')
  return `${shortMonth}/${String(year).slice(-2)}`
}

function quarterLabel(quarter: number): string {
  return `${QUARTER_ORDINALS[quarter - 1]} Trimestre`
}

function tooltipValue(context: TooltipContext): string {
  return formatCurrencyBRL(context.parsed.y)
}

const currentYear = new Date().getFullYear()
const yearOptions = Array.from({ length: YEAR_OPTIONS_COUNT }, (_, index) => currentYear - index)

const periodLabels = computed(
  () => props.evolution?.total.map((month) => monthLabel(month.year, month.month)) ?? [],
)

const totalChartData = computed(() => ({
  labels: periodLabels.value,
  datasets: [
    {
      label: 'Patrimônio',
      data: props.evolution?.total.map((month) => month.amount) ?? [],
      backgroundColor: TOTAL_COLOR,
      borderRadius: 4,
    },
  ],
}))

function categoryChartData(category: CategoryDefinition) {
  return {
    labels: periodLabels.value,
    datasets: [
      {
        label: category.label,
        data: props.evolution?.by_category[category.key]?.map((month) => month.amount) ?? [],
        backgroundColor: category.color,
        borderRadius: 4,
      },
    ],
  }
}

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      callbacks: {
        label: tooltipValue,
      },
    },
  },
  scales: {
    y: {
      beginAtZero: true,
    },
  },
}

function onYearChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('change-year', value === '' ? null : Number(value))
}

function onQuarterChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('change-quarter', value === '' ? null : Number(value))
}
</script>

<template>
  <section class="evolution" data-testid="dashboard-evolution">
    <div class="evolution__toolbar">
      <div class="evolution__filter">
        <label class="evolution__filter-label" for="evolution-year">Ano</label>
        <select
          id="evolution-year"
          class="evolution__select"
          data-testid="year-select"
          :value="year ?? ''"
          @change="onYearChange"
        >
          <option value="">Todos</option>
          <option v-for="yearOption in yearOptions" :key="yearOption" :value="yearOption">
            {{ yearOption }}
          </option>
        </select>
      </div>

      <div class="evolution__filter">
        <label class="evolution__filter-label" for="evolution-quarter">Trimestre</label>
        <select
          id="evolution-quarter"
          class="evolution__select"
          data-testid="quarter-select"
          :value="quarter ?? ''"
          :disabled="year === null"
          @change="onQuarterChange"
        >
          <option value="">Ano inteiro</option>
          <option v-for="quarterOption in 4" :key="quarterOption" :value="quarterOption">
            {{ quarterLabel(quarterOption) }}
          </option>
        </select>
      </div>
    </div>

    <article class="evolution__card" data-testid="evolution-chart">
      <h2 class="evolution__title">Evolução do Patrimônio</h2>
      <div class="evolution__chart evolution__chart--main">
        <Chart type="bar" :data="totalChartData" :options="barChartOptions" />
      </div>
    </article>

    <div class="evolution__minis">
      <article
        v-for="category in CATEGORIES"
        :key="category.key"
        class="evolution__card evolution__card--mini"
        :data-testid="`mini-chart-${category.key}`"
      >
        <h2 class="evolution__title">{{ category.label }}</h2>
        <div class="evolution__chart evolution__chart--mini">
          <Chart type="bar" :data="categoryChartData(category)" :options="barChartOptions" />
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.evolution {
  grid-column: 1 / -1;
  display: grid;
  gap: 1rem;
}

.evolution__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.evolution__filter {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.evolution__filter-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--p-text-muted-color, #64748b);
}

.evolution__select {
  min-width: 10rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  background: var(--p-content-background, #ffffff);
  color: inherit;
  font: inherit;
}

.evolution__select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.evolution__card {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-radius: 0.75rem;
  background: var(--p-content-background, #ffffff);
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.06);
}

.evolution__title {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.evolution__chart {
  position: relative;
  width: 100%;
}

.evolution__chart--main {
  height: 320px;
}

.evolution__chart--mini {
  height: 160px;
}

.evolution__minis {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media (max-width: 960px) {
  .evolution__minis {
    grid-template-columns: 1fr;
  }
}
</style>
