import { describe, expect, it } from 'vitest'

import { formatCurrencyBRL } from './utils'

describe('formatCurrencyBRL', () => {
  it('formats cents amount as brazilian currency', () => {
    expect(formatCurrencyBRL(1050)).toBe('R$ 10,50')
  })

  it('formats zero as brazilian currency', () => {
    expect(formatCurrencyBRL(0)).toBe('R$ 0,00')
  })

  it('formats thousands with brazilian separators', () => {
    expect(formatCurrencyBRL(150000)).toBe('R$ 1.500,00')
    expect(formatCurrencyBRL(123456789)).toBe('R$ 1.234.567,89')
  })

  it('formats negative amounts', () => {
    expect(formatCurrencyBRL(-1050)).toBe('-R$ 10,50')
  })
})
