import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { SkeletonList } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import { Badge } from '@/components/common/Badge'
import { StatCard, StatGrid } from '@/components/common/Card'
import type { Sale } from '@/api/types'

export default function Sales() {
  const shop = useAuthStore((state) => state.shop)
  const [sales, setSales] = useState<Sale[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'cash' | 'mpesa' | 'card'>('all')
  const [search, setSearch] = useState('')
  const [showViewModal, setShowViewModal] = useState(false)
  const [viewingSale, setViewingSale] = useState<Sale | null>(null)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingSale, setEditingSale] = useState<Sale | null>(null)
  const [editForm, setEditForm] = useState({ payment_method: 'cash' })
  const [selectedSales, setSelectedSales] = useState<number[]>([])
  const [selectAll, setSelectAll] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    const debounce = setTimeout(() => {
      fetchSales()
    }, 300)
    return () => clearTimeout(debounce)
  }, [shop?.id, filter, search])

  const fetchSales = async () => {
    if (!shop?.id) return
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      if (filter !== 'all') params.append('payment_method', filter)
      if (search) params.append('search', search)
      
      const response = await api.get(`/v1/sales?${params}`)
      const responseData = response.data
      const salesData = responseData?.data || responseData || []
      setSales(Array.isArray(salesData) ? salesData : [])
    } catch (err) {
      console.error(err)
      setSales([])
    } finally {
      setIsLoading(false)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const totalSales = sales.reduce((sum, s) => sum + s.total_amount, 0)
  const totalProfit = sales.reduce((sum, s) => sum + (s.profit || 0), 0)

  const getPaymentBadge = (method: string) => {
    switch (method) {
      case 'mpesa':
        return <Badge variant="warning" size="sm">M-Pesa</Badge>
      case 'card':
        return <Badge variant="info" size="sm">Card</Badge>
      case 'cash':
        return <Badge variant="success" size="sm">Cash</Badge>
      default:
        return <Badge variant="default" size="sm">{method}</Badge>
    }
  }

  const handleView = (sale: Sale) => {
    setViewingSale(sale)
    setShowViewModal(true)
  }

  const handleEdit = (sale: Sale) => {
    setEditingSale(sale)
    setEditForm({ payment_method: sale.payment_method || 'cash' })
    setShowEditModal(true)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this sale?')) return
    
    try {
      await api.delete(`/v1/sales/${id}`)
      fetchSales()
    } catch (err) {
      setError('Failed to delete sale')
    }
  }

  const handleUpdateStatus = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingSale) return

    try {
      await api.put(`/v1/sales/${editingSale.id}`, editForm)
      setShowEditModal(false)
      setEditingSale(null)
      fetchSales()
    } catch (err) {
      setError('Failed to update sale')
    }
  }

  const toggleSelectSale = (id: number) => {
    setSelectedSales(prev => 
      prev.includes(id) ? prev.filter(s => s !== id) : [...prev, id]
    )
  }

  const toggleSelectAll = () => {
    if (selectAll) {
      setSelectedSales([])
    } else {
      setSelectedSales(sales.map(s => s.id))
    }
    setSelectAll(!selectAll)
  }

  const handleBulkDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${selectedSales.length} sales?`)) return
    
    try {
      await Promise.all(selectedSales.map(id => api.delete(`/v1/sales/${id}`)))
      setSelectedSales([])
      setSelectAll(false)
      fetchSales()
    } catch (err) {
      setError('Failed to delete some sales')
    }
  }

  const exportSelected = () => {
    const selectedData = sales.filter(s => selectedSales.includes(s.id))
    const csv = [
      ['ID', 'Product', 'Quantity', 'Unit Price', 'Total', 'Payment Method', 'Date'].join(','),
      ...selectedData.map(s => [
        s.id,
        `"${s.product?.name || `Product #${s.product_id}`}"`,
        s.quantity,
        s.unit_price,
        s.total_amount,
        s.payment_method,
        new Date(s.created_at).toISOString()
      ].join(','))
    ].join('\n')
    
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'sales_export.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  const getPaymentIcon = (method: string) => {
    switch (method) {
      case 'mpesa':
        return <span className="inline-flex items-center gap-1 text-yellow-600"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" /></svg>M-Pesa</span>
      case 'card':
        return <span className="inline-flex items-center gap-1 text-blue-600"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>Card</span>
      default:
        return <span className="inline-flex items-center gap-1 text-green-600"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" /></svg>Cash</span>
    }
  }

  return (
    <div className="-mx-4 md:-mx-6">
      {/* Header */}
      <div className="px-4 md:px-6 pb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-surface-900">Sales</h1>
            <p className="text-surface-500 mt-1">{sales.length} transactions</p>
          </div>
          <Link 
            to="/sales/new" 
            className="inline-flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New Sale
          </Link>
        </div>

        {/* Search Bar */}
        <div className="relative mb-4">
          <svg className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search sales by product name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-12 pr-4 py-3.5 bg-white border border-surface-200 rounded-2xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none shadow-sm"
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-surface-400 hover:text-surface-600"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>

        {/* Bulk Actions Bar */}
        {sales.length > 0 && (
          <div className="flex items-center justify-between mb-4 p-3 bg-primary/5 rounded-xl border border-primary/10">
            <div className="flex items-center gap-3">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectAll}
                  onChange={toggleSelectAll}
                  className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                />
                <span className="text-sm font-medium text-surface-700">
                  {selectedSales.length > 0 ? `${selectedSales.length} selected` : 'Select all'}
                </span>
              </label>
            </div>
            {selectedSales.length > 0 && (
              <div className="flex items-center gap-2">
                <button
                  onClick={exportSelected}
                  className="px-3 py-1.5 text-sm font-medium text-surface-700 bg-white border border-surface-200 rounded-lg hover:bg-surface-50 transition"
                >
                  Export
                </button>
                <button
                  onClick={handleBulkDelete}
                  className="px-3 py-1.5 text-sm font-medium text-red-600 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100 transition"
                >
                  Delete
                </button>
              </div>
            )}
          </div>
        )}

        {/* Error Alert */}
        {error && (
          <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-xl">
            {error}
          </div>
        )}
      </div>

      {/* Stats */}
      <div className="px-4 md:px-6 mb-6">
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-white rounded-xl p-4 border border-surface-100 shadow-sm">
            <p className="text-sm text-surface-500">Total Sales</p>
            <p className="text-xl font-bold text-surface-900">{formatCurrency(totalSales)}</p>
          </div>
          <div className="bg-white rounded-xl p-4 border border-surface-100 shadow-sm">
            <p className="text-sm text-surface-500">Total Profit</p>
            <p className="text-xl font-bold text-green-600">{formatCurrency(totalProfit)}</p>
          </div>
        </div>
      </div>

      {/* Filter */}
      <div className="px-4 md:px-6 mb-4">
        <div className="flex gap-2 overflow-x-auto pb-2">
          {['all', 'cash', 'mpesa', 'card'].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f as typeof filter)}
              className={`px-4 py-2 rounded-xl text-sm font-medium transition whitespace-nowrap ${
                filter === f 
                  ? 'bg-primary text-white shadow-md shadow-primary/25' 
                  : 'bg-white text-surface-600 hover:bg-surface-50 border border-surface-200'
              }`}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="px-4 md:px-6">
        {isLoading ? (
          <SkeletonList items={5} />
        ) : !shop ? (
          <Card className="text-center py-12">
            <EmptyState
              variant="generic"
              title="No Shop Selected"
              description="Please select a shop to view sales"
            />
          </Card>
        ) : sales.length === 0 ? (
          <Card className="text-center py-12">
            <EmptyState
              variant={search || filter !== 'all' ? 'search' : 'sales'}
              title={search || filter !== 'all' ? 'No sales found' : 'No Sales Yet'}
              description={search || filter !== 'all' 
                ? 'Try adjusting your search or filters' 
                : 'Record your first sale to get started'}
              action={!search && filter === 'all' ? {
                label: 'Record Sale',
                to: '/sales/new',
              } : undefined}
            />
          </Card>
        ) : (
          <>
            {/* Stats */}
            <StatGrid columns={3}>
              <StatCard
                title="Total Sales"
                value={formatCurrency(totalSales)}
                variant="success"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                }
              />
              <StatCard
                title="Transactions"
                value={sales.length}
                variant="info"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                }
              />
              <StatCard
                title="Profit"
                value={formatCurrency(totalProfit)}
                variant="success"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                  </svg>
                }
              />
            </StatGrid>
            <Card padding="none">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-surface-50 border-b border-surface-200">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-surface-500 uppercase w-10">
                        <input
                          type="checkbox"
                          checked={selectAll}
                          onChange={toggleSelectAll}
                          className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                        />
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-surface-500 uppercase">Product</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Qty</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Price</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Total</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-surface-500 uppercase">Payment</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-surface-500 uppercase">Date</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-surface-100">
                    {sales.map((sale) => (
                      <tr key={sale.id} className="hover:bg-surface-50 transition-colors">
                        <td className="px-4 py-3">
                          <input
                            type="checkbox"
                            checked={selectedSales.includes(sale.id)}
                            onChange={() => toggleSelectSale(sale.id)}
                            className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                          />
                        </td>
                        <td className="px-4 py-3">
                          <p className="font-medium text-surface-900">
                            {sale.product?.name || `Product #${sale.product_id}`}
                          </p>
                        </td>
                        <td className="px-4 py-3 text-right text-surface-600">x{sale.quantity}</td>
                        <td className="px-4 py-3 text-right text-surface-600">
                          {formatCurrency(sale.unit_price)}
                        </td>
                        <td className="px-4 py-3 text-right font-semibold text-surface-900">
                          {formatCurrency(sale.total_amount)}
                        </td>
                        <td className="px-4 py-3">
                          {getPaymentBadge(sale.payment_method)}
                        </td>
                        <td className="px-4 py-3 text-surface-500 text-sm">
                          {new Date(sale.created_at).toLocaleDateString('en-KE')}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              onClick={() => handleView(sale)}
                              className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                              </svg>
                            </button>
                            <button
                              onClick={() => handleEdit(sale)}
                              className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                              </svg>
                            </button>
                            <button
                              onClick={() => handleDelete(sale.id)}
                              className="p-2 text-surface-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>
          </>
        )}
      </div>

      {/* View Details Modal */}
      {showViewModal && viewingSale && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-lg rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Sale Details</h3>
                <button
                  onClick={() => { setShowViewModal(false); setViewingSale(null); }}
                  className="p-2 hover:bg-surface-100 rounded-xl transition-colors"
                >
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="p-6 space-y-6">
              {/* Sale Info */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Product</p>
                  <p className="font-semibold text-surface-900">{viewingSale.product?.name || `Product #${viewingSale.product_id}`}</p>
                </div>
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Quantity</p>
                  <p className="font-semibold text-surface-900">{viewingSale.quantity}</p>
                </div>
              </div>

              {/* Pricing */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Unit Price</p>
                  <p className="text-lg font-bold text-surface-900">{formatCurrency(viewingSale.unit_price)}</p>
                </div>
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Total Amount</p>
                  <p className="text-lg font-bold text-primary">{formatCurrency(viewingSale.total_amount)}</p>
                </div>
              </div>

              {/* Payment & Profit */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Payment Method</p>
                  <p className="font-semibold text-surface-900">{getPaymentIcon(viewingSale.payment_method)}</p>
                </div>
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Profit</p>
                  <p className="text-lg font-bold text-green-600">{formatCurrency(viewingSale.profit || 0)}</p>
                </div>
              </div>

              {/* Date */}
              <div className="p-4 bg-surface-50 rounded-xl">
                <p className="text-sm text-surface-500">Date & Time</p>
                <p className="font-semibold text-surface-900">{new Date(viewingSale.created_at).toLocaleString()}</p>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-4">
                <Button
                  variant="secondary"
                  onClick={() => { setShowViewModal(false); setViewingSale(null); }}
                  className="flex-1"
                >
                  Close
                </Button>
                <Button
                  variant="primary"
                  onClick={() => {
                    setShowViewModal(false)
                    handleEdit(viewingSale)
                  }}
                  className="flex-1"
                >
                  Edit
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Status Modal */}
      {showEditModal && editingSale && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-md rounded-t-3xl md:rounded-3xl shadow-2xl">
            <div className="p-6 border-b border-surface-100">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Edit Sale</h3>
                <button
                  onClick={() => { setShowEditModal(false); setEditingSale(null); }}
                  className="p-2 hover:bg-surface-100 rounded-xl transition-colors"
                >
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <form onSubmit={handleUpdateStatus} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Payment Method</label>
                <select
                  value={editForm.payment_method}
                  onChange={(e) => setEditForm({ ...editForm, payment_method: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                >
                  <option value="cash">Cash</option>
                  <option value="mpesa">M-Pesa</option>
                  <option value="card">Card</option>
                </select>
              </div>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => { setShowEditModal(false); setEditingSale(null); }}
                  className="flex-1"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  className="flex-1"
                >
                  Update
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
