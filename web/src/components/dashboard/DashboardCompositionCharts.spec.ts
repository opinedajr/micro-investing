import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import type { DashboardAllocation, DashboardRisk } from '@/lib/api'
import DashboardCompositionCharts from './DashboardCompositionCharts.vue'

const allocation: DashboardAllocation = {
  items: [
    { type: 'stocks', amount: 500000, percentage: 50 },
    { type: 'fixed_income', amount: 300000, percentage: 30 },
    { type: 'emergency_reserve', amount: 200000, percentage: 20 },
  ],
  total: 1000000,
}

const risk: DashboardRisk = {
  items: [
    { rank: 1, amount: 250000, percentage: 25 },
    { rank: 5, amount: 750000, percentage: 75 },
  ],
  total: 1000000,
}

function stubChartData(wrapper: ReturnType<typeof mount>, containerTestId: string) {
  const stub = wrapper.find(`[data-testid="${containerTestId}"] [data-testid="chart-stub"]`)
  return {
    exists: stub.exists(),
    type: stub.attributes('data-chart-type'),
    data: JSON.parse(stub.attributes('data-chart-data') ?? '{}'),
    options: JSON.parse(stub.attributes('data-chart-options') ?? '{}'),
    plugins: JSON.parse(stub.attributes('data-chart-plugins') ?? '[]'),
  }
}

describe('DashboardCompositionCharts', () => {
  it('renders the allocation doughnut chart with human readable labels and amounts', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'allocation-chart')

    expect(chart.exists).toBe(true)
    expect(chart.type).toBe('doughnut')
    expect(chart.data.labels).toEqual(['Ações', 'Renda Fixa', 'Reserva de Emergência'])
    expect(chart.data.datasets[0].data).toEqual([500000, 300000, 200000])
  })

  it('draws percentage labels inside the allocation chart', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'allocation-chart')

    expect(chart.options.plugins.percentageLabels.labels).toEqual(['50%', '30%', '20%'])
    expect(chart.plugins).toEqual([{ id: 'percentageLabels' }])
  })

  it('renders the risk gauge as a half circle driven by the weighted risk score', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.exists).toBe(true)
    expect(chart.type).toBe('doughnut')
    expect(chart.options.circumference).toBe(180)
    expect(chart.options.rotation).toBe(-90)
    expect(chart.data.datasets[0].data[0]).toBeCloseTo(4, 5)
    expect(chart.data.datasets[0].data[1]).toBeCloseTo(1, 5)
  })

  it('draws the score percentage label inside the risk gauge', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.options.plugins.percentageLabels.labels).toEqual(['80%'])
    expect(chart.plugins).toEqual([{ id: 'percentageLabels' }])
  })

  it('renders an empty risk gauge when risk data is not loaded', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk: null },
    })

    const chart = stubChartData(wrapper, 'risk-chart')
    const legend = wrapper.find('[data-testid="risk-legend"]')

    expect(chart.data.datasets[0].data).toEqual([0, 5])
    expect(chart.options.plugins.percentageLabels.labels).toEqual([])
    expect(legend.text()).toBe('Sem pontuação')
  })

  it('renders empty allocation datasets when allocation data is not loaded', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation: null, risk },
    })

    const chart = stubChartData(wrapper, 'allocation-chart')

    expect(chart.data.labels).toEqual([])
    expect(chart.data.datasets[0].data).toEqual([])
    expect(chart.options.plugins.percentageLabels.labels).toEqual([])
  })

  it('renders the dividends card between allocation and risk cards', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const titles = wrapper.findAll('.composition__card .composition__title')

    expect(titles.map((title) => title.text())).toEqual([
      'Alocação de Patrimônio',
      'Dividendos',
      'Gerenciamento de Risco',
    ])
  })

  it('renders the dividends area as a reserved empty state', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const dividends = wrapper.find('[data-testid="dividends-placeholder"]')

    expect(dividends.exists()).toBe(true)
    expect(dividends.text()).toContain('Dividendos')
    expect(dividends.text()).toContain('Em breve')
  })

  it('removes the static risk profile text from the risk card', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    expect(wrapper.text()).not.toContain('Perfil de risco')
  })

  it.each([
    { items: [{ rank: 1, amount: 1000000, percentage: 100 }], expected: 'Alto Risco' },
    { items: [{ rank: 2, amount: 1000000, percentage: 100 }], expected: 'Risco' },
    { items: [{ rank: 3, amount: 1000000, percentage: 100 }], expected: 'Boa' },
    { items: [{ rank: 4, amount: 1000000, percentage: 100 }], expected: 'Ótima' },
    { items: [{ rank: 5, amount: 1000000, percentage: 100 }], expected: 'Excelente' },
    {
      items: [
        { rank: 4, amount: 750000, percentage: 75 },
        { rank: 5, amount: 250000, percentage: 25 },
      ],
      expected: 'Ótima',
    },
  ])(
    'maps the weighted risk score $items.0.rank to the dynamic legend $expected',
    ({ items, expected }) => {
      const wrapper = mount(DashboardCompositionCharts, {
        props: {
          allocation,
          risk: { items, total: 1000000 },
        },
      })

      const legend = wrapper.find('[data-testid="risk-legend"]')

      expect(legend.text()).toBe(expected)
    },
  )

  it('standardizes both charts to the same responsive size', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const allocationChart = stubChartData(wrapper, 'allocation-chart')
    const riskChart = stubChartData(wrapper, 'risk-chart')

    expect(allocationChart.options.responsive).toBe(true)
    expect(allocationChart.options.maintainAspectRatio).toBe(true)
    expect(riskChart.options.responsive).toBe(true)
    expect(riskChart.options.maintainAspectRatio).toBe(true)
  })
})
