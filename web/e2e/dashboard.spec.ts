import { test, expect } from '@playwright/test'
import type { Page } from '@playwright/test'

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

function canvasContainsColor(page: Page, containerTestId: string, rgb: readonly [number, number, number]) {
  return page
    .locator(`[data-testid="${containerTestId}"] canvas`)
    .evaluate((canvas, color) => {
      const canvasElement = canvas as HTMLCanvasElement
      const context = canvasElement.getContext('2d')
      if (!context) {
        return false
      }
      const { data } = context.getImageData(0, 0, canvasElement.width, canvasElement.height)
      for (let i = 0; i < data.length; i += 4) {
        if (data[i] === color[0] && data[i + 1] === color[1] && data[i + 2] === color[2]) {
          return true
        }
      }
      return false
    }, rgb)
}

function riskGaugeShape(page: Page, containerTestId: string, rgbs: ReadonlyArray<readonly [number, number, number]>) {
  return page
    .locator(`[data-testid="${containerTestId}"] canvas`)
    .evaluate((canvas, colors) => {
      const canvasElement = canvas as HTMLCanvasElement
      const context = canvasElement.getContext('2d')
      if (!context) {
        return null
      }
      const { width, height } = canvasElement
      const image = context.getImageData(0, 0, width, height).data
      let minY = height
      let maxY = -1
      let minX = width
      let maxX = -1
      for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
          const index = (y * width + x) * 4
          for (const color of colors) {
            if (image[index] === color[0] && image[index + 1] === color[1] && image[index + 2] === color[2]) {
              if (y < minY) minY = y
              if (y > maxY) maxY = y
              if (x < minX) minX = x
              if (x > maxX) maxX = x
              break
            }
          }
        }
      }
      const yBottomStart = Math.floor(height * 0.85)
      const bottomImage = context.getImageData(0, yBottomStart, width, height - yBottomStart).data
      let bottomHasColors = false
      for (let i = 0; i < bottomImage.length && !bottomHasColors; i += 4) {
        for (const color of colors) {
          if (bottomImage[i] === color[0] && bottomImage[i + 1] === color[1] && bottomImage[i + 2] === color[2]) {
            bottomHasColors = true
            break
          }
        }
      }
      return {
        heightOverWidth: maxY < 0 ? 1 : (maxY - minY) / Math.max(1, maxX - minX),
        bottomHasColors,
      }
    }, rgbs.map((rgb) => [...rgb]))
}

const RISK_RANK_COLOR = {
  red: [220, 38, 38],
  amber: [245, 158, 11],
  yellow: [234, 179, 8],
  cyan: [8, 145, 178],
  green: [22, 163, 74],
} as const

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
      'Gerenciamento de Risco (notas)',
    ])
  })

  test('keeps the three composition cards equally sized with a large unclipped gauge', async ({ page, request }) => {
    const walletId = await createWallet(request)
    await createPatrimony(request, walletId, 'stocks', 500000)
    await createPatrimony(request, walletId, 'fixed_income', 300000)
    await createPosition(request, walletId, 'PETR4')

    await page.goto('/')

    const cards = page.locator('.composition__card')
    await expect(cards).toHaveCount(3)

    await expect.poll(async () => {
      const boxes = []
      for (let index = 0; index < 3; index++) {
        const box = await cards.nth(index).boundingBox()
        if (!box) {
          return false
        }
        boxes.push(box)
      }
      const heights = boxes.map((box) => box.height)
      const widths = boxes.map((box) => box.width)
      const heightSpread = Math.max(...heights) - Math.min(...heights)
      const widthSpread = Math.max(...widths) - Math.min(...widths)
      return heightSpread <= 2 && widthSpread <= 2
    }).toBe(true)

    const cardBox = await page.locator('[data-testid="risk-chart"]').boundingBox()
    const gaugeBox = await page.locator('[data-testid="risk-chart"] .composition__chart').boundingBox()
    const allocationCanvasBox = await page.locator('[data-testid="allocation-chart"] canvas').boundingBox()
    expect(cardBox).not.toBeNull()
    expect(gaugeBox).not.toBeNull()
    expect(allocationCanvasBox).not.toBeNull()

    expect(Math.abs(gaugeBox!.height - gaugeBox!.width)).toBeLessThanOrEqual(2)

    expect(gaugeBox!.width).toBeGreaterThan(allocationCanvasBox!.width / 2)

    expect(gaugeBox!.y).toBeGreaterThanOrEqual(cardBox!.y)
    expect(gaugeBox!.y + gaugeBox!.height).toBeLessThanOrEqual(cardBox!.y + cardBox!.height + 1)
  })

  test('draws the risk half-circle gauge with per rank colors and a right side legend', async ({ page, request }) => {
    await createWallet(request)

    await page.route('**/api/v1/wallets/*/dashboard/risk', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            items: [
              { rank: 3, amount: 4620000, percentage: 62 },
              { rank: 4, amount: 1940000, percentage: 26 },
              { rank: 5, amount: 920000, percentage: 12 },
            ],
            total: 7480000,
          },
        }),
      })
    })

    await page.goto('/')

    const riskCanvas = page.locator('[data-testid="risk-chart"] canvas')
    await expect(riskCanvas).toBeVisible()

    const gaugeColors = [
      RISK_RANK_COLOR.yellow,
      RISK_RANK_COLOR.cyan,
      RISK_RANK_COLOR.green,
    ] as const
    await expect.poll(() =>
      canvasContainsColor(page, 'risk-chart', RISK_RANK_COLOR.yellow),
    ).toBe(true)
    await expect.poll(() =>
      canvasContainsColor(page, 'risk-chart', RISK_RANK_COLOR.cyan),
    ).toBe(true)
    await expect.poll(() =>
      canvasContainsColor(page, 'risk-chart', RISK_RANK_COLOR.green),
    ).toBe(true)

    await expect.poll(async () => {
      const shape = await riskGaugeShape(page, 'risk-chart', gaugeColors)
      return shape !== null && shape.heightOverWidth <= 0.75 && !shape.bottomHasColors
    }).toBe(true)

    const legend = page.locator('[data-testid="risk-legend"]')
    await expect(legend).toBeVisible()
    await expect(legend.locator('.composition__risk-legend-item')).toHaveText([
      'Boa (3)',
      'Ótimo (4)',
      'Excelente (5)',
    ])

    const canvasBox = await riskCanvas.boundingBox()
    const legendBox = await legend.boundingBox()
    expect(canvasBox).not.toBeNull()
    expect(legendBox).not.toBeNull()
    expect(legendBox!.x).toBeGreaterThanOrEqual(canvasBox!.x + canvasBox!.width - 2)
  })

  test('draws the high risk slice in red for a rank 1 portfolio', async ({ page, request }) => {
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

    await expect(page.locator('[data-testid="risk-chart"] canvas')).toBeVisible()

    await expect.poll(() =>
      canvasContainsColor(page, 'risk-chart', RISK_RANK_COLOR.red),
    ).toBe(true)
  })

  test('renders an empty risk chart when the wallet has no positions', async ({ page, request }) => {
    await createWallet(request)

    await page.goto('/')

    await expect(page.locator('[data-testid="risk-chart"] canvas')).toBeVisible()
    await expect(page.locator('[data-testid="risk-legend"]')).toHaveCount(0)
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
