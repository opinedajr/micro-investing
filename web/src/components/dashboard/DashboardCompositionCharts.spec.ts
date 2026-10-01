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
    { rank: 3, amount: 4620000, percentage: 62 },
    { rank: 4, amount: 1940000, percentage: 26 },
    { rank: 5, amount: 920000, percentage: 12 },
  ],
  total: 7480000,
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

  it('renders one risk slice per dashboard/risk item', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.exists).toBe(true)
    expect(chart.type).toBe('doughnut')
    expect(chart.data.labels).toEqual(['Boa', 'Ótimo', 'Excelente'])
    expect(chart.data.datasets[0].data).toEqual([4620000, 1940000, 920000])
  })

  it('draws percentage labels inside each risk slice', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.options.plugins.percentageLabels.labels).toEqual(['62%', '26%', '12%'])
    expect(chart.plugins).toEqual([{ id: 'percentageLabels' }])
  })

  it('applies semantic colors per risk rank', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.datasets[0].backgroundColor).toEqual(['#eab308', '#0891b2', '#16a34a'])
  })

  it('renders the risk chart as a half-circle gauge without the native legend', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.options.cutout).toBe('65%')
    expect(chart.options.circumference).toBe(180)
    expect(chart.options.rotation).toBe(-90)
    expect(chart.options.plugins.legend).toEqual({ display: false })
  })

  it('renders one risk slice per dashboard/risk item, skipping ranks without data', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: {
        allocation,
        risk: {
          items: [
            { rank: 2, amount: 3000000, percentage: 40 },
            { rank: 5, amount: 4500000, percentage: 60 },
          ],
          total: 7500000,
        },
      },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual(['Risco', 'Excelente'])
    expect(chart.data.datasets[0].data).toEqual([3000000, 4500000])
    expect(chart.data.datasets[0].backgroundColor).toEqual(['#f59e0b', '#16a34a'])
    expect(chart.data.datasets[0].data).toHaveLength(2)
  })

  it('renders the custom risk legend beside the gauge with label and rank', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk },
    })

    const legend = wrapper.find('[data-testid="risk-legend"]')
    const items = legend.findAll('.composition__risk-legend-item')
    const swatches = legend.findAll('.composition__risk-legend-swatch')

    expect(legend.exists()).toBe(true)
    expect(items.map((item) => item.text())).toEqual(['Boa (3)', 'Ótimo (4)', 'Excelente (5)'])
    expect(swatches).toHaveLength(3)
    expect((swatches[0].element as HTMLElement).style.backgroundColor).toBe('rgb(234, 179, 8)')
    expect((swatches[1].element as HTMLElement).style.backgroundColor).toBe('rgb(8, 145, 178)')
    expect((swatches[2].element as HTMLElement).style.backgroundColor).toBe('rgb(22, 163, 74)')
  })

  it('renders an empty risk chart when risk data is not loaded', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk: null },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual([])
    expect(chart.data.datasets[0].data).toEqual([])
    expect(chart.data.datasets[0].backgroundColor).toEqual([])
    expect(chart.options.plugins.percentageLabels.labels).toEqual([])
    expect(wrapper.find('[data-testid="risk-legend"]').exists()).toBe(false)
  })

  it('renders an empty risk chart when the wallet has no risk items', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk: { items: [], total: 0 } },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual([])
    expect(chart.data.datasets[0].data).toEqual([])
    expect(chart.data.datasets[0].backgroundColor).toEqual([])
    expect(chart.options.plugins.percentageLabels.labels).toEqual([])
    expect(wrapper.find('[data-testid="risk-legend"]').exists()).toBe(false)
  })

  it('renders an empty risk chart when the wallet has no risk items', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk: { items: [], total: 0 } },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual([])
    expect(chart.data.datasets[0].data).toEqual([])
    expect(chart.data.datasets[0].backgroundColor).toEqual([])
    expect(chart.options.plugins.percentageLabels.labels).toEqual([])
    expect(wrapper.find('[data-testid="risk-legend"]').exists()).toBe(false)
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
      'Gerenciamento de Risco (notas)',
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
    { rank: 1, expected: ['Alto Risco'] },
    { rank: 2, expected: ['Risco'] },
    { rank: 3, expected: ['Boa'] },
    { rank: 4, expected: ['Ótimo'] },
    { rank: 5, expected: ['Excelente'] },
  ])('maps rank $rank to the risk chart label $expected', ({ rank, expected }) => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: {
        allocation,
        risk: { items: [{ rank, amount: 1000000, percentage: 100 }], total: 1000000 },
      },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual(expected)
  })

  it('maps unknown risk ranks to a fallback label and fallback color', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: {
        allocation,
        risk: { items: [{ rank: 9, amount: 500000, percentage: 50 }], total: 500000 },
      },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.labels).toEqual(['Risco 9'])
    expect(chart.data.datasets[0].backgroundColor).toEqual(['#0ea5e9'])
  })

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
