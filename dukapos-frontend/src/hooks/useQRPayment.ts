import { useState, useCallback } from 'react'

export interface QRPaymentData {
  type: 'mpesa' | 'bank' | 'cash'
  amount?: number
  phone?: string
  reference?: string
}

export interface ScanResult {
  displayValue: string
  rawValue: string
  format: string
}

export function useQRPayment() {
  const [isScanning, setIsScanning] = useState(false)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [_lastScan, _setLastScan] = useState<ScanResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const isCapacitor = false

  const parseQRCode = useCallback((rawValue: string): QRPaymentData | null => {
    try {
      if (rawValue.startsWith('{')) {
        const data = JSON.parse(rawValue)
        
        if (data.t === 'mpesa' || data.type === 'mpesa') {
          return {
            type: 'mpesa',
            amount: data.amount || data.a,
            phone: data.phone || data.p,
            reference: data.reference || data.r || data.tx
          }
        }
        
        if (data.t === 'bank' || data.type === 'bank') {
          return {
            type: 'bank',
            amount: data.amount || data.a,
            reference: data.reference || data.r
          }
        }
        
        return data
      }
      
      if (rawValue.startsWith('dukapos://') || rawValue.startsWith('https://dukapos.com/pay')) {
        const url = new URL(rawValue)
        return {
          type: (url.searchParams.get('type') as 'mpesa' | 'bank') || 'mpesa',
          amount: Number(url.searchParams.get('amount')) || undefined,
          phone: url.searchParams.get('phone') || undefined,
          reference: url.searchParams.get('ref') || undefined
        }
      }
      
      if (/^2547\d{8}$/.test(rawValue) || /^07\d{8}$/.test(rawValue)) {
        return {
          type: 'mpesa',
          phone: rawValue
        }
      }
      
      return null
    } catch {
      return null
    }
  }, [])

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const startScanning = useCallback(async (_onScan: (data: QRPaymentData) => void): Promise<void> => {
    setIsScanning(true)
    setError(null)

    try {
      setError('Camera scanning requires the native mobile app')
      setIsScanning(false)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to start scanner')
      setIsScanning(false)
    }
  }, [isCapacitor, parseQRCode])

  const stopScanning = useCallback(async (): Promise<void> => {
    setIsScanning(false)
  }, [isCapacitor])

  const generatePaymentQR = useCallback((data: {
    amount: number
    phone?: string
    reference?: string
    type?: 'mpesa' | 'bank'
  }): string => {
    const payload = {
      t: data.type || 'mpesa',
      a: data.amount,
      p: data.phone,
      r: data.reference || `DUKA${Date.now()}`,
      ts: new Date().toISOString()
    }
    
    return JSON.stringify(payload)
  }, [])

  return {
    isScanning,
    lastScan: _lastScan,
    error,
    startScanning,
    stopScanning,
    parseQRCode,
    generatePaymentQR,
    isCapacitor
  }
}
