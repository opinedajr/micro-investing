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

const RISK_LEGENDS: Record<number, string> = {
  1: 'Alto Risco',
  2: 'Risco',
  3: 'Boa',
  4: 'Ótima',
  5: 'Excelente',
}

const RISK_EMPTY_LEGEND = 'Sem pontuação'
const PERCENTAGE_LABEL_COLOR = '#ffffff'

interface TooltipContext {
  label: string
  parsed: number
}

interface PercentageLabelsConfig {
  labels: string[]
  color: string
}

interface ArcGeometry {
  x: number
  y: number
  innerRadius: number
  outerRadius: number
  startAngle: number
  endAngle: number
}

interface ArcElement {
  getProps(keys: string[], final: boolean): ArcGeometry
}

interface ChartCanvas {
  ctx: CanvasRenderingContext2D
  options: {
    plugins?: {
      percentageLabels?: PercentageLabelsConfig
    }
  }
  getDatasetMeta(datasetIndex: number): {
    data: ArcElement[]
  }
}

const percentageLabelsPlugin = {
  id: 'percentageLabels',
  afterDatasetsDraw(chart: ChartCanvas) {
    const config = chart.options.plugins?.percentageLabels
    if (!config || config.labels.length === 0) {
      return
    }
    const { ctx } = chart
    ctx.save()
    ctx.font = '600 12px Inter, sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillStyle = config.color
    chart.getDatasetMeta(0).data.forEach((arc, index) => {
      const label = config.labels[index]
      if (!label) {
        return
      }
      const geometry = arc.getProps(
        ['x', 'y', 'innerRadius', 'outerRadius', 'startAngle', 'endAngle'],
        true,
      )
      const midAngle = (geometry.startAngle + geometry.endAngle) / 2
      const radius = (geometry.innerRadius + geometry.outerRadius) / 2
      ctx.fillText(
        label,
        geometry.x + Math.cos(midAngle) * radius,
        geometry.y + Math.sin(midAngle) * radius,
      )
    })
    ctx.restore()
  },
}

const chartPlugins = [percentageLabelsPlugin]

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

function percentageLabel(percentage: number): string {
  return `${Math.round(percentage)}%`
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

const allocationPercentageLabels = computed(() =>
  props.allocation?.items.map((item) => percentageLabel(item.percentage)) ?? [],
)

const allocationChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: true,
  cutout: '65%',
  plugins: {
    percentageLabels: {
      labels: allocationPercentageLabels.value,
      color: PERCENTAGE_LABEL_COLOR,
    },
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
}))

const riskScore = computed(() => {
  const items = props.risk?.items ?? []
  const totalPercentage = items.reduce((sum, item) => sum + item.percentage, 0)
  if (totalPercentage <= 0) {
    return 0
  }
  return items.reduce((sum, item) => sum + item.rank * item.percentage, 0) / totalPercentage
})

const riskLegend = computed(() => {
  if (riskScore.value <= 0) {
    return RISK_EMPTY_LEGEND
  }
  const key = Math.min(MAX_RISK_SCORE, Math.max(1, Math.round(riskScore.value)))
  return RISK_LEGENDS[key] ?? RISK_EMPTY_LEGEND
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

const riskPercentageLabels = computed(() =>
  riskScore.value > 0 ? [percentageLabel((riskScore.value / MAX_RISK_SCORE) * 100)] : [],
)

const riskChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: true,
  cutout: '72%',
  circumference: 180,
  rotation: -90,
  plugins: {
    percentageLabels: {
      labels: riskPercentageLabels.value,
      color: PERCENTAGE_LABEL_COLOR,
    },
    legend: {
      display: false,
    },
  },
}))
</script>

<template>
  <section class="composition" data-testid="dashboard-composition-charts">
    <article class="composition__card" data-testid="allocation-chart">
      <h2 class="composition__title">Alocação de Patrimônio</h2>
      <div class="composition__chart">
        <Chart
          type="doughnut"
          :data="allocationChartData"
          :options="allocationChartOptions"
          :plugins="chartPlugins"
        />
      </div>
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

    <article class="composition__card" data-testid="risk-chart">
      <h2 class="composition__title">Gerenciamento de Risco</h2>
      <div class="composition__chart">
        <Chart
          type="doughnut"
          :data="riskChartData"
          :options="riskChartOptions"
          :plugins="chartPlugins"
        />
      </div>
      <p class="composition__risk-legend" data-testid="risk-legend">{{ riskLegend }}</p>
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

.composition__risk-legend {
  margin: 0;
  font-size: 0.875rem;
  text-align: center;
  color: var(--p-text-muted-color, #64748b);
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
