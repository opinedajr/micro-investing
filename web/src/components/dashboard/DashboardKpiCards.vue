<script setup lang="ts">
import type { DashboardSummary } from '@/lib/api'
import { formatCurrencyBRL } from '@/lib/utils'

defineProps<{
  summary: DashboardSummary | null
}>()
</script>

<template>
  <section class="kpi-cards" data-testid="dashboard-kpi-cards">
    <article class="kpi-card kpi-card--patrimony">
      <span class="kpi-card__title">Patrimônio</span>
      <strong class="kpi-card__value">{{ formatCurrencyBRL(summary?.current_patrimony ?? 0) }}</strong>
    </article>
    <article class="kpi-card kpi-card--dividends">
      <span class="kpi-card__title">Dividendo</span>
      <strong class="kpi-card__value">{{ formatCurrencyBRL(summary?.yearly_dividends ?? 0) }}</strong>
    </article>
    <article class="kpi-card kpi-card--stocks">
      <span class="kpi-card__title">Ações</span>
      <strong class="kpi-card__value">{{ formatCurrencyBRL(summary?.stocks_invested ?? 0) }}</strong>
    </article>
  </section>
</template>

<style scoped>
.kpi-cards {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.kpi-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1.25rem 1.5rem;
  border-radius: 0.75rem;
  background: var(--p-content-background, #ffffff);
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.06);
}

.kpi-card__title {
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.kpi-card__value {
  font-size: 1.75rem;
  font-weight: 700;
}

.kpi-card--patrimony .kpi-card__value {
  color: var(--p-green-600, #16a34a);
}

.kpi-card--dividends .kpi-card__value {
  color: var(--p-orange-500, #f97316);
}

.kpi-card--stocks .kpi-card__value {
  color: var(--p-sky-600, #0284c7);
}
</style>
