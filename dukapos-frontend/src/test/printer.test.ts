import { describe, it, expect, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { usePrinter, formatReceipt } from '@/hooks/usePrinter'

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn().mockResolvedValue({ data: { data: [] } }),
    post: vi.fn().mockResolvedValue({ data: { success: true } })
  }
}))

describe('usePrinter Hook', () => {
  it('should initialize with default values', () => {
    const { result } = renderHook(() => usePrinter())
    
    expect(result.current.isPrinting).toBe(false)
    expect(result.current.isScanning).toBe(false)
    expect(result.current.connectedPrinter).toBeNull()
    expect(result.current.availablePrinters).toEqual([])
    expect(result.current.error).toBeNull()
    expect(result.current.isCapacitor).toBe(false)
  })

  it('should scan for network printers', async () => {
    const { result } = renderHook(() => usePrinter())
    
    let printers
    await act(async () => {
      printers = await result.current.scanForPrinters()
    })
    
    expect(printers).toBeDefined()
  })

  it('should format receipt correctly', () => {
    const receipt = formatReceipt({
      shopName: 'Test Shop',
      shopPhone: '+254712345678',
      items: [
        { name: 'Product 1', quantity: 2, price: 100, total: 200 },
        { name: 'Product 2', quantity: 1, price: 50, total: 50 }
      ],
      subtotal: 250,
      tax: 25,
      total: 275,
      paymentMethod: 'cash',
      receiptNumber: 'REC001',
      cashier: 'John'
    })
    
    expect(receipt).toContain('Test Shop')
    expect(receipt).toContain('REC001')
    expect(receipt).toContain('Product 1')
    expect(receipt).toContain('CASH')
    expect(receipt).toContain('John')
  })
})
