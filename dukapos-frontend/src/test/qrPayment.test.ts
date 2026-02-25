import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useQRPayment } from '@/hooks/useQRPayment'

describe('useQRPayment Hook', () => {
  it('should initialize with default values', () => {
    const { result } = renderHook(() => useQRPayment())
    
    expect(result.current.isScanning).toBe(false)
    expect(result.current.error).toBeNull()
    expect(result.current.isCapacitor).toBe(false)
  })

  it('should parse QR code with JSON format', async () => {
    const { result } = renderHook(() => useQRPayment())
    
    const mpesaQR = JSON.stringify({
      t: 'mpesa',
      a: 100,
      p: '254712345678',
      r: 'TX123'
    })
    
    const parsed = result.current.parseQRCode(mpesaQR)
    
    expect(parsed).toEqual({
      type: 'mpesa',
      amount: 100,
      phone: '254712345678',
      reference: 'TX123'
    })
  })

  it('should parse bank QR code', async () => {
    const { result } = renderHook(() => useQRPayment())
    
    const bankQR = JSON.stringify({
      t: 'bank',
      a: 500,
      r: 'INV001'
    })
    
    const parsed = result.current.parseQRCode(bankQR)
    
    expect(parsed).toEqual({
      type: 'bank',
      amount: 500,
      reference: 'INV001'
    })
  })

  it('should parse URL format QR code', async () => {
    const { result } = renderHook(() => useQRPayment())
    
    const urlQR = 'dukapos://pay?amount=100&phone=254712345678&type=mpesa'
    
    const parsed = result.current.parseQRCode(urlQR)
    
    expect(parsed).toEqual({
      type: 'mpesa',
      amount: 100,
      phone: '254712345678',
      reference: undefined
    })
  })

  it('should parse phone number as M-Pesa', async () => {
    const { result } = renderHook(() => useQRPayment())
    
    expect(result.current.parseQRCode('254712345678')).toEqual({
      type: 'mpesa',
      phone: '254712345678'
    })
    
    expect(result.current.parseQRCode('0712345678')).toEqual({
      type: 'mpesa',
      phone: '0712345678'
    })
  })

  it('should generate payment QR code', () => {
    const { result } = renderHook(() => useQRPayment())
    
    const qr = result.current.generatePaymentQR({
      amount: 100,
      phone: '254712345678',
      type: 'mpesa'
    })
    
    const parsed = JSON.parse(qr)
    expect(parsed.t).toBe('mpesa')
    expect(parsed.a).toBe(100)
    expect(parsed.p).toBe('254712345678')
  })
})
