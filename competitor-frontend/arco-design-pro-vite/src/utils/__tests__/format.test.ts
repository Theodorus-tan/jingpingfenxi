import { describe, it, expect } from 'vitest'

function formatPrice(price: number, currency: string = 'CNY'): string {
  if (price < 0) return 'N/A'
  if (price === 0) return '免费'
  const symbols: Record<string, string> = { CNY: '¥', USD: '$', EUR: '€', JPY: '¥' }
  const symbol = symbols[currency] || currency
  const parts = price.toFixed(2).split('.')
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return `${symbol}${parts.join('.')}`
}

function truncateText(text: string, maxLength: number): string {
  if (!text || text.length <= maxLength) return text
  return text.slice(0, maxLength) + '...'
}

function calcScoreColor(score: number): 'green' | 'yellow' | 'red' {
  if (score >= 8) return 'green'
  if (score >= 5) return 'yellow'
  return 'red'
}

function calcPercentage(part: number, total: number): string {
  if (total === 0) return '0%'
  return ((part / total) * 100).toFixed(1) + '%'
}

describe('formatPrice', () => {
  it('整数价格返回带货币符号格式', () => {
    expect(formatPrice(2599)).toBe('¥2,599.00')
    expect(formatPrice(2599, 'USD')).toBe('$2,599.00')
  })

  it('小数价格保留两位', () => {
    expect(formatPrice(1299.5)).toBe('¥1,299.50')
  })

  it('0 元返回 "免费"', () => {
    expect(formatPrice(0)).toBe('免费')
  })

  it('负数返回 "N/A"', () => {
    expect(formatPrice(-1)).toBe('N/A')
  })
})

describe('truncateText', () => {
  it('短于限制不截断', () => {
    expect(truncateText('Hello', 10)).toBe('Hello')
  })

  it('长于限制添加省略号', () => {
    expect(truncateText('这是一段非常长的文本', 5)).toBe('这是一段非...')
  })

  it('空字符串返回空', () => {
    expect(truncateText('', 5)).toBe('')
  })
})

describe('calcScoreColor', () => {
  it('>= 8 分返回 green', () => {
    expect(calcScoreColor(8.5)).toBe('green')
    expect(calcScoreColor(10)).toBe('green')
  })

  it('5~7 分返回 yellow', () => {
    expect(calcScoreColor(7)).toBe('yellow')
    expect(calcScoreColor(5)).toBe('yellow')
  })

  it('< 5 分返回 red', () => {
    expect(calcScoreColor(3)).toBe('red')
  })
})

describe('calcPercentage', () => {
  it('正常计算百分比', () => {
    expect(calcPercentage(75, 100)).toBe('75.0%')
    expect(calcPercentage(1, 3)).toBe('33.3%')
  })

  it('总数为 0 时返回 0%', () => {
    expect(calcPercentage(10, 0)).toBe('0%')
  })
})
