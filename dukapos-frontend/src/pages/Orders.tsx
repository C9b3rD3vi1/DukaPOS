import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard } from '@/components/common/Card'
import { SkeletonList } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { Order, Supplier, Product } from '@/api/types'

export default function Orders() {
  const shop = useAuthStore((state) => state.shop)
  const [orders, setOrders] = useState<Order[]>([])
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [filter, setFilter] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<number | 'bulk' | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [productSearch, setProductSearch] = useState('')
  const [selectedProducts, setSelectedProducts] = useState<{product_id: number; quantity: number; price: number; name: string}[]>([])
  const [formData, setFormData] = useState({ supplier_id: 0, notes: '', total_amount: 0, status: 'pending' })
  const [showProductList, setShowProductList] = useState(false)
  const [showNewProductModal, setShowNewProductModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editForm, setEditForm] = useState({ status: 'pending', notes: '', total_amount: 0 })
  const [newProductForm, setNewProductForm] = useState({
    name: '',
    category: '',
    cost_price: 0,
    selling_price: 0,
    current_stock: 0,
    unit: 'pcs'
  })

  useEffect(() => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    fetchOrders()
    fetchSuppliers()
    fetchProducts()
  }, [shop?.id, filter])

  const fetchOrders = async () => {
    if (!shop?.id) return
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      if (filter !== 'all') params.append('status', filter)
      
      const response = await api.get(`/v1/orders?${params}`)
      const responseData = response.data
      const ordersData = responseData?.data || responseData || []
      setOrders(Array.isArray(ordersData) ? ordersData : [])
    } catch (err) {
      console.error(err)
      setOrders([])
    } finally {
      setIsLoading(false)
    }
  }

  const fetchSuppliers = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get('/v1/suppliers')
      const responseData = response.data
      const suppliersData = responseData?.data || responseData || []
      setSuppliers(Array.isArray(suppliersData) ? suppliersData : [])
    } catch (err) {
      console.error(err)
      setSuppliers([])
    }
  }

  const fetchProducts = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get('/v1/products')
      const responseData = response.data
      const productsData = responseData?.data || responseData || []
      setProducts(Array.isArray(productsData) ? productsData : [])
    } catch (err) {
      console.error(err)
      setProducts([])
    }
  }

  const handleCreateProduct = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      const response = await api.post('/v1/products', { ...newProductForm, shop_id: shop?.id })
      const newProduct = response.data?.data || response.data
      if (newProduct) {
        setSelectedProducts([...selectedProducts, {
          product_id: newProduct.id,
          quantity: 1,
          price: newProduct.cost_price || newProduct.selling_price || 0,
          name: newProduct.name
        }])
        setProducts([...products, newProduct])
      }
      setShowNewProductModal(false)
      setNewProductForm({ name: '', category: '', cost_price: 0, selling_price: 0, current_stock: 0, unit: 'pcs' })
    } catch (err) {
      console.error(err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const addProduct = (product: Product) => {
    if (selectedProducts.find(p => p.product_id === product.id)) return
    setSelectedProducts([...selectedProducts, {
      product_id: product.id,
      quantity: 1,
      price: product.cost_price || product.selling_price || 0,
      name: product.name
    }])
    setProductSearch('')
  }

  const removeProduct = (productId: number) => {
    setSelectedProducts(selectedProducts.filter(p => p.product_id !== productId))
  }

  const updateProductQuantity = (productId: number, quantity: number) => {
    setSelectedProducts(selectedProducts.map(p => 
      p.product_id === productId ? { ...p, quantity } : p
    ))
  }

  const updateProductPrice = (productId: number, price: number) => {
    setSelectedProducts(selectedProducts.map(p => 
      p.product_id === productId ? { ...p, price } : p
    ))
  }

  const getOrderTotal = () => {
    return selectedProducts.reduce((sum, p) => sum + (p.quantity * p.price), 0)
  }

  useEffect(() => {
    setFormData({ ...formData, total_amount: getOrderTotal() })
  }, [selectedProducts])

  useEffect(() => {
    const debounce = setTimeout(() => {
      if (shop?.id) fetchOrders()
    }, 300)
    return () => clearTimeout(debounce)
  }, [search, filter, shop?.id])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      await api.post('/v1/orders', { 
        ...formData, 
        shop_id: shop?.id,
        items: selectedProducts
      })
      setShowModal(false)
      setFormData({ supplier_id: 0, notes: '', total_amount: 0, status: 'pending' })
      setSelectedProducts([])
      fetchOrders()
    } catch (err) { 
      console.error(err) 
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleUpdateOrder = async (id: number, data: { status?: string; notes?: string; total_amount?: number }) => {
    try {
      await api.put(`/v1/orders/${id}`, data)
      fetchOrders()
    } catch (err) {
      console.error(err)
    }
  }

  const handleView = (order: Order) => {
    setSelectedOrder(order)
    setEditForm({ status: order.status, notes: order.notes || '', total_amount: order.total_amount })
    setShowDetailsModal(true)
  }

  const handleEditOrder = () => {
    setShowEditModal(true)
  }

  const handleSaveEdit = async () => {
    if (!selectedOrder) return
    setIsSubmitting(true)
    try {
      await api.put(`/v1/orders/${selectedOrder.id}`, editForm)
      setShowEditModal(false)
      fetchOrders()
      const updated = { ...selectedOrder, ...editForm }
      setSelectedOrder(updated)
    } catch (err) {
      console.error(err)
    } finally {
      setIsSubmitting(false)
    }
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
      await api.delete(`/v1/orders/${id}`)
      fetchOrders()
    } catch (err) {
      console.error(err)
    }
  }

  const handleBulkDelete = async () => {
    try {
      await api.post('/v1/orders/bulk-delete', { ids: selectedIds })
      setSelectedIds([])
      fetchOrders()
    } catch (err) {
      console.error(err)
    }
  }

  const toggleSelectAll = () => {
    if (selectedIds.length === orders.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(orders.map(o => o.id))
    }
  }

  const toggleSelect = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(i => i !== id))
    } else {
      setSelectedIds([...selectedIds, id])
    }
  }

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'completed': return 'bg-green-100 text-green-700'
      case 'confirmed': return 'bg-blue-100 text-blue-700'
      case 'shipped': return 'bg-purple-100 text-purple-700'
      case 'pending': return 'bg-yellow-100 text-yellow-700'
      case 'cancelled': return 'bg-red-100 text-red-700'
      default: return 'bg-surface-100 text-surface-700'
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const totalOrders = orders.length
  const pendingOrders = orders.filter(o => o.status === 'pending').length
  const completedOrders = orders.filter(o => o.status === 'completed').length
  const totalValue = orders.reduce((sum, o) => sum + (o.total_amount || 0), 0)

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Orders</h1>
          <p className="text-gray-500 mt-1">{orders.length} orders</p>
        </div>
        <button 
          onClick={() => { setSelectedProducts([]); setFormData({ supplier_id: 0, notes: '', total_amount: 0, status: 'pending' }); setShowModal(true); }}
          className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Order
        </button>
      </div>

      {/* Stats */}
      {!isLoading && shop && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <StatCard
            title="Total Orders"
            value={totalOrders}
            variant="default"
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
            title="Completed"
            value={completedOrders}
            variant="success"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
          <StatCard
            title="Total Value"
            value={formatCurrency(totalValue)}
            variant="info"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-xl border border-gray-200 p-4 mb-6">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              placeholder="Search orders..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
            />
          </div>
          <div className="flex gap-2">
            {['all', 'pending', 'confirmed', 'shipped', 'completed', 'cancelled'].map((status) => (
              <button
                key={status}
                onClick={() => setFilter(status)}
                className={`px-4 py-2 rounded-xl text-sm font-medium transition ${
                  filter === status
                    ? 'bg-primary text-white'
                    : 'bg-white text-surface-600 hover:bg-surface-50 border border-surface-200'
                }`}
              >
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </button>
            ))}
          </div>
          {selectedIds.length > 0 && (
            <button
              onClick={() => { setDeleteTarget('bulk'); setShowDeleteConfirm(true); }}
              className="px-4 py-2 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 text-sm font-medium"
            >
              Delete ({selectedIds.length})
            </button>
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
            description="Please select a shop to view orders"
          />
        </Card>
      ) : orders.length === 0 ? (
        <Card className="text-center py-12">
          <EmptyState
            variant="orders"
            title={search ? 'No orders found' : 'No Orders Yet'}
            description={search ? 'Try adjusting your search' : 'Create your first order to track purchases'}
            action={!search ? {
              label: 'New Order',
              onClick: () => { setSelectedProducts([]); setFormData({ supplier_id: 0, notes: '', total_amount: 0, status: 'pending' }); setShowModal(true); },
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
                    <input type="checkbox" checked={selectedIds.length === orders.length && orders.length > 0} onChange={toggleSelectAll} className="w-4 h-4 rounded" />
                  </th>
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
                    <td className="px-4 py-3">
                      <input type="checkbox" checked={selectedIds.includes(order.id)} onChange={() => toggleSelect(order.id)} className="w-4 h-4 rounded" />
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-900">#{order.id}</td>
                    <td className="px-4 py-3 text-gray-600">{order.supplier?.name || `Supplier #${order.supplier_id}`}</td>
                    <td className="px-4 py-3 text-right font-medium text-gray-900">{formatCurrency(order.total_amount)}</td>
                    <td className="px-4 py-3 text-center">
                      <select
                        value={order.status}
                        onChange={(e) => handleUpdateOrder(order.id, { status: e.target.value })}
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
                        <button onClick={() => handleView(order)} className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg" title="View">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                          </svg>
                        </button>
                        <button onClick={() => { setDeleteTarget(order.id); setShowDeleteConfirm(true); }} className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg" title="Delete">
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

      {/* Create Order Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-2xl rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Create Order</h3>
                <button onClick={() => { setShowModal(false); setSelectedProducts([]); }} className="p-2 hover:bg-surface-100 rounded-xl transition-colors">
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="p-6 overflow-y-auto flex-1">
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Supplier *</label>
                  <select 
                    value={formData.supplier_id} 
                    onChange={(e) => setFormData({...formData, supplier_id: parseInt(e.target.value)})}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    required
                  >
                    <option value="">Select Supplier</option>
                    {suppliers.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                  </select>
                </div>

                {/* Product Selection */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="block text-sm font-semibold text-surface-700">Add Products</label>
                    <button
                      type="button"
                      onClick={() => setShowProductList(!showProductList)}
                      className="text-sm text-primary font-medium"
                    >
                      {showProductList ? 'Hide Products' : 'Show Products'}
                    </button>
                  </div>
                  
                  {showProductList && (
                    <div className="border border-surface-200 rounded-xl overflow-hidden">
                      {/* Search in product list */}
                      <div className="p-3 bg-surface-50 border-b border-surface-200">
                        <div className="relative">
                          <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                          </svg>
                          <input 
                            type="text"
                            value={productSearch}
                            onChange={(e) => setProductSearch(e.target.value)}
                            placeholder="Search products..."
                            className="w-full pl-10 pr-4 py-2 bg-white border border-surface-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                          />
                        </div>
                      </div>
                      
                      {/* Products List */}
                      <div className="max-h-64 overflow-y-auto">
                        {products.length === 0 ? (
                          <div className="p-4 text-center text-surface-500">No products available</div>
                        ) : (
                          products.filter(p => p.name.toLowerCase().includes(productSearch.toLowerCase())).slice(0, 20).map(product => {
                            const isAdded = selectedProducts.some(sp => sp.product_id === product.id)
                            return (
                              <div 
                                key={product.id} 
                                className={`p-3 flex items-center justify-between border-b border-surface-100 last:border-0 ${isAdded ? 'bg-primary/5' : 'hover:bg-surface-50'}`}
                              >
                                <div className="flex-1 min-w-0">
                                  <p className="font-medium text-surface-900 truncate">{product.name}</p>
                                  <p className="text-sm text-surface-500">{product.category || 'Uncategorized'}</p>
                                </div>
                                <div className="flex items-center gap-3">
                                  <span className="text-sm font-semibold text-primary">KES {(product.cost_price || product.selling_price || 0).toLocaleString()}</span>
                                  {isAdded ? (
                                    <span className="px-2 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-lg">Added</span>
                                  ) : (
                                    <button
                                      type="button"
                                      onClick={() => addProduct(product)}
                                      className="px-3 py-1 bg-primary text-white text-xs font-medium rounded-lg hover:bg-primary-dark"
                                    >
                                      Add
                                    </button>
                                  )}
                                </div>
                              </div>
                            )
                          })
                        )}
                      </div>
                      
                      {/* Create New Product */}
                      <div className="p-3 bg-surface-50 border-t border-surface-200">
                        <button
                          type="button"
                          onClick={() => setShowNewProductModal(true)}
                          className="w-full flex items-center justify-center gap-2 px-4 py-2 border-2 border-dashed border-surface-300 text-surface-600 rounded-xl hover:border-primary hover:text-primary transition"
                        >
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                          </svg>
                          Create New Product
                        </button>
                      </div>
                    </div>
                  )}
                </div>

                {/* Selected Products */}
                {selectedProducts.length > 0 ? (
                  <div className="border border-surface-200 rounded-xl overflow-hidden">
                    <div className="bg-surface-50 px-4 py-2 border-b border-surface-200 flex items-center justify-between">
                      <span className="text-sm font-semibold text-surface-700">Order Items ({selectedProducts.length})</span>
                      <span className="text-sm font-bold text-primary">KES {getOrderTotal().toLocaleString()}</span>
                    </div>
                    <div className="divide-y divide-surface-100 max-h-64 overflow-y-auto">
                      {selectedProducts.map((item) => (
                        <div key={item.product_id} className="p-4 flex items-center gap-3">
                          <div className="flex-1 min-w-0">
                            <p className="font-medium text-surface-900 truncate">{item.name}</p>
                            <p className="text-sm text-surface-500">KES {item.price.toLocaleString()} each</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-sm text-surface-500">Qty:</span>
                            <input
                              type="number"
                              min="1"
                              value={item.quantity}
                              onChange={(e) => updateProductQuantity(item.product_id, parseInt(e.target.value) || 1)}
                              className="w-20 px-3 py-2 text-center bg-surface-50 border border-surface-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                            />
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-sm text-surface-500">KES</span>
                            <input
                              type="number"
                              min="0"
                              value={item.price}
                              onChange={(e) => updateProductPrice(item.product_id, parseFloat(e.target.value) || 0)}
                              className="w-24 px-3 py-2 text-center bg-surface-50 border border-surface-200 rounded-lg text-sm focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                            />
                          </div>
                          <button
                            type="button"
                            onClick={() => removeProduct(item.product_id)}
                            className="p-2 text-surface-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition"
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      ))}
                    </div>
                    <div className="bg-surface-50 px-4 py-3 border-t border-surface-200 flex justify-between items-center">
                      <span className="font-semibold text-surface-700">Total Amount</span>
                      <span className="text-xl font-bold text-primary">KES {getOrderTotal().toLocaleString()}</span>
                    </div>
                  </div>
                ) : (
                  <div className="border-2 border-dashed border-surface-200 rounded-xl p-8 text-center">
                    <svg className="w-12 h-12 text-surface-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                    <p className="text-surface-500">No products added yet</p>
                    <p className="text-sm text-surface-400">Search and add products above</p>
                  </div>
                )}

                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Notes</label>
                  <textarea 
                    value={formData.notes}
                    onChange={(e) => setFormData({...formData, notes: e.target.value})}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    rows={2}
                    placeholder="Add any notes for this order..."
                  />
                </div>
              </form>
            </div>
            
              <div className="p-6 border-t border-surface-100 bg-white sticky bottom-0">
              <div className="flex gap-3">
                <button type="button" onClick={() => { setShowModal(false); setSelectedProducts([]); }} className="flex-1 px-4 py-3 border border-surface-200 text-surface-700 rounded-xl hover:bg-surface-50 font-medium">Cancel</button>
                <button 
                  type="submit"
                  disabled={isSubmitting || !formData.supplier_id || selectedProducts.length === 0} 
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  {isSubmitting ? 'Creating...' : 'Create Order'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create New Product Modal */}
      {showNewProductModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[60] flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-md rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Create New Product</h3>
                <button
                  onClick={() => setShowNewProductModal(false)}
                  className="p-2 hover:bg-surface-100 rounded-xl transition-colors"
                >
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <form onSubmit={handleCreateProduct} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Product Name *</label>
                <input
                  type="text"
                  value={newProductForm.name}
                  onChange={(e) => setNewProductForm({ ...newProductForm, name: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  placeholder="Enter product name"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Category</label>
                  <input
                    type="text"
                    value={newProductForm.category}
                    onChange={(e) => setNewProductForm({ ...newProductForm, category: e.target.value })}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    placeholder="e.g., Electronics"
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Unit</label>
                  <select
                    value={newProductForm.unit}
                    onChange={(e) => setNewProductForm({ ...newProductForm, unit: e.target.value })}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  >
                    <option value="pcs">Pieces</option>
                    <option value="kg">Kilograms</option>
                    <option value="g">Grams</option>
                    <option value="L">Liters</option>
                    <option value="ml">Milliliters</option>
                    <option value="pack">Pack</option>
                    <option value="box">Box</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Cost Price *</label>
                  <div className="relative">
                    <span className="absolute left-4 top-1/2 -translate-y-1/2 text-surface-400">KES</span>
                    <input
                      type="number"
                      value={newProductForm.cost_price}
                      onChange={(e) => setNewProductForm({ ...newProductForm, cost_price: parseFloat(e.target.value) || 0 })}
                      className="w-full pl-16 pr-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                      required
                      min="0"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Selling Price *</label>
                  <div className="relative">
                    <span className="absolute left-4 top-1/2 -translate-y-1/2 text-surface-400">KES</span>
                    <input
                      type="number"
                      value={newProductForm.selling_price}
                      onChange={(e) => setNewProductForm({ ...newProductForm, selling_price: parseFloat(e.target.value) || 0 })}
                      className="w-full pl-16 pr-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                      required
                      min="0"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Initial Stock</label>
                <input
                  type="number"
                  value={newProductForm.current_stock}
                  onChange={(e) => setNewProductForm({ ...newProductForm, current_stock: parseInt(e.target.value) || 0 })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  min="0"
                />
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowNewProductModal(false)}
                  className="flex-1 px-4 py-3 border border-surface-200 text-surface-700 rounded-xl hover:bg-surface-50 font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50 font-medium"
                >
                  {isSubmitting ? 'Creating...' : 'Create & Add'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Details Modal */}
      {showDetailsModal && selectedOrder && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-lg rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Order Details</h3>
                <button onClick={() => setShowDetailsModal(false)} className="p-2 hover:bg-surface-100 rounded-xl transition-colors">
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="overflow-y-auto flex-1">
              <div className="p-6 space-y-4">
                {/* Order Header */}
                <div className="text-center pb-4 border-b border-surface-100">
                  <h4 className="text-2xl font-bold text-surface-900">#{selectedOrder.id}</h4>
                  <p className="text-surface-500">{new Date(selectedOrder.created_at).toLocaleDateString()}</p>
                </div>

                {/* Order Info */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-3 bg-surface-50 rounded-xl">
                    <p className="text-xs text-surface-500 mb-1">Supplier</p>
                    <p className="font-semibold text-surface-900">{selectedOrder.supplier?.name || `Supplier #${selectedOrder.supplier_id}`}</p>
                  </div>
                  <div className="p-3 bg-surface-50 rounded-xl">
                    <p className="text-xs text-surface-500 mb-1">Total Amount</p>
                    <p className="font-semibold text-primary text-lg">{formatCurrency(selectedOrder.total_amount)}</p>
                  </div>
                </div>

                {/* Status */}
                <div className="p-3 bg-surface-50 rounded-xl">
                  <p className="text-xs text-surface-500 mb-2">Status</p>
                  <span className={`inline-flex px-3 py-1.5 rounded-full text-sm font-medium ${getStatusColor(selectedOrder.status)}`}>
                    {selectedOrder.status}
                  </span>
                </div>

                {/* Notes */}
                {selectedOrder.notes && (
                  <div className="p-3 bg-surface-50 rounded-xl">
                    <p className="text-xs text-surface-500 mb-1">Notes</p>
                    <p className="text-surface-700">{selectedOrder.notes}</p>
                  </div>
                )}

                {/* Order Items */}
                <div>
                  <p className="text-sm font-semibold text-surface-700 mb-2">Order Items</p>
                  {selectedOrder.items && selectedOrder.items.length > 0 ? (
                    <div className="border border-surface-200 rounded-xl overflow-hidden">
                      <table className="w-full">
                        <thead className="bg-surface-50 border-b border-surface-200">
                          <tr>
                            <th className="px-3 py-2 text-left text-xs font-medium text-surface-500">Product</th>
                            <th className="px-3 py-2 text-center text-xs font-medium text-surface-500">Qty</th>
                            <th className="px-3 py-2 text-right text-xs font-medium text-surface-500">Price</th>
                            <th className="px-3 py-2 text-right text-xs font-medium text-surface-500">Total</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-surface-100">
                          {selectedOrder.items.map((item, idx) => (
                            <tr key={item.id || idx}>
                              <td className="px-3 py-2 text-sm text-surface-900">{item.product?.name || `Product #${item.product_id}`}</td>
                              <td className="px-3 py-2 text-sm text-surface-600 text-center">{item.quantity}</td>
                              <td className="px-3 py-2 text-sm text-surface-600 text-right">{formatCurrency(item.price)}</td>
                              <td className="px-3 py-2 text-sm font-medium text-surface-900 text-right">{formatCurrency(item.quantity * item.price)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="p-4 text-center text-surface-500 bg-surface-50 rounded-xl">
                      No items in this order
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="p-4 border-t border-surface-100 bg-white sticky bottom-0">
              <div className="flex gap-3">
                <button 
                  onClick={() => { setDeleteTarget(selectedOrder.id); setShowDetailsModal(false); setShowDeleteConfirm(true); }} 
                  className="flex-1 px-4 py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 font-medium"
                >
                  Delete
                </button>
                <button 
                  onClick={handleEditOrder}
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark font-medium"
                >
                  Edit Order
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Order Modal */}
      {showEditModal && selectedOrder && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[60] flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-md rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Edit Order</h3>
                <button onClick={() => setShowEditModal(false)} className="p-2 hover:bg-surface-100 rounded-xl transition-colors">
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Status</label>
                <select 
                  value={editForm.status}
                  onChange={(e) => setEditForm({ ...editForm, status: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                >
                  <option value="pending">Pending</option>
                  <option value="confirmed">Confirmed</option>
                  <option value="shipped">Shipped</option>
                  <option value="completed">Completed</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Total Amount</label>
                <div className="relative">
                  <span className="absolute left-4 top-1/2 -translate-y-1/2 text-surface-400">KES</span>
                  <input
                    type="number"
                    value={editForm.total_amount}
                    onChange={(e) => setEditForm({ ...editForm, total_amount: parseFloat(e.target.value) || 0 })}
                    className="w-full pl-16 pr-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Notes</label>
                <textarea 
                  value={editForm.notes}
                  onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  rows={3}
                  placeholder="Add notes..."
                />
              </div>
            </div>
            <div className="p-6 border-t border-surface-100">
              <div className="flex gap-3">
                <button onClick={() => setShowEditModal(false)} className="flex-1 px-4 py-3 border border-surface-200 text-surface-700 rounded-xl hover:bg-surface-50 font-medium">
                  Cancel
                </button>
                <button onClick={handleSaveEdit} disabled={isSubmitting} className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50 font-medium">
                  {isSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
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
              <p className="text-gray-500 mb-6">{deleteTarget === 'bulk' ? `Delete ${selectedIds.length} orders? This cannot be undone.` : 'This action cannot be undone.'}</p>
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
