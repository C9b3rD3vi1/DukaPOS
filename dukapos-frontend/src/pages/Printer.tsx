import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { usePrinter } from '@/hooks/usePrinter'
import { Card, StatCard } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Skeleton } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { PrinterConfig } from '@/api/types'

export default function Printer() {
  const shop = useAuthStore((state) => state.shop)
  const [printers, setPrinters] = useState<PrinterConfig[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [testing, setTesting] = useState(false)
  const [form, setForm] = useState({ name: '', type: 'thermal', connection_type: 'usb', is_default: false })
  const [editingPrinter, setEditingPrinter] = useState<PrinterConfig | null>(null)
  const [showScanModal, setShowScanModal] = useState(false)
  
  const {
    isPrinting,
    isScanning,
    connectedPrinter,
    availablePrinters,
    scanForPrinters,
    connectPrinter,
    disconnectPrinter,
    printTestPage,
    printReceipt,
    isCapacitor
  } = usePrinter()

  useEffect(() => {
    if (shop?.id) {
      fetchPrinters()
    } else {
      setIsLoading(false)
    }
  }, [shop?.id])

  const fetchPrinters = async () => {
    try {
      const response = await api.get('/v1/print/printers')
      const responseData = response.data
      const printersData = responseData?.data || responseData || []
      setPrinters(Array.isArray(printersData) ? printersData : [])
    } catch (err) {
      console.error(err)
      setError('Unable to load printers')
      setPrinters([])
    } finally {
      setIsLoading(false)
    }
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await api.post('/v1/print/printers', { ...form, shop_id: shop?.id })
      setShowConfigModal(false)
      setForm({ name: '', type: 'thermal', connection_type: 'usb', is_default: false })
      fetchPrinters()
    } catch (err) {
      setError('Failed to add printer')
    }
  }

  const handleEdit = (printer: PrinterConfig) => {
    setEditingPrinter(printer)
    setForm({
      name: printer.name,
      type: printer.type,
      connection_type: printer.connection_type,
      is_default: printer.is_default
    })
    setShowConfigModal(true)
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingPrinter) return
    setError('')
    try {
      await api.put(`/v1/print/printers/${editingPrinter.id}`, form)
      setShowConfigModal(false)
      setEditingPrinter(null)
      setForm({ name: '', type: 'thermal', connection_type: 'usb', is_default: false })
      fetchPrinters()
    } catch (err) {
      setError('Failed to update printer')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this printer?')) return
    try {
      await api.delete(`/v1/print/printers/${id}`)
      fetchPrinters()
    } catch (err) {
      setError('Failed to delete printer')
    }
  }

  const handleSetDefault = async (printer: PrinterConfig) => {
    try {
      await api.put(`/v1/print/printers/${printer.id}`, { ...printer, is_default: true })
      fetchPrinters()
    } catch (err) {
      setError('Failed to set default printer')
    }
  }

  const handleTest = async () => {
    setTesting(true)
    try {
      // Try Bluetooth first if connected
      if (connectedPrinter && connectedPrinter.type === 'bluetooth') {
        const success = await printTestPage()
        if (success) {
          alert('Test print sent via Bluetooth!')
        } else {
          alert('Failed to print via Bluetooth')
        }
      } else {
        // Fallback to API
        await api.post('/v1/print/test')
        alert('Test print sent!')
      }
    } catch (err) {
      alert('Failed to send test print. Make sure a printer is configured.')
    } finally {
      setTesting(false)
    }
  }

  const handlePrintReceipt = async () => {
    try {
      if (connectedPrinter && connectedPrinter.type === 'bluetooth') {
        const testReceipt = `
DUKAPOS TEST
------------
Shop: ${shop?.name || 'Test Shop'}
Date: ${new Date().toLocaleString()}
Thank you for testing!
        `
        const success = await printReceipt(testReceipt)
        if (success) {
          alert('Receipt sent to Bluetooth printer!')
        } else {
          alert('Failed to print receipt')
        }
      } else {
        await api.post('/v1/print/receipt', { sale_id: 1 })
        alert('Receipt sent to printer!')
      }
    } catch (err) {
      alert('Failed to print receipt')
    }
  }

  const handleScanForPrinters = async () => {
    setShowScanModal(true)
    await scanForPrinters()
  }

  const handleConnectPrinter = async (printer: typeof availablePrinters[0]) => {
    const success = await connectPrinter(printer)
    if (success) {
      alert(`Connected to ${printer.name}`)
      setShowScanModal(false)
    }
  }

  const handleDisconnect = async () => {
    await disconnectPrinter()
    alert('Disconnected from printer')
  }

  const closeModal = () => {
    setShowConfigModal(false)
    setEditingPrinter(null)
    setForm({ name: '', type: 'thermal', connection_type: 'usb', is_default: false })
  }

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <Skeleton className="h-8 w-48 mb-2" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="space-y-3">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
      </div>
    )
  }

  const totalPrinters = printers.length
  const defaultPrinter = printers.filter(p => p.is_default).length

  return (
    <div className="-mx-4 md:-mx-6">
      <div className="px-4 md:px-6 pb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-surface-900">Printer Settings</h1>
            <p className="text-surface-500 mt-1">Configure thermal printers for receipts</p>
          </div>
          <Button onClick={() => setShowConfigModal(true)} leftIcon={
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          }>
            Add Printer
          </Button>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-xl">
            {error}
          </div>
        )}

        {/* Stats */}
        {shop && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
            <StatCard
              title="Total Printers"
              value={totalPrinters}
              variant="default"
              icon={
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
              }
            />
            <StatCard
              title="Default Printer"
              value={defaultPrinter}
              variant="success"
              icon={
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              }
            />
          </div>
        )}

        {/* Quick Actions */}
        <div className="grid md:grid-cols-3 gap-4 mb-6">
          <button
            onClick={handleTest}
            disabled={testing || isPrinting}
            className="bg-white rounded-xl border border-surface-200 p-4 hover:bg-surface-50 transition flex items-center gap-4 disabled:opacity-50"
          >
            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
              </svg>
            </div>
            <div className="text-left">
              <p className="font-medium text-surface-900">Test Print</p>
              <p className="text-sm text-surface-500">Print a test page</p>
            </div>
          </button>

          <button
            onClick={handlePrintReceipt}
            disabled={isPrinting}
            className="bg-white rounded-xl border border-surface-200 p-4 hover:bg-surface-50 transition flex items-center gap-4 disabled:opacity-50"
          >
            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div className="text-left">
              <p className="font-medium text-surface-900">Print Receipt</p>
              <p className="text-sm text-surface-500">Print last sale receipt</p>
            </div>
          </button>

          {isCapacitor && (
            <button
              onClick={handleScanForPrinters}
              disabled={isScanning}
              className="bg-white rounded-xl border border-surface-200 p-4 hover:bg-surface-50 transition flex items-center gap-4 disabled:opacity-50"
            >
              <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
              </div>
              <div className="text-left">
                <p className="font-medium text-surface-900">Scan Bluetooth</p>
                <p className="text-sm text-surface-500">Find nearby printers</p>
              </div>
            </button>
          )}
        </div>

        {/* Connected Printer Status */}
        {connectedPrinter && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-xl">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                  <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div>
                  <p className="font-medium text-green-800">Connected to {connectedPrinter.name}</p>
                  <p className="text-sm text-green-600">{connectedPrinter.type} • {connectedPrinter.address}</p>
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={handleDisconnect}>
                Disconnect
              </Button>
            </div>
          </div>
        )}

        {/* Printers List */}
        <Card padding="none">
          <div className="p-4 border-b border-surface-100">
            <h2 className="font-semibold text-surface-900">Configured Printers</h2>
          </div>
          {!shop ? (
            <div className="p-8">
              <EmptyState
                variant="generic"
                title="No Shop Selected"
                description="Please select a shop to view printers"
              />
            </div>
          ) : printers.length === 0 ? (
            <div className="p-8">
              <EmptyState
                variant="generic"
                title="No Printers Configured"
                description="Add a printer to start printing receipts"
                action={{
                  label: 'Add Printer',
                  onClick: () => setShowConfigModal(true),
                }}
              />
            </div>
          ) : (
            <div className="divide-y divide-surface-100">
              {printers.map((printer) => (
                <div key={printer.id} className="p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`w-3 h-3 rounded-full ${printer.is_default ? 'bg-green-500' : 'bg-surface-300'}`}></div>
                    <div>
                      <p className="font-medium text-surface-900">{printer.name}</p>
                      <p className="text-sm text-surface-500 capitalize">{printer.type} - {printer.connection_type}</p>
                    </div>
                    {printer.is_default && (
                      <span className="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded-full">Default</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    {!printer.is_default && (
                      <button
                        onClick={() => handleSetDefault(printer)}
                        className="px-3 py-1.5 text-sm text-surface-600 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                      >
                        Set Default
                      </button>
                    )}
                    <button
                      onClick={() => handleEdit(printer)}
                      className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDelete(printer.id)}
                      className="p-2 text-surface-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>

        {/* Printer Info */}
        <Card className="mt-6">
          <h3 className="font-semibold text-surface-900 mb-4">Supported Printers</h3>
          <div className="grid md:grid-cols-3 gap-4">
            <div className="p-4 bg-surface-50 rounded-xl">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                  <span className="text-blue-600 font-bold text-xs">USB</span>
                </div>
                <p className="font-medium text-surface-900">USB Printers</p>
              </div>
              <p className="text-sm text-surface-500">Epson TM-T88, Star TSP100, thermal receipt printers</p>
            </div>
            <div className="p-4 bg-surface-50 rounded-xl">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
                  <span className="text-purple-600 font-bold text-xs">BT</span>
                </div>
                <p className="font-medium text-surface-900">Bluetooth</p>
              </div>
              <p className="text-sm text-surface-500">Star SM-S220i, Epson TM-P80, portable printers</p>
            </div>
            <div className="p-4 bg-surface-50 rounded-xl">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
                  <span className="text-green-600 font-bold text-xs">IP</span>
                </div>
                <p className="font-medium text-surface-900">Network</p>
              </div>
              <p className="text-sm text-surface-500">Epson TM-T88VI, Star TSP700II, cloud printers</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Config Modal */}
      {showConfigModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-md rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">
                  {editingPrinter ? 'Edit Printer' : 'Add Printer'}
                </h3>
                <button onClick={closeModal} className="p-2 hover:bg-surface-100 rounded-xl transition-colors">
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <form onSubmit={editingPrinter ? handleUpdate : handleAdd} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Printer Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="e.g., Receipt Printer"
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Printer Type</label>
                <select
                  value={form.type}
                  onChange={(e) => setForm({ ...form, type: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                >
                  <option value="thermal">Thermal Printer</option>
                  <option value="laser">Laser Printer</option>
                  <option value="inkjet">Inkjet Printer</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Connection</label>
                <select
                  value={form.connection_type}
                  onChange={(e) => setForm({ ...form, connection_type: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                >
                  <option value="usb">USB</option>
                  <option value="bluetooth">Bluetooth</option>
                  <option value="network">Network (WiFi/Ethernet)</option>
                </select>
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.is_default}
                  onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
                  className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                />
                <span className="text-sm text-surface-700">Set as default printer</span>
              </label>
              <div className="flex gap-3 pt-4">
                <Button type="button" variant="secondary" onClick={closeModal} className="flex-1">
                  Cancel
                </Button>
                <Button type="submit" variant="primary" className="flex-1">
                  {editingPrinter ? 'Update' : 'Save'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Bluetooth Scan Modal */}
      {showScanModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-md max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-surface-100 sticky top-0 bg-white">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Scan for Printers</h3>
                <button onClick={() => setShowScanModal(false)} className="p-2 hover:bg-surface-100 rounded-xl">
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="p-4 space-y-4">
              {isScanning ? (
                <div className="text-center py-8">
                  <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                  <p className="text-surface-600">Scanning for printers...</p>
                </div>
              ) : availablePrinters.length === 0 ? (
                <div className="text-center py-8">
                  <div className="w-16 h-16 bg-surface-100 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg className="w-8 h-8 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <p className="text-surface-600 mb-4">No printers found nearby</p>
                  <Button variant="outline" onClick={() => scanForPrinters()}>
                    Scan Again
                  </Button>
                </div>
              ) : (
                <div className="space-y-2">
                  <p className="text-sm text-surface-500 mb-4">Found {availablePrinters.length} printer(s):</p>
                  {availablePrinters.map((printer) => (
                    <button
                      key={printer.id}
                      onClick={() => handleConnectPrinter(printer)}
                      className="w-full p-4 bg-surface-50 hover:bg-surface-100 rounded-xl border border-surface-200 flex items-center gap-4 transition"
                    >
                      <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                        printer.type === 'bluetooth' ? 'bg-blue-100' : 'bg-green-100'
                      }`}>
                        <svg className={`w-5 h-5 ${printer.type === 'bluetooth' ? 'text-blue-600' : 'text-green-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                        </svg>
                      </div>
                      <div className="text-left flex-1">
                        <p className="font-medium text-surface-900">{printer.name}</p>
                        <p className="text-sm text-surface-500">{printer.type}</p>
                      </div>
                      <svg className="w-5 h-5 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
