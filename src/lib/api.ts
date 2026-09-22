export interface Wallet {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface DashboardSummary {
  current_patrimony: number
  yearly_dividends: number
  stocks_invested: number
}

export interface AllocationItem {
  type: string
  amount: number
  percentage: number
}

export interface DashboardAllocation {
  items: AllocationItem[]
  total: number
}

export interface RiskItem {
  rank: number
  amount: number
  percentage: number
}

export interface DashboardRisk {
  items: RiskItem[]
  total: number
}

export interface EvolutionMonth {
  year: number
  month: number
  amount: number
}

export interface EvolutionByCategory {
  fixed_income: EvolutionMonth[]
  stocks: EvolutionMonth[]
  emergency_reserve: EvolutionMonth[]
}

export interface DashboardEvolution {
  total: EvolutionMonth[]
  by_category: EvolutionByCategory
}

const API_BASE_URL = '/api/v1'

async function request<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`)

  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new Error(body?.error?.message ?? `Request failed with status ${response.status}`)
  }

  const body: { data: T } = await response.json()
  return body.data
}

export const walletApi = {
  list(): Promise<Wallet[]> {
    return request<Wallet[]>('/wallets')
  },
}

export const dashboardApi = {
  summary(walletId: string): Promise<DashboardSummary> {
    return request<DashboardSummary>(`/wallets/${walletId}/dashboard/summary`)
  },
  allocation(walletId: string): Promise<DashboardAllocation> {
    return request<DashboardAllocation>(`/wallets/${walletId}/dashboard/allocation`)
  },
  risk(walletId: string): Promise<DashboardRisk> {
    return request<DashboardRisk>(`/wallets/${walletId}/dashboard/risk`)
  },
  evolution(walletId: string, year?: number, quarter?: number): Promise<DashboardEvolution> {
    const params = new URLSearchParams()
    if (year !== undefined) {
      params.set('year', String(year))
    }
    if (quarter !== undefined) {
      params.set('quarter', String(quarter))
    }
    const query = params.toString()
    return request<DashboardEvolution>(`/wallets/${walletId}/dashboard/evolution${query ? `?${query}` : ''}`)
  },
}
