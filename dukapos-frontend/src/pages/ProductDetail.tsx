import { useState, useEffect, useRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { Card, Button, Modal, Loader } from '@/components/common'
import { ProductForm, ProductCard, type ProductFormData } from '@/components/products'
import type { Product } from '@/api/types'

interface ProductDetailProps {
  productId?: number
}

export function ProductDetail({ productId: propProductId }: ProductDetailProps) {
  const params = useParams()
  const navigate = useNavigate()
  const productIdFromParams = params.id
  const isNewProduct = !productIdFromParams || productIdFromParams === 'new'
  const productId = isNewProduct ? null : (propProductId || Number(productIdFromParams))
  
  const token = useAuthStore((state) => state.token)
  const shop = useAuthStore((state) => state.shop)
  const [product, setProduct] = useState<Product | null>(null)
  const [relatedProducts, setRelatedProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState('')
  const [imagePreview, setImagePreview] = useState<string>('')
  const [selectedImageUrl, setSelectedImageUrl] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onloadend = () => {
        const result = reader.result as string
        setImagePreview(result)
        setSelectedImageUrl(result)
      }
      reader.readAsDataURL(file)
    }
  }

  const clearImage = () => {
    setImagePreview('')
    setSelectedImageUrl('')
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  useEffect(() => {
    if (productId) {
      fetchProduct()
    } else {
      setLoading(false)
    }
  }, [productId])

  const fetchProduct = async () => {
    if (!token) return
    
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/v1/products/${productId}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      if (!response.ok) {
        throw new Error('Failed to fetch product')
      }
      
      const data = await response.json()
      setProduct(data.data)
      
      // Fetch related products (same category)
      if (data.data.category) {
        const relatedRes = await fetch(`/api/v1/products?category=${data.data.category}`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
        if (relatedRes.ok) {
          const relatedData = await relatedRes.json()
          setRelatedProducts(relatedData.data.filter((p: Product) => p.id !== productId).slice(0, 4))
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdate = async (formData: ProductFormData) => {
    if (!token || !product) return
    
    try {
      const response = await fetch(`/api/v1/products/${product.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      })
      
      if (!response.ok) {
        throw new Error('Failed to update product')
      }
      
      const data = await response.json()
      setProduct(data.data)
      setIsEditing(false)
    } catch (err) {
      throw err
    }
  }

  const handleDelete = async () => {
    if (!token || !product || deleteConfirm !== product.name) return
    
    try {
      const response = await fetch(`/api/v1/products/${product.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      if (!response.ok) {
        throw new Error('Failed to delete product')
      }
      
      navigate('/products')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    } finally {
      setIsDeleting(false)
      setDeleteConfirm('')
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const calculateProfit = () => {
    if (!product) return 0
    return product.selling_price - product.cost_price
  }

  const calculateMargin = () => {
    if (!product || product.cost_price === 0) return 0
    return ((product.selling_price - product.cost_price) / product.selling_price * 100).toFixed(1)
  }

  const handleCreate = async (formData: ProductFormData) => {
    if (!token || !shop) return
    
    try {
      const response = await fetch('/api/v1/products', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ ...formData, shop_id: shop.id })
      })
      
      if (!response.ok) {
        throw new Error('Failed to create product')
      }
      
      const data = await response.json()
      navigate(`/products/${data.data.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create product')
    }
  }

  if (isNewProduct) {
    return (
      <div className="p-6">
        <div className="mb-6">
          <Link to="/products" className="text-primary hover:underline text-sm flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Products
          </Link>
          <h1 className="text-2xl md:text-3xl font-bold text-surface-900 mt-4">Add New Product</h1>
          <p className="text-surface-500 mt-1">Create a new product in your inventory</p>
        </div>
        
        <Card>
          <form onSubmit={async (e) => {
            e.preventDefault()
            const formData = new FormData(e.currentTarget)
            await handleCreate({
              name: formData.get('name') as string,
              category: formData.get('category') as string,
              unit: formData.get('unit') as string,
              cost_price: Number(formData.get('cost_price')),
              selling_price: Number(formData.get('selling_price')),
              current_stock: Number(formData.get('current_stock')),
              low_stock_threshold: Number(formData.get('low_stock_threshold')),
              barcode: formData.get('barcode') as string,
              image_url: selectedImageUrl
            })
          }} className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">Product Name *</label>
              <input name="name" required className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" placeholder="Enter product name" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Category</label>
                <input name="category" className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" placeholder="Category" />
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Unit</label>
                <select name="unit" className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none">
                  <option value="pcs">Pieces</option>
                  <option value="kg">Kilograms</option>
                  <option value="g">Grams</option>
                  <option value="L">Liters</option>
                  <option value="ml">Milliliters</option>
                </select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Cost Price *</label>
                <input name="cost_price" type="number" required className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" placeholder="0" />
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Selling Price *</label>
                <input name="selling_price" type="number" required className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" placeholder="0" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Stock</label>
                <input name="current_stock" type="number" defaultValue="0" className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" />
              </div>
              <div>
                <label className="block text-sm font-semibold text-surface-700 mb-2">Low Stock Alert</label>
                <input name="low_stock_threshold" type="number" defaultValue="10" className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" />
              </div>
            </div>
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">Barcode</label>
              <input name="barcode" className="w-full px-4 py-3 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" placeholder="Barcode (optional)" />
            </div>
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">Product Image</label>
              <div className="flex flex-col items-center gap-4">
                <div className="w-32 h-32 bg-surface-100 rounded-xl overflow-hidden flex items-center justify-center border-2 border-dashed border-surface-300">
                  {imagePreview ? (
                    <img src={imagePreview} alt="Preview" className="w-full h-full object-cover" />
                  ) : (
                    <svg className="w-12 h-12 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  )}
                </div>
                <div className="flex gap-2">
                  <button type="button" onClick={() => fileInputRef.current?.click()} className="px-3 py-2 bg-surface-100 text-surface-700 text-sm rounded-lg hover:bg-surface-200 flex items-center gap-1">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    Upload
                  </button>
                  {imagePreview && (
                    <button type="button" onClick={clearImage} className="px-3 py-2 text-red-600 text-sm hover:bg-red-50 rounded-lg">
                      Remove
                    </button>
                  )}
                </div>
                <input ref={fileInputRef} type="file" accept="image/*" onChange={handleFileSelect} className="hidden" />
              </div>
            </div>
            <div className="flex gap-3 pt-4">
              <button type="button" onClick={() => navigate('/products')} className="flex-1 px-4 py-3 border border-surface-200 text-surface-700 rounded-xl hover:bg-surface-50 font-medium">Cancel</button>
              <button type="submit" className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark font-medium">Create Product</button>
            </div>
          </form>
        </Card>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader size="lg" />
      </div>
    )
  }

  if (error || !product) {
    return (
      <div className="text-center py-12">
        <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <h3 className="font-semibold text-gray-900 mb-2">Error loading product</h3>
        <p className="text-gray-500 mb-4">{error || 'Product not found'}</p>
        <Link
          to="/products"
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg"
        >
          Back to Products
        </Link>
      </div>
    )
  }

  const isLowStock = product.current_stock <= product.low_stock_threshold

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-gray-500">
        <Link to="/products" className="hover:text-primary">Products</Link>
        <span>/</span>
        <span className="text-gray-900">{product.name}</span>
      </nav>

      {/* Header */}
      <div className="flex flex-col md:flex-row gap-6">
        {/* Image */}
        <div className="w-full md:w-64">
          <div className="aspect-square bg-gray-100 rounded-2xl overflow-hidden">
            {product.image_url ? (
              <img src={product.image_url} alt={product.name} className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <svg className="w-16 h-16 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                </svg>
              </div>
            )}
          </div>
        </div>

        {/* Details */}
        <div className="flex-1">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{product.name}</h1>
              {product.category && (
                <span className="inline-block mt-2 px-3 py-1 bg-gray-100 text-gray-600 rounded-full text-sm">
                  {product.category}
                </span>
              )}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setIsEditing(true)}>
                Edit
              </Button>
              <Button variant="danger" onClick={() => setIsDeleting(true)}>
                Delete
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-6">
            <Card padding="sm">
              <p className="text-xs text-gray-500">Selling Price</p>
              <p className="text-lg font-bold text-gray-900">{formatCurrency(product.selling_price)}</p>
            </Card>
            <Card padding="sm">
              <p className="text-xs text-gray-500">Cost Price</p>
              <p className="text-lg font-bold text-gray-900">{formatCurrency(product.cost_price)}</p>
            </Card>
            <Card padding="sm">
              <p className="text-xs text-gray-500">Profit</p>
              <p className="text-lg font-bold text-green-600">{formatCurrency(calculateProfit())}</p>
            </Card>
            <Card padding="sm">
              <p className="text-xs text-gray-500">Margin</p>
              <p className="text-lg font-bold text-gray-900">{calculateMargin()}%</p>
            </Card>
          </div>

          {/* Stock Status */}
          <div className={`mt-6 p-4 rounded-xl ${isLowStock ? 'bg-red-50 border border-red-200' : 'bg-green-50 border border-green-200'}`}>
            <div className="flex items-center justify-between">
              <div>
                <p className={`font-semibold ${isLowStock ? 'text-red-700' : 'text-green-700'}`}>
                  {isLowStock ? 'Low Stock!' : 'In Stock'}
                </p>
                <p className="text-sm text-gray-600">
                  {product.current_stock} {product.unit} available
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-500">Low Stock Alert</p>
                <p className="font-medium text-gray-700">{product.low_stock_threshold} {product.unit}</p>
              </div>
            </div>
            <div className="mt-3 h-2 bg-gray-200 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full ${isLowStock ? 'bg-red-500' : 'bg-green-500'}`}
                style={{ width: `${Math.min((product.current_stock / (product.low_stock_threshold * 3)) * 100, 100)}%` }}
              />
            </div>
          </div>

          {/* Barcode */}
          {product.barcode && (
            <div className="mt-4 flex items-center gap-3">
              <span className="text-sm text-gray-500">Barcode:</span>
              <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">{product.barcode}</span>
            </div>
          )}
        </div>
      </div>

      {/* Related Products */}
      {relatedProducts.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Related Products</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {relatedProducts.map(p => (
              <ProductCard key={p.id} product={p} compact />
            ))}
          </div>
        </div>
      )}

      {/* Edit Modal */}
      <Modal isOpen={isEditing} onClose={() => setIsEditing(false)} title="Edit Product">
        <ProductForm
          product={product}
          categories={product.category ? [product.category] : []}
          onSubmit={handleUpdate}
          onCancel={() => setIsEditing(false)}
        />
      </Modal>

      {/* Delete Confirmation */}
      <Modal isOpen={isDeleting} onClose={() => { setIsDeleting(false); setDeleteConfirm('') }} title="Delete Product">
        <div className="space-y-4">
          <div className="p-4 bg-red-50 rounded-xl">
            <p className="text-red-700 font-medium">Warning: This action cannot be undone!</p>
            <p className="text-red-600 text-sm mt-1">All sales data associated with this product will also be affected.</p>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Type <span className="font-semibold">{product.name}</span> to confirm
            </label>
            <input
              type="text"
              value={deleteConfirm}
              onChange={(e) => setDeleteConfirm(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-red-500 focus:border-transparent outline-none"
              placeholder="Product name"
            />
          </div>

          <div className="flex gap-3">
            <Button variant="outline" className="flex-1" onClick={() => { setIsDeleting(false); setDeleteConfirm('') }}>
              Cancel
            </Button>
            <Button
              variant="danger"
              className="flex-1"
              disabled={deleteConfirm !== product.name}
              onClick={handleDelete}
            >
              Delete Product
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default ProductDetail
