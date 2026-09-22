<script setup lang="ts">
import { computed } from 'vue'
import Chart from 'primevue/chart'

import type { DashboardAllocation, DashboardRisk } from '@/lib/api'
import { formatCurrencyBRL } from '@/lib/utils'

const props = defineProps<{
  allocation: DashboardAllocation | null
  risk: DashboardRisk | null
}>()

const ALLOCATION_LABELS: Record<string, string> = {
  stocks: 'Ações',
  fixed_income: 'Renda Fixa',
  emergency_reserve: 'Reserva de Emergência',
  liquid_cash: 'Caixa',
  fiis: 'FIIs',
}

const ALLOCATION_COLORS: Record<string, string> = {
  stocks: '#0ea5e9',
  fixed_income: '#22c55e',
  emergency_reserve: '#f59e0b',
  liquid_cash: '#64748b',
  fiis: '#a855f7',
}

const FALLBACK_COLORS = ['#0ea5e9', '#22c55e', '#f59e0b', '#a855f7', '#64748b']

const RISK_TRACK_COLOR = '#e5e7eb'
const MAX_RISK_SCORE = 5

interface TooltipContext {
  label: string
  parsed: number
}

function allocationLabel(type: string): string {
  return ALLOCATION_LABELS[type] ?? type
}

function allocationColor(type: string, index: number): string {
  return ALLOCATION_COLORS[type] ?? FALLBACK_COLORS[index % FALLBACK_COLORS.length]
}

function riskScoreColor(score: number): string {
  if (score <= 1.5) {
    return '#22c55e'
  }
  if (score <= 2.5) {
    return '#84cc16'
  }
  if (score <= 3.5) {
    return '#facc15'
  }
  if (score <= 4.5) {
    return '#f97316'
  }
  return '#ef4444'
}

const allocationChartData = computed(() => ({
  labels: props.allocation?.items.map((item) => allocationLabel(item.type)) ?? [],
  datasets: [
    {
      data: props.allocation?.items.map((item) => item.amount) ?? [],
      backgroundColor:
        props.allocation?.items.map((item, index) => allocationColor(item.type, index)) ?? [],
      borderWidth: 2,
    },
  ],
}))

const allocationChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '65%',
  plugins: {
    legend: {
      position: 'bottom',
    },
    tooltip: {
      callbacks: {
        label(context: TooltipContext) {
          return `${context.label}: ${formatCurrencyBRL(context.parsed)}`
        },
      },
    },
  },
}

const riskScore = computed(() => {
  const items = props.risk?.items ?? []
  const totalPercentage = items.reduce((sum, item) => sum + item.percentage, 0)
  if (totalPercentage <= 0) {
    return 0
  }
  return items.reduce((sum, item) => sum + item.rank * item.percentage, 0) / totalPercentage
})

const riskChartData = computed(() => {
  const score = riskScore.value
  return {
    labels: ['Risco', 'Restante'],
    datasets: [
      {
        data: [score, Math.max(0, MAX_RISK_SCORE - score)],
        backgroundColor:
          score > 0 ? [riskScoreColor(score), RISK_TRACK_COLOR] : [RISK_TRACK_COLOR, RISK_TRACK_COLOR],
        borderWidth: 0,
      },
    ],
  }
})

const riskChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '72%',
  circumference: 180,
  rotation: -90,
  plugins: {
    legend: {
      display: false,
    },
  },
}

const riskScoreLabel = computed(() =>
  new Intl.NumberFormat('pt-BR', { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(
    riskScore.value,
  ),
)
</script>

<template>
  <section class="composition" data-testid="dashboard-composition-charts">
    <article class="composition__card" data-testid="allocation-chart">
      <h2 class="composition__title">Alocação de Patrimônio</h2>
      <div class="composition__chart composition__chart--allocation">
        <Chart type="doughnut" :data="allocationChartData" :options="allocationChartOptions" />
      </div>
    </article>

    <article class="composition__card" data-testid="risk-chart">
      <h2 class="composition__title">Gerenciamento de Risco</h2>
      <div class="composition__chart composition__chart--risk">
        <Chart type="doughnut" :data="riskChartData" :options="riskChartOptions" />
      </div>
      <p class="composition__risk-score">
        Perfil de risco: <strong>{{ riskScoreLabel }} / 5</strong>
      </p>
    </article>

    <article class="composition__card composition__card--reserved" data-testid="dividends-placeholder">
      <h2 class="composition__title">Dividendos</h2>
      <div class="composition__placeholder">
        <i class="pi pi-chart-line composition__placeholder-icon" aria-hidden="true"></i>
        <p class="composition__placeholder-title">Em breve</p>
        <span class="composition__placeholder-hint">
          O gerenciamento de dividendos ainda não está disponível.
        </span>
      </div>
    </article>
  </section>
</template>

<style scoped>
.composition {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.composition__card {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-radius: 0.75rem;
  background: var(--p-content-background, #ffffff);
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.06);
}

.composition__title {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.composition__chart {
  position: relative;
  width: 100%;
}

.composition__chart--allocation {
  height: 260px;
}

.composition__chart--risk {
  height: 160px;
}

.composition__risk-score {
  margin: 0;
  font-size: 0.875rem;
  text-align: center;
  color: var(--p-text-muted-color, #64748b);
}

.composition__risk-score strong {
  color: var(--p-text-color, #0f172a);
}

.composition__card--reserved {
  border-style: dashed;
  border-color: var(--p-orange-300, #fdba74);
}

.composition__placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 1.5rem 1rem;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--p-orange-500, #f97316) 6%, transparent);
}

.composition__placeholder-icon {
  font-size: 1.5rem;
  color: var(--p-orange-500, #f97316);
}

.composition__placeholder-title {
  margin: 0.25rem 0 0;
  font-weight: 600;
  color: var(--p-orange-600, #ea580c);
}

.composition__placeholder-hint {
  font-size: 0.8125rem;
  color: var(--p-text-muted-color, #64748b);
  text-align: center;
}

@media (max-width: 960px) {
  .composition {
    grid-template-columns: 1fr;
  }
}
</style>
