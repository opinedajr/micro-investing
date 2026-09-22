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
