import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { SkeletonGrid } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import { PageTransition } from '@/components/common/PageTransition'
import type { Product } from '@/api/types'

export default function Products() {
  const shop = useAuthStore((state) => state.shop)
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [search, setSearch] = useState('')
  const [activeCategory, setActiveCategory] = useState('All')
  const [categories, setCategories] = useState<string[]>(['All'])
  const [error, setError] = useState('')
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid')
  const [formData, setFormData] = useState({
    name: '',
    category: '',
    unit: 'pcs',
    cost_price: 0,
    selling_price: 0,
    current_stock: 0,
    low_stock_threshold: 10,
    barcode: '',
    image_url: ''
  })
  const [imagePreview, setImagePreview] = useState('')
  const [uploadingImage, setUploadingImage] = useState(false)
  const [selectedProducts, setSelectedProducts] = useState<number[]>([])
  const [showViewModal, setShowViewModal] = useState(false)
  const [viewingProduct, setViewingProduct] = useState<Product | null>(null)
  const [selectAll, setSelectAll] = useState(false)

  useEffect(() => {
    if (shop?.id) {
      fetchCategories()
    }
  }, [shop?.id])

  useEffect(() => {
    if (shop?.id) {
      const debounce = setTimeout(() => {
        fetchProducts()
      }, 300)
      return () => clearTimeout(debounce)
    } else {
      setIsLoading(false)
    }
  }, [shop?.id, search, activeCategory])

  const fetchProducts = async () => {
    if (!shop?.id) return
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      if (search) params.append('search', search)
      if (activeCategory !== 'All') params.append('category', activeCategory)
      
      const response = await api.get(`/v1/products?${params}`)
      const responseData = response.data
      const productsData = responseData?.data || responseData || []
      setProducts(Array.isArray(productsData) ? productsData : [])
    } catch (err) {
      setError('Failed to load products')
      console.error(err)
      setProducts([])
    } finally {
      setIsLoading(false)
    }
  }

  const fetchCategories = async () => {
    try {
      const response = await api.get('/v1/products/categories')
      const responseData = response.data
      const cats = responseData?.categories || responseData?.data || responseData || []
      setCategories(['All', ...cats])
    } catch (err) {
      console.error(err)
      setCategories(['All'])
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    
    try {
      if (editingProduct) {
        await api.put(`/v1/products/${editingProduct.id}`, formData)
      } else {
        await api.post('/v1/products', { ...formData, shop_id: shop?.id })
      }
      setShowModal(false)
      setEditingProduct(null)
      resetForm()
      fetchProducts()
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      setError(error.response?.data?.error || 'Failed to save product')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this product?')) return
    
    try {
      await api.delete(`/v1/products/${id}`)
      fetchProducts()
    } catch (err) {
      setError('Failed to delete product')
    }
  }

  const handleEdit = (product: Product) => {
    setEditingProduct(product)
    setFormData({
      name: product.name,
      category: product.category || '',
      unit: product.unit,
      cost_price: product.cost_price,
      selling_price: product.selling_price,
      current_stock: product.current_stock,
      low_stock_threshold: product.low_stock_threshold,
      barcode: product.barcode || '',
      image_url: product.image_url || ''
    })
    setImagePreview(product.image_url || '')
    setShowModal(true)
  }

  const resetForm = () => {
    setFormData({
      name: '',
      category: '',
      unit: 'pcs',
      cost_price: 0,
      selling_price: 0,
      current_stock: 0,
      low_stock_threshold: 10,
      barcode: '',
      image_url: ''
    })
    setImagePreview('')
  }

  const openAddModal = () => {
    resetForm()
    setEditingProduct(null)
    setShowModal(true)
  }

  const handleView = (product: Product) => {
    setViewingProduct(product)
    setShowViewModal(true)
  }

  const toggleSelectProduct = (id: number) => {
    setSelectedProducts(prev => 
      prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]
    )
  }

  const toggleSelectAll = () => {
    if (selectAll) {
      setSelectedProducts([])
    } else {
      setSelectedProducts(products.map(p => p.id))
    }
    setSelectAll(!selectAll)
  }

  const handleBulkDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${selectedProducts.length} products?`)) return
    
    try {
      await Promise.all(selectedProducts.map(id => api.delete(`/v1/products/${id}`)))
      setSelectedProducts([])
      setSelectAll(false)
      fetchProducts()
    } catch (err) {
      setError('Failed to delete some products')
    }
  }

  const exportSelected = () => {
    const selectedData = products.filter(p => selectedProducts.includes(p.id))
    const csv = [
      ['Name', 'Category', 'Cost Price', 'Selling Price', 'Stock', 'Unit', 'Barcode'].join(','),
      ...selectedData.map(p => [
        `"${p.name}"`,
        `"${p.category || ''}"`,
        p.cost_price,
        p.selling_price,
        p.current_stock,
        p.unit,
        `"${p.barcode || ''}"`
      ].join(','))
    ].join('\n')
    
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'products_export.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploadingImage(true)
    try {
      const reader = new FileReader()
      reader.onloadend = () => {
        const base64 = reader.result as string
        setFormData({ ...formData, image_url: base64 })
        setImagePreview(base64)
        setUploadingImage(false)
      }
      reader.readAsDataURL(file)
    } catch (err) {
      setError('Failed to upload image')
      setUploadingImage(false)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const getStockStatus = (product: Product) => {
    if (product.current_stock === 0) return { color: 'bg-red-500', text: 'text-red-600', bg: 'bg-red-50', label: 'Out of Stock' }
    if (product.current_stock <= product.low_stock_threshold) return { color: 'bg-amber-500', text: 'text-amber-600', bg: 'bg-amber-50', label: 'Low Stock' }
    return { color: 'bg-green-500', text: 'text-green-600', bg: 'bg-green-50', label: 'In Stock' }
  }

  return (
    <div className="-mx-4 md:-mx-6">
      {/* Header */}
      <div className="px-4 md:px-6 pb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-surface-900">Products</h1>
            <p className="text-surface-500 mt-1">
              {products.length} {products.length === 1 ? 'item' : 'items'} in inventory
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex bg-surface-100 rounded-xl p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-lg transition-all ${viewMode === 'grid' ? 'bg-white shadow-sm text-primary' : 'text-surface-500 hover:text-surface-700'}`}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
              </button>
              <button
                onClick={() => setViewMode('table')}
                className={`p-2 rounded-lg transition-all ${viewMode === 'table' ? 'bg-white shadow-sm text-primary' : 'text-surface-500 hover:text-surface-700'}`}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                </svg>
              </button>
            </div>
            <Button onClick={openAddModal} leftIcon={
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            }>
              Add Product
            </Button>
          </div>
        </div>

        {/* Search Bar */}
        <div className="relative mb-4">
          <svg className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search products by name or barcode..."
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
        {products.length > 0 && (
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
                  {selectedProducts.length > 0 ? `${selectedProducts.length} selected` : 'Select all'}
                </span>
              </label>
            </div>
            {selectedProducts.length > 0 && (
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

        {/* Category Tabs */}
        <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-hide">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap transition-all ${
                activeCategory === cat
                  ? 'bg-primary text-white shadow-md shadow-primary/25'
                  : 'bg-white text-surface-600 hover:bg-surface-50 border border-surface-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="mx-4 md:mx-6 mb-4 p-4 bg-red-50 text-red-600 rounded-xl">
          {error}
        </div>
      )}

      {/* Content */}
      <div className="px-4 md:px-6">
        <PageTransition>
          {isLoading ? (
            <SkeletonGrid columns={4} items={8} />
          ) : !shop ? (
            <Card className="text-center py-12">
              <EmptyState
                variant="generic"
                title="No Shop Selected"
                description="Please select a shop to view products"
              />
            </Card>
          ) : products.length === 0 ? (
            <Card className="text-center py-12">
              <EmptyState
                variant={search || activeCategory !== 'All' ? 'search' : 'products'}
                title={search || activeCategory !== 'All' ? 'No products found' : 'No Products Yet'}
                description={search || activeCategory !== 'All' 
                  ? 'Try adjusting your search or filters' 
                  : 'Add your first product to start selling'}
                action={!search && activeCategory === 'All' ? {
                  label: 'Add Your First Product',
                  onClick: openAddModal,
                } : undefined}
              />
            </Card>
          ) : viewMode === 'grid' ? (
          /* Product Grid */
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            {products.map((product) => {
              const stock = getStockStatus(product)
              return (
                <Card 
                  key={product.id} 
                  hover 
                  className="group relative overflow-hidden"
                >
                  {/* Checkbox */}
                  <div className="absolute top-3 left-3 z-10">
                    <input
                      type="checkbox"
                      checked={selectedProducts.includes(product.id)}
                      onChange={(e) => {
                        e.stopPropagation()
                        toggleSelectProduct(product.id)
                      }}
                      className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                    />
                  </div>
                  
                  {/* Stock Badge */}
                  <div className={`absolute top-3 right-3 px-2 py-1 ${stock.bg} ${stock.text} text-xs font-semibold rounded-full z-10`}>
                    {product.current_stock} {product.unit}
                  </div>
                  
                  {/* Product Image */}
                  <div className="aspect-square bg-surface-50 rounded-xl mb-3 overflow-hidden relative">
                    {product.image_url ? (
                      <img src={product.image_url} alt={product.name} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <svg className="w-12 h-12 text-surface-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                        </svg>
                      </div>
                    )}
                    
                    {/* Quick Action Overlay */}
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                      <button 
                        onClick={(e) => {
                          e.stopPropagation()
                          handleView(product)
                        }}
                        className="p-2 bg-white rounded-full hover:bg-primary hover:text-white transition-colors"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                        </svg>
                      </button>
                      <button 
                        onClick={(e) => {
                          e.stopPropagation()
                          handleEdit(product)
                        }}
                        className="p-2 bg-white rounded-full hover:bg-primary hover:text-white transition-colors"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                        </svg>
                      </button>
                      <Link 
                        to={`/sales/new?product_id=${product.id}`}
                        onClick={(e) => e.stopPropagation()}
                        className="p-2 bg-white rounded-full hover:bg-primary hover:text-white transition-colors"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
                        </svg>
                      </Link>
                    </div>
                  </div>
                  
                  {/* Product Info */}
                  <div className="cursor-pointer" onClick={() => handleView(product)}>
                    <h3 className="font-semibold text-surface-900 truncate">{product.name}</h3>
                    <p className="text-sm text-surface-500 truncate">{product.category || 'Uncategorized'}</p>
                    <div className="flex items-center justify-between mt-2">
                      <span className="text-lg font-bold text-primary">{formatCurrency(product.selling_price)}</span>
                      {product.barcode && (
                        <span className="text-xs text-surface-400 font-mono">{product.barcode}</span>
                      )}
                    </div>
                  </div>
                </Card>
              )
            })}
          </div>
        ) : (
          /* Table View */
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
                    <th className="px-4 py-3 text-left text-xs font-semibold text-surface-500 uppercase">Category</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Cost</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Price</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Stock</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-surface-500 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-100">
                  {products.map((product) => {
                    const stock = getStockStatus(product)
                    return (
                      <tr key={product.id} className="hover:bg-surface-50 transition-colors">
                        <td className="px-4 py-3">
                          <input
                            type="checkbox"
                            checked={selectedProducts.includes(product.id)}
                            onChange={() => toggleSelectProduct(product.id)}
                            className="w-4 h-4 rounded border-surface-300 text-primary focus:ring-primary"
                          />
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 bg-surface-50 rounded-xl flex items-center justify-center overflow-hidden">
                              {product.image_url ? (
                                <img src={product.image_url} alt={product.name} className="w-full h-full object-cover" />
                              ) : (
                                <svg className="w-5 h-5 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                                </svg>
                              )}
                            </div>
                            <div>
                              <p className="font-medium text-surface-900">{product.name}</p>
                              {product.barcode && (
                                <p className="text-xs text-surface-400 font-mono">{product.barcode}</p>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span className="px-2 py-1 bg-surface-100 rounded-lg text-sm text-surface-600">
                            {product.category || 'Uncategorized'}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right text-surface-600">
                          {formatCurrency(product.cost_price)}
                        </td>
                        <td className="px-4 py-3 text-right font-semibold text-surface-900">
                          {formatCurrency(product.selling_price)}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <span className={`px-2 py-1 ${stock.bg} ${stock.text} rounded-lg text-sm font-medium`}>
                            {product.current_stock} {product.unit}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              onClick={() => handleView(product)}
                              className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                              </svg>
                            </button>
                            <Link
                              to={`/sales/new?product_id=${product.id}`}
                              className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
                              </svg>
                            </Link>
                            <button
                              onClick={() => handleEdit(product)}
                              className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                              </svg>
                            </button>
                            <button
                              onClick={() => handleDelete(product.id)}
                              className="p-2 text-surface-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                            </button>
                          </div>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </Card>
        )}
        </PageTransition>
      </div>

      {/* Add/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-lg rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">
                  {editingProduct ? 'Edit Product' : 'Add Product'}
                </h3>
                <button
                  onClick={() => { setShowModal(false); setEditingProduct(null); }}
                  className="p-2 hover:bg-surface-100 rounded-xl transition-colors"
                >
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Product Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
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
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    list="categories"
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    placeholder="Select or type"
                  />
                  <datalist id="categories">
                    {categories.filter(c => c !== 'All').map((cat) => (
                      <option key={cat} value={cat} />
                    ))}
                  </datalist>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Unit</label>
                  <select
                    value={formData.unit}
                    onChange={(e) => setFormData({ ...formData, unit: e.target.value })}
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
                      value={formData.cost_price}
                      onChange={(e) => setFormData({ ...formData, cost_price: parseFloat(e.target.value) || 0 })}
                      className="w-full pl-16 pr-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                      required
                      min="0"
                      step="1"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Selling Price *</label>
                  <div className="relative">
                    <span className="absolute left-4 top-1/2 -translate-y-1/2 text-surface-400">KES</span>
                    <input
                      type="number"
                      value={formData.selling_price}
                      onChange={(e) => setFormData({ ...formData, selling_price: parseFloat(e.target.value) || 0 })}
                      className="w-full pl-16 pr-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                      required
                      min="0"
                      step="1"
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Current Stock *</label>
                  <input
                    type="number"
                    value={formData.current_stock}
                    onChange={(e) => setFormData({ ...formData, current_stock: parseInt(e.target.value) || 0 })}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    required
                    min="0"
                  />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-surface-700 mb-2">Low Stock Alert</label>
                  <input
                    type="number"
                    value={formData.low_stock_threshold}
                    onChange={(e) => setFormData({ ...formData, low_stock_threshold: parseInt(e.target.value) || 0 })}
                    className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                    min="0"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Barcode</label>
                <input
                  type="text"
                  value={formData.barcode}
                  onChange={(e) => setFormData({ ...formData, barcode: e.target.value })}
                  className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none"
                  placeholder="Scan or enter barcode"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Product Image</label>
                <div className="flex items-center gap-4">
                  <div className="w-20 h-20 bg-surface-100 rounded-xl flex items-center justify-center overflow-hidden">
                    {imagePreview || formData.image_url ? (
                      <img src={imagePreview || formData.image_url} alt="Preview" className="w-full h-full object-cover" />
                    ) : (
                      <svg className="w-8 h-8 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    )}
                  </div>
                  <label className="flex-1 cursor-pointer">
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleImageUpload}
                      className="hidden"
                    />
                    <span className="inline-flex items-center px-4 py-2.5 bg-surface-100 hover:bg-surface-200 text-surface-700 rounded-xl text-sm font-medium transition">
                      {uploadingImage ? 'Uploading...' : 'Choose Image'}
                    </span>
                  </label>
                </div>
              </div>

              <div className="flex gap-3 pt-4">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => { setShowModal(false); setEditingProduct(null); }}
                  className="flex-1"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  className="flex-1"
                >
                  {editingProduct ? 'Update Product' : 'Add Product'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View Details Modal */}
      {showViewModal && viewingProduct && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-lg rounded-t-3xl md:rounded-3xl shadow-2xl max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-surface-100 sticky top-0 bg-white z-10">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-surface-900">Product Details</h3>
                <button
                  onClick={() => { setShowViewModal(false); setViewingProduct(null); }}
                  className="p-2 hover:bg-surface-100 rounded-xl transition-colors"
                >
                  <svg className="w-6 h-6 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="p-6 space-y-6">
              {/* Product Image */}
              <div className="aspect-square bg-surface-50 rounded-2xl overflow-hidden">
                {viewingProduct.image_url ? (
                  <img src={viewingProduct.image_url} alt={viewingProduct.name} className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <svg className="w-24 h-24 text-surface-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  </div>
                )}
              </div>

              {/* Product Info */}
              <div>
                <h2 className="text-2xl font-bold text-surface-900">{viewingProduct.name}</h2>
                <p className="text-surface-500 mt-1">{viewingProduct.category || 'Uncategorized'}</p>
              </div>

              {/* Stock Status */}
              <div className={`p-4 rounded-xl ${getStockStatus(viewingProduct).bg}`}>
                <div className="flex items-center justify-between">
                  <span className={`font-semibold ${getStockStatus(viewingProduct).text}`}>
                    {getStockStatus(viewingProduct).label}
                  </span>
                  <span className="text-2xl font-bold text-surface-900">
                    {viewingProduct.current_stock} {viewingProduct.unit}
                  </span>
                </div>
                {viewingProduct.low_stock_threshold > 0 && (
                  <p className="text-sm text-surface-500 mt-1">
                    Low stock alert at {viewingProduct.low_stock_threshold} {viewingProduct.unit}
                  </p>
                )}
              </div>

              {/* Pricing */}
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Cost Price</p>
                  <p className="text-lg font-bold text-surface-900">{formatCurrency(viewingProduct.cost_price)}</p>
                </div>
                <div className="p-4 bg-surface-50 rounded-xl">
                  <p className="text-sm text-surface-500">Selling Price</p>
                  <p className="text-lg font-bold text-primary">{formatCurrency(viewingProduct.selling_price)}</p>
                </div>
              </div>

              {/* Additional Info */}
              <div className="space-y-3">
                {viewingProduct.barcode && (
                  <div className="flex items-center justify-between p-3 bg-surface-50 rounded-xl">
                    <span className="text-sm text-surface-500">Barcode</span>
                    <span className="font-mono text-surface-900">{viewingProduct.barcode}</span>
                  </div>
                )}
                <div className="flex items-center justify-between p-3 bg-surface-50 rounded-xl">
                  <span className="text-sm text-surface-500">Unit</span>
                  <span className="text-surface-900">{viewingProduct.unit}</span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-4">
                <Button
                  variant="secondary"
                  onClick={() => {
                    setShowViewModal(false)
                    setViewingProduct(null)
                  }}
                  className="flex-1"
                >
                  Close
                </Button>
                <Button
                  variant="primary"
                  onClick={() => {
                    setShowViewModal(false)
                    handleEdit(viewingProduct)
                  }}
                  className="flex-1"
                >
                  Edit Product
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
