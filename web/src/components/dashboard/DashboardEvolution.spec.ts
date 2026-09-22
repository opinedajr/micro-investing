import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { DashboardEvolution as DashboardEvolutionData } from '@/lib/api'
import DashboardEvolution from './DashboardEvolution.vue'

const evolution: DashboardEvolutionData = {
  total: [
    { year: 2026, month: 8, amount: 100000 },
    { year: 2026, month: 9, amount: 150000 },
  ],
  by_category: {
    fixed_income: [
      { year: 2026, month: 8, amount: 40000 },
      { year: 2026, month: 9, amount: 60000 },
    ],
    stocks: [
      { year: 2026, month: 8, amount: 50000 },
      { year: 2026, month: 9, amount: 70000 },
    ],
    emergency_reserve: [
      { year: 2026, month: 8, amount: 10000 },
      { year: 2026, month: 9, amount: 20000 },
    ],
  },
}

function stubChartData(wrapper: ReturnType<typeof mount>, containerTestId: string) {
  const stub = wrapper.find(`[data-testid="${containerTestId}"] [data-testid="chart-stub"]`)
  return {
    exists: stub.exists(),
    type: stub.attributes('data-chart-type'),
    data: JSON.parse(stub.attributes('data-chart-data') ?? '{}'),
    options: JSON.parse(stub.attributes('data-chart-options') ?? '{}'),
  }
}

function mountEvolution(props: Partial<{ evolution: DashboardEvolutionData | null; year: number | null; quarter: number | null }> = {}) {
  return mount(DashboardEvolution, {
    props: {
      evolution: props.evolution !== undefined ? props.evolution : evolution,
      year: props.year ?? null,
      quarter: props.quarter ?? null,
    },
  })
}

describe('DashboardEvolution', () => {
  it('renders the main bar chart with month labels and amounts from the evolution data', () => {
    const wrapper = mountEvolution()

    const chart = stubChartData(wrapper, 'evolution-chart')

    expect(chart.exists).toBe(true)
    expect(chart.type).toBe('bar')
    expect(chart.data.labels).toEqual(['ago/26', 'set/26'])
    expect(chart.data.datasets[0].data).toEqual([100000, 150000])
  })

  it('renders the three mini charts with each category series', () => {
    const wrapper = mountEvolution()

    const fixedIncome = stubChartData(wrapper, 'mini-chart-fixed_income')
    const stocks = stubChartData(wrapper, 'mini-chart-stocks')
    const reserve = stubChartData(wrapper, 'mini-chart-emergency_reserve')

    expect(fixedIncome.exists).toBe(true)
    expect(fixedIncome.type).toBe('bar')
    expect(fixedIncome.data.datasets[0].data).toEqual([40000, 60000])
    expect(stocks.exists).toBe(true)
    expect(stocks.data.datasets[0].data).toEqual([50000, 70000])
    expect(reserve.exists).toBe(true)
    expect(reserve.data.datasets[0].data).toEqual([10000, 20000])
    expect(wrapper.text()).toContain('Renda Fixa')
    expect(wrapper.text()).toContain('Ações')
    expect(wrapper.text()).toContain('Reserva de Emergência')
  })

  it('renders empty chart datasets when evolution data is not loaded', () => {
    const wrapper = mountEvolution({ evolution: null })

    const mainChart = stubChartData(wrapper, 'evolution-chart')
    const fixedIncome = stubChartData(wrapper, 'mini-chart-fixed_income')

    expect(mainChart.data.labels).toEqual([])
    expect(mainChart.data.datasets[0].data).toEqual([])
    expect(fixedIncome.data.datasets[0].data).toEqual([])
  })

  it('renders the year and quarter filter controls with their options', () => {
    const wrapper = mountEvolution()
    const currentYear = new Date().getFullYear()

    const yearSelect = wrapper.find('[data-testid="year-select"]')
    const quarterSelect = wrapper.find('[data-testid="quarter-select"]')
    const yearOptions = yearSelect.findAll('option').map((option) => option.text())
    const quarterOptions = quarterSelect.findAll('option').map((option) => option.text())

    expect(yearSelect.exists()).toBe(true)
    expect(quarterSelect.exists()).toBe(true)
    expect(yearOptions).toEqual(['Todos', ...Array.from({ length: 5 }, (_, index) => String(currentYear - index))])
    expect(quarterOptions).toEqual([
      'Ano inteiro',
      '1º Trimestre',
      '2º Trimestre',
      '3º Trimestre',
      '4º Trimestre',
    ])
  })

  it('disables the quarter select until a year is selected', async () => {
    const withoutYear = mountEvolution({ year: null })
    const withYear = mountEvolution({ year: 2026 })

    expect(withoutYear.find('[data-testid="quarter-select"]').attributes('disabled')).toBeDefined()
    expect(withYear.find('[data-testid="quarter-select"]').attributes('disabled')).toBeUndefined()
  })

  it('emits change-year with the selected year or null when cleared', async () => {
    const wrapper = mountEvolution()

    await wrapper.find('[data-testid="year-select"]').setValue('2025')
    expect(wrapper.emitted('change-year')?.[0]).toEqual([2025])

    await wrapper.find('[data-testid="year-select"]').setValue('')
    expect(wrapper.emitted('change-year')?.[1]).toEqual([null])
  })

  it('emits change-quarter with the selected quarter or null when cleared', async () => {
    const wrapper = mountEvolution({ year: 2026 })

    await wrapper.find('[data-testid="quarter-select"]').setValue('3')
    expect(wrapper.emitted('change-quarter')?.[0]).toEqual([3])

    await wrapper.find('[data-testid="quarter-select"]').setValue('')
    expect(wrapper.emitted('change-quarter')?.[1]).toEqual([null])
  })
})
