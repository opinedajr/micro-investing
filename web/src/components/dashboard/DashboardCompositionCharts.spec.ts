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

  it('renders an empty risk gauge when risk data is not loaded', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation, risk: null },
    })

    const chart = stubChartData(wrapper, 'risk-chart')

    expect(chart.data.datasets[0].data).toEqual([0, 5])
    expect(wrapper.text()).toContain('0,0')
  })

  it('renders empty allocation datasets when allocation data is not loaded', () => {
    const wrapper = mount(DashboardCompositionCharts, {
      props: { allocation: null, risk },
    })

    const chart = stubChartData(wrapper, 'allocation-chart')

    expect(chart.data.labels).toEqual([])
    expect(chart.data.datasets[0].data).toEqual([])
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
})
