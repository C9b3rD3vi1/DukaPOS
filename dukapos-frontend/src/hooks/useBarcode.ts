import { useState, useEffect, useCallback } from 'react'
import { Capacitor } from '@capacitor/core'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
let BarcodeScanner: any = null

try {
  BarcodeScanner = require('@capacitor-mlkit/barcode-scanning')
} catch {
  console.warn('Barcode scanner not available in web environment')
}

interface UseBarcodeScannerOptions {
  onScan?: (barcode: string) => void
  onError?: (error: string) => void
  formats?: string[]
}

interface UseBarcodeScannerReturn {
  isSupported: boolean
  isScanning: boolean
  startScanning: () => Promise<void>
  stopScanning: () => Promise<void>
  error: string | null
}

export function useBarcodeScanner(options: UseBarcodeScannerOptions = {}): UseBarcodeScannerReturn {
  const { onScan, onError, formats = ['qr_code', 'ean_13', 'ean_8', 'code_128', 'code_39'] } = options
  
  const [isSupported, setIsSupported] = useState(false)
  const [isScanning, setIsScanning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const checkSupport = async () => {
      try {
        if (!Capacitor.isNativePlatform() || !BarcodeScanner) {
          setIsSupported(false)
          return
        }
        
        const supported = await BarcodeScanner.isSupported()
        setIsSupported(supported.supported)
      } catch (err) {
        console.error('Barcode scanner support check failed:', err)
        setIsSupported(false)
      }
    }

    checkSupport()
  }, [])

  const startScanning = useCallback(async () => {
    if (!isSupported) {
      const errMsg = 'Barcode scanning is not supported on this device'
      setError(errMsg)
      onError?.(errMsg)
      return
    }

    setError(null)
    setIsScanning(true)

    try {
      const permission = await BarcodeScanner.checkPermissions()
      if (permission.camera !== 'granted') {
        const result = await BarcodeScanner.requestPermissions()
        if (result.camera !== 'granted') {
          const errMsg = 'Camera permission denied'
          setError(errMsg)
          onError?.(errMsg)
          setIsScanning(false)
          return
        }
      }

      await BarcodeScanner.startScan({
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        formats: formats as any
      })

      setIsScanning(true)
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Failed to start scanning'
      setError(errMsg)
      onError?.(errMsg)
      setIsScanning(false)
    }
  }, [isSupported, formats, onScan, onError])

  const stopScanning = useCallback(async () => {
    try {
      await BarcodeScanner.stopScan()
    } catch (err) {
      console.error('Failed to stop scanning:', err)
    } finally {
      setIsScanning(false)
    }
  }, [])

  useEffect(() => {
    return () => {
      BarcodeScanner.removeAllListeners()
    }
  }, [])

  return {
    isSupported,
    isScanning,
    startScanning,
    stopScanning,
    error
  }
}

export function useCamera() {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null)
  const [error, setError] = useState<string | null>(null)

  const requestPermission = useCallback(async () => {
    try {
      const result = await BarcodeScanner.requestPermissions()
      setHasPermission(result.camera === 'granted')
      return result.camera === 'granted'
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Permission request failed'
      setError(errMsg)
      return false
    }
  }, [])

  const checkPermission = useCallback(async () => {
    try {
      const result = await BarcodeScanner.checkPermissions()
      setHasPermission(result.camera === 'granted')
      return result.camera === 'granted'
    } catch (err) {
      return false
    }
  }, [])

  return {
    hasPermission,
    error,
    requestPermission,
    checkPermission
  }
}
