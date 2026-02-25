import { useState, useCallback } from 'react'

export interface PrinterDevice {
  id: string
  name: string
  type: 'bluetooth' | 'network' | 'usb'
  address?: string
  connected?: boolean
}

export interface PrintJob {
  id: string
  content: string
  status: 'pending' | 'printing' | 'completed' | 'failed'
  timestamp: Date
}

export interface BluetoothPrinter {
  deviceId: string
  name: string
}

export function usePrinter() {
  const [isPrinting, setIsPrinting] = useState(false)
  const [isScanning, setIsScanning] = useState(false)
  const [connectedPrinter, setConnectedPrinter] = useState<PrinterDevice | null>(null)
  const [availablePrinters, setAvailablePrinters] = useState<PrinterDevice[]>([])
  const [error, setError] = useState<string | null>(null)

  const isCapacitor = false

  const scanForPrinters = useCallback(async (): Promise<PrinterDevice[]> => {
    setIsScanning(true)
    setError(null)
    const printers: PrinterDevice[] = []

    try {
      // Network printers (fetch from API)
      try {
        const { api } = await import('@/api/client')
        const response = await api.get('/v1/print/printers')
        const networkPrinters = response.data?.data || response.data || []
        
        for (const p of networkPrinters) {
          printers.push({
            id: `network-${p.id}`,
            name: p.name,
            type: 'network',
            address: p.ip_address || p.host
          })
        }
      } catch (e) {
        console.warn('Failed to fetch network printers:', e)
      }

      setAvailablePrinters(printers)
      return printers
    } catch (e) {
      const errMsg = e instanceof Error ? e.message : 'Failed to scan for printers'
      setError(errMsg)
      return []
    } finally {
      setIsScanning(false)
    }
  }, [])

  const connectPrinter = useCallback(async (printer: PrinterDevice): Promise<boolean> => {
    setError(null)
    
    try {
      if (printer.type === 'network') {
        setConnectedPrinter({ ...printer, connected: true })
        return true
      }
      
      setError('Unsupported printer type')
      return false
    } catch (e) {
      const errMsg = e instanceof Error ? e.message : 'Failed to connect'
      setError(errMsg)
      return false
    }
  }, [])

  const disconnectPrinter = useCallback(async (): Promise<void> => {
    setConnectedPrinter(null)
  }, [])

  const printReceipt = useCallback(async (content: string): Promise<boolean> => {
    if (!connectedPrinter) {
      setError('No printer connected')
      return false
    }

    setIsPrinting(true)
    setError(null)

    try {
      if (connectedPrinter.type === 'network') {
        const { api } = await import('@/api/client')
        await api.post('/v1/print/print', {
          printer_id: connectedPrinter.id.replace('network-', ''),
          content
        })
      } else {
        // Fallback: Web printing
        const printWindow = window.open('', '_blank')
        if (printWindow) {
          printWindow.document.write(`
            <html>
              <head>
                <title>Print Receipt</title>
                <style>
                  @media print {
                    body { font-family: monospace; font-size: 12px; }
                  }
                </style>
              </head>
              <body>${content}</body>
            </html>
          `)
          printWindow.document.close()
          printWindow.print()
        }
      }

      return true
    } catch (e) {
      const errMsg = e instanceof Error ? e.message : 'Print failed'
      setError(errMsg)
      return false
    } finally {
      setIsPrinting(false)
    }
  }, [connectedPrinter])

  const printTestPage = useCallback(async (): Promise<boolean> => {
    const testContent = `
================================
       DUKAPOS TEST PRINT
================================

Printer: ${connectedPrinter?.name || 'Unknown'}
Date: ${new Date().toLocaleString()}

This is a test print to verify
your printer is working correctly.

================================
`
    return printReceipt(testContent)
  }, [connectedPrinter, printReceipt])

  return {
    isPrinting,
    isScanning,
    connectedPrinter,
    availablePrinters,
    error,
    scanForPrinters,
    connectPrinter,
    disconnectPrinter,
    printReceipt,
    printTestPage,
    isCapacitor
  }
}

export function formatReceipt(data: {
  shopName: string
  shopPhone: string
  items: Array<{ name: string; quantity: number; price: number; total: number }>
  subtotal: number
  tax: number
  total: number
  paymentMethod: string
  receiptNumber: string
  cashier?: string
}): string {
  const { shopName, shopPhone, items, subtotal, tax, total, paymentMethod, receiptNumber, cashier } = data
  
  let receipt = `
${shopName}
${shopPhone}
================================
RECEIPT #: ${receiptNumber}
Date: ${new Date().toLocaleString()}
${cashier ? `Cashier: ${cashier}\n` : ''}
--------------------------------
`

  for (const item of items) {
    const name = item.name.slice(0, 20).padEnd(20)
    const qty = `${item.quantity}`.padStart(3)
    const price = `$${item.price.toFixed(2)}`
    const line = `${name}${qty} ${price}`
    receipt += line.slice(0, 32) + '\n'
  }

  receipt += `--------------------------------
Subtotal:${subtotal.toFixed(2).padStart(14)}
Tax:    ${tax.toFixed(2).padStart(14)}
================================
TOTAL:  ${total.toFixed(2).padStart(14)}
================================
Payment: ${paymentMethod.toUpperCase()}

Thank you for your business!
`

  return receipt
}
