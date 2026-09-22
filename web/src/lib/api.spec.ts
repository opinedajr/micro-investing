import { afterEach, describe, expect, it, vi } from 'vitest'

import { dashboardApi, walletApi } from './api'

const fetchMock = vi.fn()

vi.stubGlobal('fetch', fetchMock)

afterEach(() => {
  fetchMock.mockReset()
})

describe('api client', () => {
  it('unwraps the data envelope on success', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { current_patrimony: 1050 } }), { status: 200 }),
    )

    const summary = await dashboardApi.summary('wallet-1')

    expect(summary).toEqual({ current_patrimony: 1050 })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/wallets/wallet-1/dashboard/summary')
  })

  it('builds evolution query params', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { total: [], by_category: { fixed_income: [], stocks: [], emergency_reserve: [] } } }), { status: 200 }),
    )

    await dashboardApi.evolution('wallet-1', 2026, 2)

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/wallets/wallet-1/dashboard/evolution?year=2026&quarter=2')
  })

  it('calls evolution without query params when period is not provided', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { total: [], by_category: { fixed_income: [], stocks: [], emergency_reserve: [] } } }), { status: 200 }),
    )

    await dashboardApi.evolution('wallet-1')

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/wallets/wallet-1/dashboard/evolution')
  })

  it('throws the api error message on failure', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ error: { code: 'WALLET_NOT_FOUND', message: 'Wallet not found' } }), { status: 404 }),
    )

    await expect(dashboardApi.summary('missing')).rejects.toThrow('Wallet not found')
  })

  it('lists wallets', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ data: [{ id: 'wallet-1', name: 'Main' }] }), { status: 200 }),
    )

    const wallets = await walletApi.list()

    expect(wallets).toEqual([{ id: 'wallet-1', name: 'Main' }])
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/wallets')
  })
})
