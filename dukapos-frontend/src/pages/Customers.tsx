import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { SkeletonList } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import { LoyaltyPanel } from '@/components/customers/LoyaltyPanel'
import type { Customer } from '@/api/types'

export default function Customers() {
  const shop = useAuthStore((state) => state.shop)
  const [customers, setCustomers] = useState<Customer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [search, setSearch] = useState('')
  const [formData, setFormData] = useState({ name: '', phone: '', email: '' })
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<number | 'bulk' | null>(null)
  const [activeTab, setActiveTab] = useState<'customers' | 'loyalty'>('customers')

  useEffect(() => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    fetchCustomers()
  }, [shop?.id])

  const fetchCustomers = async () => {
    if (!shop?.id) return
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      if (search) params.append('search', search)
      
      const response = await api.get(`/v1/customers?${params}`)
      const responseData = response.data
      const customersData = responseData?.data || responseData || []
      setCustomers(Array.isArray(customersData) ? customersData : [])
    } catch (err) {
      console.error(err)
      setCustomers([])
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    const debounce = setTimeout(() => {
      if (shop?.id) fetchCustomers()
    }, 300)
    return () => clearTimeout(debounce)
  }, [search, shop?.id])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)
    
    try {
      if (editingCustomer) {
        await api.put(`/v1/customers/${editingCustomer.id}`, formData)
      } else {
        await api.post('/v1/customers', { ...formData, shop_id: shop?.id })
      }
      setShowModal(false)
      setEditingCustomer(null)
      setFormData({ name: '', phone: '', email: '' })
      fetchCustomers()
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      setError(error.response?.data?.error || 'Failed to save customer')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleEdit = (customer: Customer) => {
    setEditingCustomer(customer)
    setFormData({ name: customer.name, phone: customer.phone, email: customer.email || '' })
    setShowModal(true)
  }

  const handleView = (customer: Customer) => {
    setSelectedCustomer(customer)
    setShowDetailsModal(true)
  }

  const confirmDelete = async () => {
    if (deleteTarget === 'bulk') {
      await handleBulkDelete()
    } else if (deleteTarget) {
      await handleDelete(deleteTarget)
    }
    setShowDeleteConfirm(false)
    setDeleteTarget(null)
  }

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/v1/customers/${id}`)
      fetchCustomers()
    } catch (err) {
      console.error(err)
    }
  }

  const handleBulkDelete = async () => {
    try {
      await api.post('/v1/customers/bulk-delete', { ids: selectedIds })
      setSelectedIds([])
      fetchCustomers()
    } catch (err) {
      console.error(err)
    }
  }

  const handleBulkExport = async () => {
    try {
      const response = await api.get(`/v1/customers/export?shop_id=${shop?.id}`, {
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `customers_${new Date().toISOString().split('T')[0]}.csv`)
      document.body.appendChild(link)
      link.click()
      link.remove()
    } catch (err) {
      console.error(err)
    }
  }

  const toggleSelectAll = () => {
    if (selectedIds.length === customers.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(customers.map(c => c.id))
    }
  }

  const toggleSelect = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(i => i !== id))
    } else {
      setSelectedIds([...selectedIds, id])
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Customers</h1>
          <p className="text-gray-500 mt-1">{customers.length} customers</p>
        </div>
        <div className="flex gap-2">
          <button 
            onClick={() => { setEditingCustomer(null); setFormData({ name: '', phone: '', email: '' }); setShowModal(true); }}
            className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Customer
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('customers')}
          className={`px-4 py-2 rounded-xl font-medium text-sm transition ${
            activeTab === 'customers' ? 'bg-primary text-white' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'
          }`}
        >
          All Customers
        </button>
        <button
          onClick={() => setActiveTab('loyalty')}
          className={`px-4 py-2 rounded-xl font-medium text-sm transition ${
            activeTab === 'loyalty' ? 'bg-primary text-white' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'
          }`}
        >
          Loyalty Program
        </button>
      </div>

      {/* Loyalty Panel */}
      {activeTab === 'loyalty' ? (
        <LoyaltyPanel />
      ) : (
      <>
      {/* Search & Filters */}
      <div className="bg-white rounded-xl border border-gray-200 p-4 mb-6">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              placeholder="Search customers..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
            />
          </div>
          {selectedIds.length > 0 && (
            <div className="flex gap-2">
              <span className="px-3 py-2 bg-primary-50 text-primary rounded-lg text-sm font-medium">
                {selectedIds.length} selected
              </span>
              <button
                onClick={() => { setDeleteTarget('bulk'); setShowDeleteConfirm(true); }}
                className="px-4 py-2 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 transition-all text-sm font-medium"
              >
                Delete Selected
              </button>
              <button
                onClick={handleBulkExport}
                className="px-4 py-2 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 transition-all text-sm font-medium"
              >
                Export
              </button>
            </div>
          )}
        </div>
      </div>

      {isLoading ? (
        <SkeletonList items={5} />
      ) : !shop ? (
        <Card className="text-center py-12">
          <EmptyState
            variant="generic"
            title="No Shop Selected"
            description="Please select a shop to view customers"
          />
        </Card>
      ) : customers.length === 0 ? (
        <Card className="text-center py-12">
          <EmptyState
            variant="customers"
            title={search ? 'No customers found' : 'No Customers Yet'}
            description={search ? 'Try adjusting your search' : 'Add customers to track loyalty points'}
            action={!search ? {
              label: 'Add Customer',
              onClick: () => setShowModal(true),
            } : undefined}
          />
        </Card>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={selectedIds.length === customers.length && customers.length > 0}
                      onChange={toggleSelectAll}
                      className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                    />
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Customer</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Phone</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Points</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Total Spent</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {customers.map((customer) => (
                  <tr key={customer.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedIds.includes(customer.id)}
                        onChange={() => toggleSelect(customer.id)}
                        className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 bg-primary-50 rounded-full flex items-center justify-center">
                          <span className="text-primary font-semibold">{customer.name.charAt(0)}</span>
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{customer.name}</p>
                          {customer.email && <p className="text-sm text-gray-500">{customer.email}</p>}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-600">{customer.phone}</td>
                    <td className="px-4 py-3 text-right">
                      <span className="font-semibold text-primary">{customer.loyalty_points}</span>
                    </td>
                    <td className="px-4 py-3 text-right font-medium text-gray-900">
                      {formatCurrency(customer.total_purchases || 0)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-center gap-1">
                        <button onClick={() => handleView(customer)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg transition" title="View">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                          </svg>
                        </button>
                        <button onClick={() => handleEdit(customer)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg transition" title="Edit">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                          </svg>
                        </button>
                        <button onClick={() => { setDeleteTarget(customer.id); setShowDeleteConfirm(true); }} className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition" title="Delete">
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
        </div>
      )}

      {/* Add/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-lg font-bold text-gray-900">{editingCustomer ? 'Edit' : 'Add'} Customer</h3>
            </div>
            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              {error && <div className="p-3 bg-red-50 text-red-600 rounded-xl text-sm">{error}</div>}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Phone *</label>
                <input
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button type="button" onClick={() => { setShowModal(false); setEditingCustomer(null); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Cancel</button>
                <button type="submit" disabled={isSubmitting} className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50">
                  {isSubmitting ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View Details Modal */}
      {showDetailsModal && selectedCustomer && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100 flex items-center justify-between">
              <h3 className="text-lg font-bold text-gray-900">Customer Details</h3>
              <button onClick={() => setShowDetailsModal(false)} className="text-gray-400 hover:text-gray-600">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div className="flex items-center gap-4">
                <div className="w-16 h-16 bg-primary-50 rounded-full flex items-center justify-center">
                  <span className="text-2xl text-primary font-bold">{selectedCustomer.name.charAt(0)}</span>
                </div>
                <div>
                  <h4 className="text-xl font-bold text-gray-900">{selectedCustomer.name}</h4>
                  <p className="text-gray-500">{selectedCustomer.phone}</p>
                  {selectedCustomer.email && <p className="text-gray-500">{selectedCustomer.email}</p>}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-100">
                <div className="text-center p-4 bg-primary-50 rounded-xl">
                  <p className="text-2xl font-bold text-primary">{selectedCustomer.loyalty_points}</p>
                  <p className="text-sm text-gray-500">Loyalty Points</p>
                </div>
                <div className="text-center p-4 bg-gray-50 rounded-xl">
                  <p className="text-2xl font-bold text-gray-900">{formatCurrency(selectedCustomer.total_purchases || 0)}</p>
                  <p className="text-sm text-gray-500">Total Spent</p>
                </div>
              </div>
              <div className="pt-4 border-t border-gray-100 text-sm text-gray-500">
                <p>Member since: {new Date(selectedCustomer.created_at).toLocaleDateString()}</p>
              </div>
            </div>
            <div className="p-6 pt-0 flex gap-3">
              <button onClick={() => { setShowDetailsModal(false); handleEdit(selectedCustomer); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Edit</button>
              <button onClick={() => { setShowDetailsModal(false); setDeleteTarget(selectedCustomer.id); setShowDeleteConfirm(true); }} className="flex-1 px-4 py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100">Delete</button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-sm rounded-2xl shadow-2xl p-6">
            <div className="text-center">
              <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">Confirm Delete</h3>
              <p className="text-gray-500 mb-6">
                {deleteTarget === 'bulk' 
                  ? `Are you sure you want to delete ${selectedIds.length} customers? This action cannot be undone.`
                  : 'Are you sure you want to delete this customer? This action cannot be undone.'
                }
              </p>
              <div className="flex gap-3">
                <button onClick={() => { setShowDeleteConfirm(false); setDeleteTarget(null); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Cancel</button>
                <button onClick={confirmDelete} className="flex-1 px-4 py-3 bg-red-600 text-white rounded-xl hover:bg-red-700">Delete</button>
              </div>
            </div>
          </div>
        </div>
      )}
      </>
      )}
    </div>
  )
}
