import { defineStore } from 'pinia'

import {
  dashboardApi,
  type DashboardAllocation,
  type DashboardEvolution,
  type DashboardRisk,
  type DashboardSummary,
} from '@/lib/api'

interface DashboardState {
  summary: DashboardSummary | null
  allocation: DashboardAllocation | null
  risk: DashboardRisk | null
  evolution: DashboardEvolution | null
  loading: boolean
  error: string | null
}

export const useDashboardStore = defineStore('dashboard', {
  state: (): DashboardState => ({
    summary: null,
    allocation: null,
    risk: null,
    evolution: null,
    loading: false,
    error: null,
  }),
  actions: {
    async fetchCurrentState(walletId: string) {
      this.loading = true
      this.error = null
      try {
        const [summary, allocation, risk] = await Promise.all([
          dashboardApi.summary(walletId),
          dashboardApi.allocation(walletId),
          dashboardApi.risk(walletId),
        ])
        this.summary = summary
        this.allocation = allocation
        this.risk = risk
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Unexpected error'
      } finally {
        this.loading = false
      }
    },
    async fetchEvolution(walletId: string, year?: number, quarter?: number) {
      this.error = null
      try {
        this.evolution = (await dashboardApi.evolution(walletId, year, quarter)) ?? null
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Unexpected error'
      }
    },
  },
})
