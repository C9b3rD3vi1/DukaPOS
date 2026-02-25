import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'
import { Button, Modal, Input, Select } from '@/components/common'
import { BarcodeScannerButton } from '@/components/common/BarcodeScanner'
import type { Product, Sale } from '@/api/types'

interface CartItem {
  product: Product
  quantity: number
}

interface SaleFormProps {
  onComplete?: (sale: Sale) => void
}

export function SaleForm({ onComplete }: SaleFormProps) {
  const navigate = useNavigate()
  const { token, shop } = useAuthStore()
  const { queueSale, isOnline } = useSyncStore()
  
  const [products, setProducts] = useState<Product[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [showProductModal, setShowProductModal] = useState(false)
  const [showPaymentModal, setShowPaymentModal] = useState(false)
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'mpesa' | 'card' | 'bank'>('cash')
  const [mpesaPhone, setMpesaPhone] = useState('')
  const [notes, setNotes] = useState('')
  const [processingPayment, setProcessingPayment] = useState(false)

  useEffect(() => {
    fetchProducts()
  }, [token])

  const fetchProducts = async () => {
    if (!token) return
    
    try {
      const response = await fetch('/api/v1/products', {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      if (response.ok) {
        const data = await response.json()
        setProducts(data.data || [])
      }
    } catch (error) {
      console.error('Failed to fetch products:', error)
    }
  }

  const filteredProducts = products.filter(p => {
    const matchesSearch = p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (p.barcode && p.barcode.toLowerCase().includes(searchQuery.toLowerCase()))
    const hasStock = p.current_stock > 0
    return matchesSearch && hasStock
  })

  const addToCart = useCallback((product: Product) => {
    setCart(prev => {
      const existing = prev.find(item => item.product.id === product.id)
      if (existing) {
        if (existing.quantity >= product.current_stock) return prev
        return prev.map(item =>
          item.product.id === product.id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        )
      }
      return [...prev, { product, quantity: 1 }]
    })
  }, [])

  const calculateSubtotal = () => {
    return cart.reduce((sum, item) => sum + (item.product.selling_price * item.quantity), 0)
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const handleBarcodeScan = (barcode: string) => {
    const product = products.find(p => p.barcode === barcode)
    if (product) {
      addToCart(product)
    }
  }

  const handleProcessPayment = async () => {
    if (!token || cart.length === 0) return
    
    setProcessingPayment(true)
    
    const saleData = {
      shop_id: shop?.id,
      items: cart.map(item => ({
        product_id: item.product.id,
        quantity: item.quantity,
        unit_price: item.product.selling_price
      })),
      payment_method: paymentMethod,
      mpesa_phone: paymentMethod === 'mpesa' ? mpesaPhone : undefined,
      notes
    }

    try {
      if (!isOnline) {
        for (const item of cart) {
          await queueSale({
            productId: item.product.id,
            productName: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.selling_price,
            totalAmount: item.product.selling_price * item.quantity,
            paymentMethod,
            mpesaReceipt: undefined,
            staffId: undefined,
            notes,
            createdAt: new Date()
          })
        }
        
        setCart([])
        setShowPaymentModal(false)
        setMpesaPhone('')
        setNotes('')
        return
      }

      const response = await fetch('/api/v1/sales', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(saleData)
      })

      if (!response.ok) {
        throw new Error('Failed to process sale')
      }

      const data = await response.json()
      
      setCart([])
      setShowPaymentModal(false)
      setMpesaPhone('')
      setNotes('')
      
      if (data.data?.id) {
        navigate(`/sales/${data.data.id}`)
      }
      
      onComplete?.(data.data)
    } catch (error) {
      console.error('Sale error:', error)
      alert('Failed to process sale. Please try again.')
    } finally {
      setProcessingPayment(false)
    }
  }

  const cartItemCount = cart.reduce((sum, item) => sum + item.quantity, 0)
  const subtotal = calculateSubtotal()

  return (
    <div className="flex flex-col h-full">
      <div className="sticky top-0 bg-white border-b border-gray-200 p-4 z-10">
        <div className="flex items-center justify-between mb-3">
          <h1 className="text-xl font-bold text-gray-900">New Sale</h1>
          <div className="flex items-center gap-2">
            <BarcodeScannerButton onScan={handleBarcodeScan} />
            <button
              onClick={() => setShowProductModal(true)}
              className="p-2 bg-primary text-white rounded-lg"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>

        <input
          type="text"
          placeholder="Search products..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onFocus={() => setShowProductModal(true)}
          className="w-full px-4 py-3 bg-gray-100 border-0 rounded-xl focus:ring-2 focus:ring-primary"
        />

        {cart.length > 0 && (
          <div className="mt-4 p-4 bg-primary/5 rounded-xl border border-primary/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium text-gray-900">{cartItemCount} items in cart</p>
              </div>
              <div className="text-right">
                <p className="text-xl font-bold text-primary">{formatCurrency(subtotal)}</p>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {filteredProducts.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">No products found</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {filteredProducts.map(product => {
              const inCart = cart.find(c => c.product.id === product.id)
              
              return (
                <button
                  key={product.id}
                  onClick={() => addToCart(product)}
                  disabled={product.current_stock === 0}
                  className={`p-3 rounded-xl border text-left transition ${
                    product.current_stock === 0
                      ? 'bg-gray-50 border-gray-100 opacity-50'
                      : inCart
                        ? 'bg-primary/10 border-primary'
                        : 'bg-white border-gray-200 hover:border-primary hover:shadow-md'
                  }`}
                >
                  <div className="aspect-square bg-gray-100 rounded-lg mb-2 flex items-center justify-center overflow-hidden">
                    {product.image_url ? (
                      <img src={product.image_url} alt={product.name} className="w-full h-full object-cover" />
                    ) : (
                      <svg className="w-8 h-8 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                      </svg>
                    )}
                  </div>
                  <p className="font-medium text-gray-900 truncate text-sm">{product.name}</p>
                  <p className="text-primary font-bold">{formatCurrency(product.selling_price)}</p>
                  {inCart && (
                    <span className="absolute top-2 right-2 w-5 h-5 bg-primary text-white text-xs rounded-full flex items-center justify-center">
                      {inCart.quantity}
                    </span>
                  )}
                </button>
              )
            })}
          </div>
        )}
      </div>

      {cart.length > 0 && (
        <div className="fixed bottom-6 left-4 right-4 md:left-auto md:right-6 md:w-96">
          <button
            onClick={() => setShowPaymentModal(true)}
            className="w-full py-4 bg-primary text-white rounded-xl shadow-lg hover:bg-primary/90 transition flex items-center justify-center gap-3"
          >
            <span className="font-semibold">Checkout</span>
            <span className="bg-white/20 px-3 py-1 rounded-full">
              {cartItemCount}
            </span>
            <span className="font-bold">{formatCurrency(subtotal)}</span>
          </button>
        </div>
      )}

      <Modal isOpen={showProductModal} onClose={() => setShowProductModal(false)} title="Select Product" size="lg">
        <div className="space-y-4">
          <Input
            placeholder="Search products..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          
          <div className="max-h-96 overflow-y-auto space-y-2">
            {filteredProducts.slice(0, 20).map(product => (
              <button
                key={product.id}
                onClick={() => {
                  addToCart(product)
                  setShowProductModal(false)
                }}
                className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 text-left"
              >
                <div className="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center">
                  {product.image_url ? (
                    <img src={product.image_url} alt={product.name} className="w-full h-full object-cover rounded-lg" />
                  ) : (
                    <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  )}
                </div>
                <div className="flex-1">
                  <p className="font-medium text-gray-900">{product.name}</p>
                  <p className="text-sm text-gray-500">{product.current_stock} in stock</p>
                </div>
                <p className="font-bold text-primary">{formatCurrency(product.selling_price)}</p>
              </button>
            ))}
          </div>
        </div>
      </Modal>

      <Modal isOpen={showPaymentModal} onClose={() => setShowPaymentModal(false)} title="Checkout">
        <div className="space-y-4">
          <div className="max-h-48 overflow-y-auto border rounded-lg">
            {cart.map(item => (
              <div key={item.product.id} className="flex items-center justify-between p-3 border-b last:border-0">
                <div>
                  <p className="font-medium">{item.product.name}</p>
                  <p className="text-sm text-gray-500">{item.quantity} x {formatCurrency(item.product.selling_price)}</p>
                </div>
                <p className="font-semibold">{formatCurrency(item.product.selling_price * item.quantity)}</p>
              </div>
            ))}
          </div>

          <div className="flex justify-between text-lg font-bold">
            <span>Total</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>

          <Select
            label="Payment Method"
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value as typeof paymentMethod)}
            options={[
              { value: 'cash', label: 'Cash' },
              { value: 'mpesa', label: 'M-Pesa' },
              { value: 'card', label: 'Card' },
              { value: 'bank', label: 'Bank Transfer' }
            ]}
          />

          {paymentMethod === 'mpesa' && (
            <Input
              label="M-Pesa Phone Number"
              type="tel"
              value={mpesaPhone}
              onChange={(e) => setMpesaPhone(e.target.value)}
              placeholder="2547XXXXXXXX"
            />
          )}

          <Input
            label="Notes (Optional)"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Add any notes..."
          />

          {!isOnline && (
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg">
              <p className="text-amber-700 text-sm">You're offline. Sale will be synced when connection is restored.</p>
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <Button variant="outline" className="flex-1" onClick={() => setShowPaymentModal(false)}>
              Cancel
            </Button>
            <Button
              className="flex-1"
              onClick={handleProcessPayment}
              disabled={processingPayment || cart.length === 0 || (paymentMethod === 'mpesa' && !mpesaPhone)}
            >
              {processingPayment ? 'Processing...' : 'Complete Sale'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default SaleForm
