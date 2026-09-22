import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { dashboardApi } from '@/lib/api'
import { useDashboardStore } from './dashboard'

vi.mock('@/lib/api', () => ({
  dashboardApi: {
    summary: vi.fn(),
    allocation: vi.fn(),
    risk: vi.fn(),
    evolution: vi.fn(),
  },
}))

const summaryPayload = {
  current_patrimony: 1050,
  yearly_dividends: 2050,
  stocks_invested: 3050,
}

const allocationPayload = {
  items: [{ type: 'stocks', amount: 700, percentage: 66.67 }],
  total: 1050,
}

const riskPayload = {
  items: [{ rank: 3, amount: 3050, percentage: 100 }],
  total: 3050,
}

const evolutionPayload = {
  total: [{ year: 2026, month: 9, amount: 1050 }],
  by_category: {
    fixed_income: [{ year: 2026, month: 9, amount: 350 }],
    stocks: [{ year: 2026, month: 9, amount: 700 }],
    emergency_reserve: [],
  },
}

describe('useDashboardStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchCurrentState maps summary, allocation and risk payloads into their states', async () => {
    vi.mocked(dashboardApi.summary).mockResolvedValue(summaryPayload)
    vi.mocked(dashboardApi.allocation).mockResolvedValue(allocationPayload)
    vi.mocked(dashboardApi.risk).mockResolvedValue(riskPayload)

    const store = useDashboardStore()

    await store.fetchCurrentState('wallet-1')

    expect(store.summary).toEqual(summaryPayload)
    expect(store.allocation).toEqual(allocationPayload)
    expect(store.risk).toEqual(riskPayload)
    expect(dashboardApi.summary).toHaveBeenCalledWith('wallet-1')
    expect(dashboardApi.allocation).toHaveBeenCalledWith('wallet-1')
    expect(dashboardApi.risk).toHaveBeenCalledWith('wallet-1')
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchCurrentState keeps states untouched and flags error when any request fails', async () => {
    vi.mocked(dashboardApi.summary).mockRejectedValue(new Error('Request failed'))
    vi.mocked(dashboardApi.allocation).mockResolvedValue(allocationPayload)
    vi.mocked(dashboardApi.risk).mockResolvedValue(riskPayload)

    const store = useDashboardStore()

    await store.fetchCurrentState('wallet-1')

    expect(store.summary).toBeNull()
    expect(store.allocation).toBeNull()
    expect(store.risk).toBeNull()
    expect(store.error).toBe('Request failed')
    expect(store.loading).toBe(false)
  })

  it('fetchEvolution maps the api payload into the evolution state', async () => {
    vi.mocked(dashboardApi.evolution).mockResolvedValue(evolutionPayload)

    const store = useDashboardStore()

    await store.fetchEvolution('wallet-1', 2026, 3)

    expect(store.evolution).toEqual(evolutionPayload)
    expect(dashboardApi.evolution).toHaveBeenCalledWith('wallet-1', 2026, 3)
    expect(store.error).toBeNull()
  })

  it('fetchEvolution flags error and keeps evolution untouched when the request fails', async () => {
    vi.mocked(dashboardApi.evolution).mockRejectedValue(new Error('Request failed'))

    const store = useDashboardStore()

    await store.fetchEvolution('wallet-1')

    expect(store.evolution).toBeNull()
    expect(store.error).toBe('Request failed')
  })
})
