import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'

import { dashboardApi, walletApi } from '@/lib/api'
import DashboardKpiCards from '@/components/dashboard/DashboardKpiCards.vue'
import { useDashboardStore } from '@/stores/dashboard'
import Dashboard from './Dashboard.vue'

vi.mock('@/lib/api', () => ({
  walletApi: {
    list: vi.fn(),
  },
  dashboardApi: {
    summary: vi.fn(),
    allocation: vi.fn(),
    risk: vi.fn(),
    evolution: vi.fn(),
  },
}))

const wallets = [{ id: 'wallet-1', name: 'Main', description: '', created_at: '', updated_at: '' }]

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the kpi cards with store values formatted as BRL', async () => {
    vi.mocked(walletApi.list).mockResolvedValue(wallets)
    vi.mocked(dashboardApi.summary).mockResolvedValue({
      current_patrimony: 1050,
      yearly_dividends: 2050,
      stocks_invested: 3050,
    })
    vi.mocked(dashboardApi.allocation).mockResolvedValue({ items: [], total: 0 })
    vi.mocked(dashboardApi.risk).mockResolvedValue({ items: [], total: 0 })

    const pinia = createPinia()
    const wrapper = mount(Dashboard, { global: { plugins: [pinia] } })

    await flushPromises()

    const store = useDashboardStore(pinia)

    expect(store.summary).toEqual({
      current_patrimony: 1050,
      yearly_dividends: 2050,
      stocks_invested: 3050,
    })
    expect(dashboardApi.summary).toHaveBeenCalledWith('wallet-1')

    const kpiCards = wrapper.findComponent(DashboardKpiCards)
    expect(kpiCards.exists()).toBe(true)
    expect(wrapper.text()).toContain('R$ 10,50')
    expect(wrapper.text()).toContain('R$ 20,50')
    expect(wrapper.text()).toContain('R$ 30,50')
  })

  it('reacts to store summary changes by re-rendering the kpi values', async () => {
    vi.mocked(walletApi.list).mockResolvedValue(wallets)
    vi.mocked(dashboardApi.summary).mockResolvedValue({
      current_patrimony: 0,
      yearly_dividends: 0,
      stocks_invested: 0,
    })
    vi.mocked(dashboardApi.allocation).mockResolvedValue({ items: [], total: 0 })
    vi.mocked(dashboardApi.risk).mockResolvedValue({ items: [], total: 0 })

    const pinia = createPinia()
    const wrapper = mount(Dashboard, { global: { plugins: [pinia] } })

    await flushPromises()

    expect(wrapper.text()).toContain('R$ 0,00')

    const store = useDashboardStore(pinia)
    store.summary = {
      current_patrimony: 987654,
      yearly_dividends: 0,
      stocks_invested: 0,
    }

    await flushPromises()

    expect(wrapper.text()).toContain('R$ 9.876,54')
  })

  it('does not fetch dashboard data when there is no wallet', async () => {
    vi.mocked(walletApi.list).mockResolvedValue([])

    const pinia = createPinia()
    mount(Dashboard, { global: { plugins: [pinia] } })

    await flushPromises()

    expect(dashboardApi.summary).not.toHaveBeenCalled()
  })
})
