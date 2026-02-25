import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard } from '@/components/common/Card'
import { SkeletonList } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { Supplier, Order } from '@/api/types'

export default function Suppliers() {
  const shop = useAuthStore((state) => state.shop)
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [orders, setOrders] = useState<Order[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'suppliers' | 'orders'>('suppliers')
  const [search, setSearch] = useState('')
  const [showSupplierModal, setShowSupplierModal] = useState(false)
  const [showOrderModal, setShowOrderModal] = useState(false)
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [selectedSupplier, setSelectedSupplier] = useState<Supplier | null>(null)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [editingSupplier, setEditingSupplier] = useState<Supplier | null>(null)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<number | 'bulk' | 'supplier' | 'order' | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [supplierForm, setSupplierForm] = useState({ name: '', phone: '', email: '', address: '' })
  const [orderForm, setOrderForm] = useState({ supplier_id: 0, notes: '', total_amount: 0, items: [] as any[] })

  useEffect(() => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    fetchSuppliers()
    fetchOrders()
  }, [shop?.id])

  const fetchSuppliers = async () => {
    if (!shop?.id) return
    try {
      const params = new URLSearchParams()
      if (search) params.append('search', search)
      
      const response = await api.get(`/v1/suppliers?${params}`)
      const responseData = response.data
      const suppliersData = responseData?.data || responseData || []
      setSuppliers(Array.isArray(suppliersData) ? suppliersData : [])
    } catch (err) { 
      console.error(err) 
      setSuppliers([])
    } finally { 
      setIsLoading(false) 
    }
  }

  const fetchOrders = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get('/v1/orders')
      const responseData = response.data
      const ordersData = responseData?.data || responseData || []
      setOrders(Array.isArray(ordersData) ? ordersData : [])
    } catch (err) { 
      console.error(err) 
      setOrders([])
    }
  }

  useEffect(() => {
    const debounce = setTimeout(() => {
      if (shop?.id) fetchSuppliers()
    }, 300)
    return () => clearTimeout(debounce)
  }, [search, shop?.id])

  const handleSaveSupplier = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      if (editingSupplier) {
        await api.put(`/v1/suppliers/${editingSupplier.id}`, supplierForm)
      } else {
        await api.post('/v1/suppliers', { ...supplierForm, shop_id: shop?.id })
      }
      setShowSupplierModal(false)
      setEditingSupplier(null)
      setSupplierForm({ name: '', phone: '', email: '', address: '' })
      fetchSuppliers()
    } catch (err) { 
      console.error(err) 
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCreateOrder = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      await api.post('/v1/orders', { 
        ...orderForm, 
        shop_id: shop?.id,
        status: 'pending'
      })
      setShowOrderModal(false)
      setOrderForm({ supplier_id: 0, notes: '', total_amount: 0, items: [] })
      fetchOrders()
    } catch (err) { 
      console.error(err) 
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleEditSupplier = (supplier: Supplier) => {
    setEditingSupplier(supplier)
    setSupplierForm({ 
      name: supplier.name, 
      phone: supplier.phone || '', 
      email: supplier.email || '', 
      address: supplier.address || '' 
    })
    setShowSupplierModal(true)
  }

  const handleViewSupplier = (supplier: Supplier) => {
    setSelectedSupplier(supplier)
    setShowDetailsModal(true)
  }

  const handleViewOrder = (order: Order) => {
    setSelectedOrder(order)
    setShowDetailsModal(true)
  }

  const confirmDelete = async () => {
    if (deleteTarget === 'bulk') {
      await handleBulkDelete()
    } else if (deleteTarget === 'supplier') {
      await handleDeleteSupplier(selectedSupplier?.id!)
    } else if (deleteTarget === 'order') {
      await handleDeleteOrder(selectedOrder?.id!)
    }
    setShowDeleteConfirm(false)
    setDeleteTarget(null)
  }

  const handleDeleteSupplier = async (id: number) => {
    try {
      await api.delete(`/v1/suppliers/${id}`)
      fetchSuppliers()
    } catch (err) { 
      console.error(err) 
    }
  }

  const handleDeleteOrder = async (id: number) => {
    try {
      await api.delete(`/v1/orders/${id}`)
      fetchOrders()
    } catch (err) { 
      console.error(err) 
    }
  }

  const handleBulkDelete = async () => {
    try {
      await api.post('/v1/suppliers/bulk-delete', { ids: selectedIds })
      setSelectedIds([])
      fetchSuppliers()
    } catch (err) { 
      console.error(err) 
    }
  }

  const toggleSelectAll = () => {
    if (activeTab === 'suppliers') {
      if (selectedIds.length === suppliers.length) {
        setSelectedIds([])
      } else {
        setSelectedIds(suppliers.map(s => s.id))
      }
    }
  }

  const toggleSelect = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(i => i !== id))
    } else {
      setSelectedIds([...selectedIds, id])
    }
  }

  const updateOrderStatus = async (orderId: number, status: string) => {
    try {
      await api.put(`/v1/orders/${orderId}`, { status })
      fetchOrders()
    } catch (err) {
      console.error(err)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'completed': return 'bg-green-100 text-green-700'
      case 'pending': return 'bg-yellow-100 text-yellow-700'
      case 'cancelled': return 'bg-red-100 text-red-700'
      default: return 'bg-gray-100 text-gray-700'
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const totalSuppliers = suppliers.length
  const totalOrders = orders.length
  const pendingOrders = orders.filter(o => o.status === 'pending').length
  const totalValue = orders.reduce((sum, o) => sum + (o.total_amount || 0), 0)

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Suppliers & Orders</h1>
          <p className="text-gray-500 mt-1">{suppliers.length} suppliers, {orders.length} orders</p>
        </div>
        <div className="flex gap-2">
          {activeTab === 'suppliers' && (
            <>
              <button 
                onClick={() => { setEditingSupplier(null); setSupplierForm({ name: '', phone: '', email: '', address: '' }); setShowSupplierModal(true); }}
                className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Add Supplier
              </button>
              {selectedIds.length > 0 && (
                <button
                  onClick={() => { setDeleteTarget('bulk'); setShowDeleteConfirm(true); }}
                  className="px-4 py-2.5 bg-red-50 text-red-600 rounded-xl hover:bg-red-100"
                >
                  Delete ({selectedIds.length})
                </button>
              )}
            </>
          )}
          {activeTab === 'orders' && (
            <button onClick={() => setShowOrderModal(true)} className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              New Order
            </button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('suppliers')}
          className={`px-4 py-2 rounded-xl text-sm font-medium transition ${
            activeTab === 'suppliers' ? 'bg-primary text-white' : 'bg-white text-gray-600 hover:bg-gray-50 border border-gray-200'
          }`}
        >
          Suppliers ({suppliers.length})
        </button>
        <button
          onClick={() => setActiveTab('orders')}
          className={`px-4 py-2 rounded-xl text-sm font-medium transition ${
            activeTab === 'orders' ? 'bg-primary text-white' : 'bg-white text-gray-600 hover:bg-gray-50 border border-gray-200'
          }`}
        >
          Orders ({orders.length})
        </button>
      </div>

      {/* Search */}
      <div className="bg-white rounded-xl border border-gray-200 p-4 mb-6">
        <div className="relative">
          <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder={`Search ${activeTab}...`}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
        </div>
      </div>

      {/* Stats */}
      {!isLoading && shop && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <StatCard
            title="Suppliers"
            value={totalSuppliers}
            variant="default"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.117v.547c0 .409-.252.818-.612 1.028a4.5 4.5 0 01-3.742 2.391l-2.431-2.431a1.125 1.125 0 00-1.533.083l-.884.884a1.125 1.125 0 11-1.533-.083l-.884-.884a1.125 1.125 0 00-.083-1.533l2.431-2.431a4.5 4.5 0 01-2.391-3.742A1.106 1.106 0 005.625 4.5V6h13.5z" />
              </svg>
            }
          />
          <StatCard
            title="Total Orders"
            value={totalOrders}
            variant="info"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            }
          />
          <StatCard
            title="Pending"
            value={pendingOrders}
            variant="warning"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
          <StatCard
            title="Total Value"
            value={formatCurrency(totalValue)}
            variant="success"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
        </div>
      )}

      {isLoading ? (
        <SkeletonList items={5} />
      ) : !shop ? (
        <Card className="text-center py-12">
          <EmptyState
            variant="generic"
            title="No Shop Selected"
            description="Please select a shop to view suppliers"
          />
        </Card>
      ) : activeTab === 'suppliers' ? (
        suppliers.length === 0 ? (
          <Card className="text-center py-12">
            <EmptyState
              variant="generic"
              title={search ? 'No suppliers found' : 'No Suppliers Yet'}
              description={search ? 'Try adjusting your search' : 'Add suppliers to manage your inventory'}
              action={!search ? {
                label: 'Add Supplier',
                onClick: () => { setEditingSupplier(null); setSupplierForm({ name: '', phone: '', email: '', address: '' }); setShowSupplierModal(true); },
              } : undefined}
            />
          </Card>
        ) : (
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left">
                    <input type="checkbox" checked={selectedIds.length === suppliers.length && suppliers.length > 0} onChange={toggleSelectAll} className="w-4 h-4 rounded" />
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Phone</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {suppliers.map((supplier) => (
                  <tr key={supplier.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <input type="checkbox" checked={selectedIds.includes(supplier.id)} onChange={() => toggleSelect(supplier.id)} className="w-4 h-4 rounded" />
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-900">{supplier.name}</td>
                    <td className="px-4 py-3 text-gray-600">{supplier.phone || '-'}</td>
                    <td className="px-4 py-3 text-gray-600">{supplier.email || '-'}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-center gap-1">
                        <button onClick={() => handleViewSupplier(supplier)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg" title="View">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                        </button>
                        <button onClick={() => handleEditSupplier(supplier)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg" title="Edit">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                        </button>
                        <button onClick={() => { setSelectedSupplier(supplier); setDeleteTarget('supplier'); setShowDeleteConfirm(true); }} className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg" title="Delete">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      ) : (
        orders.length === 0 ? (
          <Card className="text-center py-12">
            <EmptyState
              variant="orders"
              title={search ? 'No orders found' : 'No Orders Yet'}
              description={search ? 'Try adjusting your search' : 'Create orders from suppliers'}
              action={!search ? {
                label: 'New Order',
                onClick: () => setShowOrderModal(true),
              } : undefined}
            />
          </Card>
        ) : (
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Order #</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Supplier</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Amount</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {orders.map((order) => (
                  <tr key={order.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-medium text-gray-900">#{order.id}</td>
                    <td className="px-4 py-3 text-gray-600">{order.supplier?.name || `Supplier #${order.supplier_id}`}</td>
                    <td className="px-4 py-3 text-right font-medium text-gray-900">{formatCurrency(order.total_amount)}</td>
                    <td className="px-4 py-3 text-center">
                      <select
                        value={order.status}
                        onChange={(e) => updateOrderStatus(order.id, e.target.value)}
                        className={`px-2 py-1 rounded-full text-xs font-medium border-0 cursor-pointer ${getStatusColor(order.status)}`}
                      >
                        <option value="pending">Pending</option>
                        <option value="completed">Completed</option>
                        <option value="cancelled">Cancelled</option>
                      </select>
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-sm">{new Date(order.created_at).toLocaleDateString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-center gap-1">
                        <button onClick={() => handleViewOrder(order)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg" title="View">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                        </button>
                        <button onClick={() => { setSelectedOrder(order); setDeleteTarget('order'); setShowDeleteConfirm(true); }} className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg" title="Delete">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {/* Supplier Modal */}
      {showSupplierModal && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-lg font-bold text-gray-900">{editingSupplier ? 'Edit' : 'Add'} Supplier</h3>
            </div>
            <form onSubmit={handleSaveSupplier} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                <input type="text" value={supplierForm.name} onChange={(e) => setSupplierForm({...supplierForm, name: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" required />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                <input type="tel" value={supplierForm.phone} onChange={(e) => setSupplierForm({...supplierForm, phone: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                <input type="email" value={supplierForm.email} onChange={(e) => setSupplierForm({...supplierForm, email: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Address</label>
                <textarea value={supplierForm.address} onChange={(e) => setSupplierForm({...supplierForm, address: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" rows={2} />
              </div>
              <div className="flex gap-3 pt-4">
                <button type="button" onClick={() => { setShowSupplierModal(false); setEditingSupplier(null); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Cancel</button>
                <button type="submit" disabled={isSubmitting} className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50">
                  {isSubmitting ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Order Modal */}
      {showOrderModal && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-lg font-bold text-gray-900">Create Order</h3>
            </div>
            <form onSubmit={handleCreateOrder} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Supplier *</label>
                <select value={orderForm.supplier_id} onChange={(e) => setOrderForm({...orderForm, supplier_id: parseInt(e.target.value)})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" required>
                  <option value="">Select Supplier</option>
                  {suppliers.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Total Amount</label>
                <input type="number" value={orderForm.total_amount} onChange={(e) => setOrderForm({...orderForm, total_amount: parseFloat(e.target.value) || 0})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
                <textarea value={orderForm.notes} onChange={(e) => setOrderForm({...orderForm, notes: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none" rows={3} />
              </div>
              <div className="flex gap-3 pt-4">
                <button type="button" onClick={() => setShowOrderModal(false)} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Cancel</button>
                <button type="submit" disabled={isSubmitting || !orderForm.supplier_id} className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50">
                  {isSubmitting ? 'Creating...' : 'Create Order'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Details Modal */}
      {showDetailsModal && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100 flex justify-between items-center">
              <h3 className="text-lg font-bold text-gray-900">
                {selectedSupplier ? 'Supplier Details' : selectedOrder ? 'Order Details' : 'Details'}
              </h3>
              <button onClick={() => setShowDetailsModal(false)} className="text-gray-400 hover:text-gray-600">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div className="p-6 space-y-4">
              {selectedSupplier && (
                <>
                  <div className="text-center pb-4 border-b border-gray-100">
                    <div className="w-16 h-16 bg-primary-50 rounded-full flex items-center justify-center mx-auto mb-3">
                      <span className="text-2xl text-primary font-bold">{selectedSupplier.name.charAt(0)}</span>
                    </div>
                    <h4 className="text-xl font-bold text-gray-900">{selectedSupplier.name}</h4>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div><p className="text-sm text-gray-500">Phone</p><p className="font-medium">{selectedSupplier.phone || '-'}</p></div>
                    <div><p className="text-sm text-gray-500">Email</p><p className="font-medium">{selectedSupplier.email || '-'}</p></div>
                    <div className="col-span-2"><p className="text-sm text-gray-500">Address</p><p className="font-medium">{selectedSupplier.address || '-'}</p></div>
                  </div>
                </>
              )}
              {selectedOrder && (
                <>
                  <div className="text-center pb-4 border-b border-gray-100">
                    <h4 className="text-xl font-bold text-gray-900">Order #{selectedOrder.id}</h4>
                    <p className="text-gray-500">{new Date(selectedOrder.created_at).toLocaleDateString()}</p>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div><p className="text-sm text-gray-500">Supplier</p><p className="font-medium">{selectedOrder.supplier?.name || `Supplier #${selectedOrder.supplier_id}`}</p></div>
                    <div><p className="text-sm text-gray-500">Amount</p><p className="font-medium">{formatCurrency(selectedOrder.total_amount)}</p></div>
                    <div><p className="text-sm text-gray-500">Status</p><span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(selectedOrder.status)}`}>{selectedOrder.status}</span></div>
                    {selectedOrder.notes && <div className="col-span-2"><p className="text-sm text-gray-500">Notes</p><p className="font-medium">{selectedOrder.notes}</p></div>}
                  </div>
                </>
              )}
            </div>
            <div className="p-6 pt-0 flex gap-3">
              {selectedSupplier && (
                <>
                  <button onClick={() => { setShowDetailsModal(false); handleEditSupplier(selectedSupplier); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Edit</button>
                  <button onClick={() => { setShowDetailsModal(false); setDeleteTarget('supplier'); setShowDeleteConfirm(true); }} className="flex-1 px-4 py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100">Delete</button>
                </>
              )}
              {selectedOrder && (
                <button onClick={() => { setShowDetailsModal(false); setDeleteTarget('order'); setShowDeleteConfirm(true); }} className="flex-1 px-4 py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100">Delete Order</button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-sm rounded-2xl shadow-2xl p-6">
            <div className="text-center">
              <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">Confirm Delete</h3>
              <p className="text-gray-500 mb-6">
                {deleteTarget === 'bulk' ? `Delete ${selectedIds.length} suppliers? This cannot be undone.` : 'This action cannot be undone.'}
              </p>
              <div className="flex gap-3">
                <button onClick={() => { setShowDeleteConfirm(false); setDeleteTarget(null); }} className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50">Cancel</button>
                <button onClick={confirmDelete} className="flex-1 px-4 py-3 bg-red-600 text-white rounded-xl hover:bg-red-700">Delete</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
