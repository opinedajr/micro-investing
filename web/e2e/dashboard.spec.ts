import { test, expect } from '@playwright/test'

interface WalletResponse {
  data: {
    id: string
    name: string
  }
}

interface StockResponse {
  data: {
    id: string
    ticker: string
  }
}

async function createWallet(request: ReturnType<typeof test['request']['newContext']> extends Promise<infer T> ? T : never): Promise<string> {
  const response = await request.post('/api/v1/wallets', {
    data: { name: 'E2E KPI Wallet' },
  })
  expect(response.ok(), `createWallet failed: ${await response.text()}`).toBeTruthy()
  const body = (await response.json()) as WalletResponse
  return body.data.id
}

async function createPatrimony(request: ReturnType<typeof test['request']['newContext']> extends Promise<infer T> ? T : never, walletId: string, type: string, amount: number): Promise<void> {
  const response = await request.post(`/api/v1/wallets/${walletId}/patrimonies`, {
    data: { year: 2026, month: 9, type, amount },
  })
  expect(response.ok(), `createPatrimony failed: ${await response.text()}`).toBeTruthy()
}

async function stockIdByTicker(request: ReturnType<typeof test['request']['newContext']> extends Promise<infer T> ? T : never, ticker: string): Promise<string> {
  const response = await request.get(`/api/v1/stocks/${ticker}`)
  expect(response.ok(), `stockIdByTicker failed: ${await response.text()}`).toBeTruthy()
  const body = (await response.json()) as StockResponse
  return body.data.id
}

async function createPosition(request: ReturnType<typeof test['request']['newContext']> extends Promise<infer T> ? T : never, walletId: string, ticker: string): Promise<void> {
  const stockId = await stockIdByTicker(request, ticker)
  const response = await request.post(`/api/v1/wallets/${walletId}/positions`, {
    data: { stock_id: stockId, quantity: 100, average_price: 5000 },
  })
  expect(response.ok(), `createPosition failed: ${await response.text()}`).toBeTruthy()
}

test.describe('Dashboard KPI Cards', () => {
  test.beforeEach(async ({ request }) => {
    const walletsResponse = await request.get('/api/v1/wallets')
    const wallets = (await walletsResponse.json() as { data: Array<{ id: string }> }).data
    for (const wallet of wallets) {
      await request.delete(`/api/v1/wallets/${wallet.id}`)
    }
  })

  test('loads the dashboard page without errors', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/Micro Investing/)
    await expect(page.locator('[data-testid="dashboard-kpi-cards"]')).toBeVisible()
  })

  test('renders three KPI cards with BRL formatted values', async ({ page, request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 150000)
    await createPatrimony(request, walletId, 'fixed_income', 250000)
    await createPosition(request, walletId, 'PETR4')

    await page.goto('/')

    await expect(page.locator('.kpi-card--patrimony')).toContainText('Patrimônio')
    await expect(page.locator('.kpi-card--patrimony')).toContainText('R$ 4.000,00')

    await expect(page.locator('.kpi-card--dividends')).toContainText('Dividendo')
    await expect(page.locator('.kpi-card--dividends')).toContainText('R$ 0,00')

    await expect(page.locator('.kpi-card--stocks')).toContainText('Ações')
    await expect(page.locator('.kpi-card--stocks')).toContainText('R$ 5.000,00')
  })

  test('reacts to store changes when summary data is updated', async ({ page, request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 100000)

    await page.goto('/')
    await expect(page.locator('.kpi-card--patrimony')).toContainText('R$ 1.000,00')

    await createPatrimony(request, walletId, 'fixed_income', 50000)
    await page.reload()
    await expect(page.locator('.kpi-card--patrimony')).toContainText('R$ 1.500,00')
  })

  test('handles empty wallet with zero formatted values', async ({ page, request }) => {
    await createWallet(request)

    await page.goto('/')

    await expect(page.locator('.kpi-card--patrimony')).toContainText('R$ 0,00')
    await expect(page.locator('.kpi-card--dividends')).toContainText('R$ 0,00')
    await expect(page.locator('.kpi-card--stocks')).toContainText('R$ 0,00')
  })

  test('handles network error by exposing error state', async ({ page, request }) => {
    const walletId = await createWallet(request)

    await page.route(`/api/v1/wallets/${walletId}/dashboard/summary`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ data: null, error: { code: 'INTERNAL_ERROR', message: 'Internal server error' } }),
      })
    })

    await page.goto('/')

    await expect(page.locator('[data-testid="dashboard-kpi-cards"]')).toBeVisible()
    await expect(page.locator('.kpi-card--patrimony')).toContainText('R$ 0,00')
  })
})

test.describe('Dashboard Composition Charts', () => {
  test.beforeEach(async ({ request }) => {
    const walletsResponse = await request.get('/api/v1/wallets')
    const wallets = (await walletsResponse.json() as { data: Array<{ id: string }> }).data
    for (const wallet of wallets) {
      await request.delete(`/api/v1/wallets/${wallet.id}`)
    }
  })

  test('renders composition charts with percentage labels', async ({ page, request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 500000)
    await createPatrimony(request, walletId, 'fixed_income', 300000)
    await createPatrimony(request, walletId, 'emergency_reserve', 200000)
    await createPosition(request, walletId, 'PETR4')

    await page.goto('/')

    const allocationCanvas = page.locator('[data-testid="allocation-chart"] canvas')
    const riskCanvas = page.locator('[data-testid="risk-chart"] canvas')

    await expect(allocationCanvas).toBeVisible()
    await expect(riskCanvas).toBeVisible()
    await expect(page.locator('[data-testid="risk-legend"]')).toBeVisible()
  })

  test('renders the dividends card between allocation and risk cards', async ({ page, request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 500000)
    await createPosition(request, walletId, 'PETR4')

    await page.goto('/')

    const titles = page.locator('.composition__card .composition__title')
    await expect(titles).toHaveText([
      'Alocação de Patrimônio',
      'Dividendos',
      'Gerenciamento de Risco',
    ])
  })

  test('shows the dynamic risk legend for an optimal score', async ({ page, request }) => {
    await createWallet(request)

    await page.route('**/api/v1/wallets/*/dashboard/risk', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            items: [
              { rank: 4, amount: 800000, percentage: 80 },
              { rank: 5, amount: 200000, percentage: 20 },
            ],
            total: 1000000,
          },
        }),
      })
    })

    await page.goto('/')

    await expect(page.locator('[data-testid="risk-legend"]')).toHaveText('Ótima')
  })

  test('shows the dynamic risk legend for the worst score', async ({ page, request }) => {
    await createWallet(request)

    await page.route('**/api/v1/wallets/*/dashboard/risk', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            items: [{ rank: 1, amount: 1000000, percentage: 100 }],
            total: 1000000,
          },
        }),
      })
    })

    await page.goto('/')

    await expect(page.locator('[data-testid="risk-legend"]')).toHaveText('Alto Risco')
  })

  test('shows an empty state legend when the wallet has no positions', async ({ page, request }) => {
    await createWallet(request)

    await page.goto('/')

    await expect(page.locator('[data-testid="risk-legend"]')).toHaveText('Sem pontuação')
  })

  test('returns allocation and risk percentages as integers', async ({ request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 200000)
    await createPatrimony(request, walletId, 'fixed_income', 150000)
    await createPatrimony(request, walletId, 'emergency_reserve', 50000)
    await createPosition(request, walletId, 'PETR4')

    const allocationResponse = await request.get(`/api/v1/wallets/${walletId}/dashboard/allocation`)
    expect(allocationResponse.ok()).toBeTruthy()
    const allocationBody = (await allocationResponse.json()) as {
      data: { items: Array<{ type: string; percentage: number }> }
    }
    const percentageByType = Object.fromEntries(
      allocationBody.data.items.map((item) => [item.type, item.percentage]),
    )
    expect(Number.isInteger(percentageByType.stocks)).toBe(true)
    expect(Number.isInteger(percentageByType.fixed_income)).toBe(true)
    expect(Number.isInteger(percentageByType.emergency_reserve)).toBe(true)
    expect(percentageByType).toEqual({
      stocks: 50,
      fixed_income: 38,
      emergency_reserve: 13,
    })

    const riskResponse = await request.get(`/api/v1/wallets/${walletId}/dashboard/risk`)
    expect(riskResponse.ok()).toBeTruthy()
    const riskBody = (await riskResponse.json()) as {
      data: { items: Array<{ rank: number; percentage: number }> }
    }
    expect(riskBody.data.items).toHaveLength(1)
    expect(riskBody.data.items[0].rank).toBe(10)
    expect(riskBody.data.items[0].percentage).toBe(100)
  })
})
