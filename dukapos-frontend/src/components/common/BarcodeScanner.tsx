import { useState, useRef } from 'react'
import { Capacitor } from '@capacitor/core'
import { Html5Qrcode } from 'html5-qrcode'

const isNative = typeof Capacitor !== 'undefined' && Capacitor.isNativePlatform()

interface BarcodeScannerModalProps {
  isOpen: boolean
  onClose: () => void
  onScan: (barcode: string) => void
}

export function BarcodeScannerModal({ isOpen, onClose, onScan }: BarcodeScannerModalProps) {
  const [error, setError] = useState<string | null>(null)
  const [manualInput, setManualInput] = useState('')
  const [isScanning, setIsScanning] = useState(false)
  const scannerRef = useRef<Html5Qrcode | null>(null)

  const startNativeScanner = async () => {
    if (isScanning) return
    
    setError(null)
    setIsScanning(true)

    try {
      // Dynamic import for native scanner to avoid issues on web
      const { BarcodeScanner } = await import('@capacitor-mlkit/barcode-scanning')
      
      const permission = await BarcodeScanner.checkPermissions()
      if (permission.camera !== 'granted') {
        const result = await BarcodeScanner.requestPermissions()
        if (result.camera !== 'granted') {
          setError('Camera permission denied')
          setIsScanning(false)
          return
        }
      }

      await BarcodeScanner.startScan({})

      // For native, we'll handle results via a different mechanism
      // For now, just show a message
      setIsScanning(false)
      setError('Native scanner started. Point camera at barcode.')
    } catch (err) {
      console.error('Native scanner error:', err)
      setError('Failed to start scanner on this device')
      setIsScanning(false)
    }
  }

  const startWebScanner = async () => {
    if (isScanning) return
    
    setError(null)
    setIsScanning(true)

    try {
      scannerRef.current = new Html5Qrcode('barcode-scanner')
      await scannerRef.current.start(
        { facingMode: 'environment' },
        {
          fps: 10,
          qrbox: { width: 250, height: 150 }
        },
        (decodedText) => {
          onScan(decodedText)
          onClose()
        },
        () => {}
      )
    } catch (err) {
      console.error('Web scanner error:', err)
      setError('Camera not available. Please allow camera access or enter barcode manually.')
      setIsScanning(false)
    }
  }

  const stopScanning = async () => {
    try {
      if (scannerRef.current?.isScanning) {
        await scannerRef.current.stop()
      }
    } catch (err) {
      console.error('Error stopping scanner:', err)
    }
    setIsScanning(false)
  }

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (manualInput.trim()) {
      onScan(manualInput.trim())
      onClose()
      setManualInput('')
    }
  }

  const handleOpen = () => {
    if (isNative) {
      startNativeScanner()
    } else {
      startWebScanner()
    }
  }

  const handleClose = () => {
    stopScanning()
    setError(null)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/80 z-50 flex items-center justify-center">
      <div className="w-full max-w-md p-4">
        <div className="bg-white rounded-2xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold">Scan Barcode</h3>
            <button onClick={handleClose} className="p-2">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          
          {error ? (
            <div className="bg-amber-50 text-amber-700 p-4 rounded-xl mb-4">
              <p className="text-sm">{error}</p>
              <button
                onClick={handleOpen}
                className="mt-2 text-sm font-medium underline"
              >
                Try Again
              </button>
            </div>
          ) : !isScanning ? (
            <div className="text-center py-4">
              <button
                onClick={handleOpen}
                className="px-6 py-3 bg-primary text-white rounded-xl font-medium"
              >
                Start Scanner
              </button>
            </div>
          ) : !isNative ? (
            <div id="barcode-scanner" className="w-full h-64 rounded-xl overflow-hidden bg-black mb-4"></div>
          ) : null}
          
          <form onSubmit={handleManualSubmit}>
            <input
              type="text"
              placeholder="Or enter barcode manually..."
              value={manualInput}
              onChange={(e) => setManualInput(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              autoFocus
            />
            <button
              type="submit"
              className="w-full mt-3 px-4 py-3 bg-primary text-white rounded-xl font-medium"
              disabled={!manualInput.trim()}
            >
              Submit
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}

interface BarcodeScannerButtonProps {
  onScan: (barcode: string) => void
  className?: string
}

export function BarcodeScannerButton({ onScan, className = '' }: BarcodeScannerButtonProps) {
  const [showModal, setShowModal] = useState(false)

  return (
    <>
      <button
        onClick={() => setShowModal(true)}
        className={`p-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all ${className}`}
        title="Scan Barcode"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
        </svg>
      </button>
      <BarcodeScannerModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onScan={onScan}
      />
    </>
  )
}
