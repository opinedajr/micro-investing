export function formatCurrencyBRL(cents: number): string {
  const formatted = new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(cents / 100)

  return formatted.replace(/[\u00A0\u202F]/g, ' ')
}
