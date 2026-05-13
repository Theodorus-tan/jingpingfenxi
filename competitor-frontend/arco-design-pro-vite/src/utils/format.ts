export function formatPrice(price: number, currency: string = 'CNY'): string {
  if (price < 0) return 'N/A'
  if (price === 0) return '免费'

  const symbols: Record<string, string> = {
    CNY: '¥', USD: '$', EUR: '€', JPY: '¥',
  }

  const symbol = symbols[currency] || currency
  const parts = price.toFixed(2).split('.')
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')

  return `${symbol}${parts.join('.')}`
}

export function truncateText(text: string, maxLength: number): string {
  if (!text || text.length <= maxLength) return text
  return text.slice(0, maxLength) + '...'
}

export function calcScoreColor(score: number): 'green' | 'yellow' | 'red' {
  if (score >= 8) return 'green'
  if (score >= 5) return 'yellow'
  return 'red'
}

export function calcPercentage(part: number, total: number): string {
  if (total === 0) return '0%'
  return ((part / total) * 100).toFixed(1) + '%'
}
