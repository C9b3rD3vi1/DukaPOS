import { describe, it, expect } from 'vitest'
import { formatCurrency, formatNumber, formatPercent, formatCompactNumber } from '@/utils/format'

describe('formatCurrency', () => {
  it('formats KES currency correctly', () => {
    expect(formatCurrency(1000)).toContain('1,000')
    expect(formatCurrency(1000000)).toContain('1,000,000')
  })

  it('handles zero', () => {
    expect(formatCurrency(0)).toContain('0')
  })
})

describe('formatNumber', () => {
  it('formats numbers with commas', () => {
    expect(formatNumber(1000)).toBe('1,000')
    expect(formatNumber(1000000)).toBe('1,000,000')
  })
})

describe('formatPercent', () => {
  it('formats percentage correctly', () => {
    expect(formatPercent(50)).toBe('50.0%')
    expect(formatPercent(33.333, 2)).toBe('33.33%')
  })
})

describe('formatCompactNumber', () => {
  it('formats large numbers compactly', () => {
    expect(formatCompactNumber(1500000)).toBe('1.5M')
    expect(formatCompactNumber(2500)).toBe('2.5K')
    expect(formatCompactNumber(500)).toBe('500')
  })
})
