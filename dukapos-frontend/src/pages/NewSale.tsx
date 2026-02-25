import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'
import { useQRPayment } from '@/hooks/useQRPayment'
import { dbSales } from '@/db/db'
import { syncEngine } from '@/db/sync'
import { BarcodeScannerButton, BarcodeScannerModal } from '@/components/common/BarcodeScanner'
import type { Product, Sale } from '@/api/types'

interface CartItem {
  product: Product
  quantity: number
}

export default function NewSale() {
  const shop = useAuthStore((state) => state.shop)
  const { isOnline, syncNow } = useSyncStore()
  
  const {
    startScanning: startQRScan,
    isScanning: isQRScaning,
    isCapacitor: isMobile
  } = useQRPayment()
  
  const [products, setProducts] = useState<Product[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isProcessing, setIsProcessing] = useState(false)
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'mpesa' | 'card' | 'bank'>('cash')
  const [mpesaPhone, setMpesaPhone] = useState('')
  const [showCheckout, setShowCheckout] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)
  const [lastSaleTotal, setLastSaleTotal] = useState(0)
  const [error, setError] = useState('')
  const [sales, setSales] = useState<Sale[]>([])
  const [activeTab, setActiveTab] = useState<'new' | 'history'>('new')
  const [showBarcode, setShowBarcode] = useState(false)
  const [pendingCount, setPendingCount] = useState(0)

  const handleBarcodeScan = (barcode: string) => {
    const product = products.find(p => p.barcode === barcode)
    if (product) {
      addToCart(product)
    } else {
      setError(`Product not found for barcode: ${barcode}`)
    }
    setShowBarcode(false)
  }

  const handleQRScan = (data: { type: string; amount?: number; phone?: string }) => {
    if (data.type === 'mpesa' || data.type === 'bank') {
      setPaymentMethod('mpesa')
      if (data.phone) setMpesaPhone(data.phone)
    }
    setShowCheckout(true)
  }

  const startQRScanPayment = async () => {
    await startQRScan(handleQRScan)
  }

  useEffect(() => {
    fetchProducts()
    fetchSales()
    loadPendingSales()
  }, [shop?.id])

  const loadPendingSales = async () => {
    try {
      const unsynced = await dbSales.getUnsynced()
      setPendingCount(unsynced.length)
    } catch (e) {
      console.error(e)
    }
  }

  const fetchProducts = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get<{ data: Product[] }>('/v1/products?limit=100')
      const data = response.data?.data || []
      setProducts(data)
    } catch (err) {
      // Try to load from offline cache
      const cached = localStorage.getItem('products_cache')
      if (cached) {
        setProducts(JSON.parse(cached))
      }
      setError('Failed to load products')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchSales = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get<{ data: Sale[] }>('/v1/sales?limit=20')
      const data = response.data?.data || []
      setSales(data)
    } catch (err) {
      console.error(err)
    }
  }

  const addToCart = (product: Product) => {
    const existing = cart.find(item => item.product.id === product.id)
    if (existing) {
      if (existing.quantity < product.current_stock) {
        setCart(cart.map(item => 
          item.product.id === product.id 
            ? { ...item, quantity: item.quantity + 1 }
            : item
        ))
      }
    } else {
      if (product.current_stock > 0) {
        setCart([...cart, { product, quantity: 1 }])
      }
    }
  }

  const removeFromCart = (productId: number) => {
    setCart(cart.filter(item => item.product.id !== productId))
  }

  const updateQuantity = (productId: number, quantity: number) => {
    if (quantity <= 0) {
      removeFromCart(productId)
      return
    }
    const product = products.find(p => p.id === productId)
    if (product && quantity <= product.current_stock) {
      setCart(cart.map(item => 
        item.product.id === productId 
          ? { ...item, quantity }
          : item
      ))
    }
  }

  const cartTotal = cart.reduce((sum, item) => sum + (item.product.selling_price * item.quantity), 0)
  const cartItems = cart.reduce((sum, item) => sum + item.quantity, 0)

  const handleCheckout = async () => {
    if (cart.length === 0) return
    setError('')
    setIsProcessing(true)

    try {
      for (const item of cart) {
        const saleData = {
          product_id: item.product.id,
          quantity: item.quantity,
          unit_price: item.product.selling_price,
          payment_method: paymentMethod,
          mpesa_phone: paymentMethod === 'mpesa' ? mpesaPhone : undefined
        }

        if (isOnline) {
          await api.post('/v1/sales', saleData)
        } else {
          // Queue for offline sync using proper Dexie.js sync engine
          await dbSales.add({
            productId: item.product.id,
            productName: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.selling_price,
            totalAmount: item.product.selling_price * item.quantity,
            paymentMethod: paymentMethod,
            createdAt: new Date(),
            synced: false
          })
          
          await syncEngine.queueForSync('sale', 'create', {
            shop_id: shop?.id,
            product_id: item.product.id,
            quantity: item.quantity,
            unit_price: item.product.selling_price,
            payment_method: paymentMethod,
            mpesa_phone: paymentMethod === 'mpesa' ? mpesaPhone : undefined
          })
        }
      }

      setCart([])
      setShowCheckout(false)
      setMpesaPhone('')
      setLastSaleTotal(cartTotal)
      setShowSuccess(true)
      fetchProducts()
      fetchSales()
      loadPendingSales()
      
      // Hide success after 3 seconds
      setTimeout(() => setShowSuccess(false), 3000)
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      setError(error.response?.data?.error || 'Failed to process sale')
    } finally {
      setIsProcessing(false)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const filteredProducts = products.filter(p => 
    p.name.toLowerCase().includes(search.toLowerCase()) &&
    p.current_stock > 0
  )

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    )
  }

  return (
    <div className="flex flex-col lg:flex-row gap-6">
      <BarcodeScannerModal
        isOpen={showBarcode}
        onClose={() => setShowBarcode(false)}
        onScan={handleBarcodeScan}
      />
      
      {/* Offline Banner */}
      {!isOnline && (
        <div className="bg-amber-500 text-white px-4 py-3 rounded-xl flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
            </svg>
            <span className="font-medium">Offline Mode</span>
          </div>
          <span className="text-sm">Sales will sync when online</span>
        </div>
      )}

      {/* Pending sync indicator */}
      {pendingCount > 0 && (
        <div className="bg-blue-500 text-white px-4 py-3 rounded-xl flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span className="font-medium">{pendingCount} pending sale(s)</span>
          </div>
          {isOnline && (
            <button 
              onClick={async () => {
                await syncNow()
                loadPendingSales()
              }}
              className="px-3 py-1 bg-white text-blue-600 rounded-lg text-sm font-medium"
            >
              Sync Now
            </button>
          )}
        </div>
      )}

      {/* Products Grid */}
      <div className="flex-1">
        <div className="mb-4">
          <h1 className="text-2xl font-bold text-gray-900">New Sale</h1>
          <p className="text-gray-500">{isOnline ? 'Online' : 'Offline - Sales will sync when online'}</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-4">
          <button
            onClick={() => setActiveTab('new')}
            className={`px-4 py-2 rounded-xl font-medium transition ${
              activeTab === 'new' 
                ? 'bg-primary text-white' 
                : 'bg-white text-gray-600 hover:bg-gray-50'
            }`}
          >
            New Sale
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`px-4 py-2 rounded-xl font-medium transition ${
              activeTab === 'history' 
                ? 'bg-primary text-white' 
                : 'bg-white text-gray-600 hover:bg-gray-50'
            }`}
          >
            History
          </button>
        </div>

        {activeTab === 'new' ? (
          <>
            {/* Search & Barcode */}
            <div className="flex gap-2 mb-4">
              <div className="relative flex-1">
                <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  type="text"
                  placeholder="Search products..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="w-full pl-10 pr-4 py-3 bg-white border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                />
              </div>
              <BarcodeScannerButton onScan={handleBarcodeScan} />
            </div>

            {/* Products Grid */}
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-3">
              {filteredProducts.map((product) => (
                <button
                  key={product.id}
                  onClick={() => addToCart(product)}
                  disabled={product.current_stock === 0}
                  className="bg-white p-4 rounded-xl border border-gray-200 hover:border-primary hover:shadow-md transition-all text-left disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <div className="w-full h-20 bg-gray-100 rounded-lg mb-3 flex items-center justify-center">
                    {product.image_url ? (
                      <img src={product.image_url} alt={product.name} className="w-full h-full object-cover rounded-lg" />
                    ) : (
                      <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                      </svg>
                    )}
                  </div>
                  <p className="font-medium text-gray-900 truncate">{product.name}</p>
                  <p className="text-sm text-gray-500">{formatCurrency(product.selling_price)}</p>
                  <p className={`text-xs mt-1 ${product.current_stock <= product.low_stock_threshold ? 'text-red-500' : 'text-green-500'}`}>
                    {product.current_stock} in stock
                  </p>
                </button>
              ))}
            </div>

            {filteredProducts.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No products found
              </div>
            )}
          </>
        ) : (
          /* Sales History */
          <div className="bg-white rounded-xl border border-gray-200">
            {sales.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                No sales yet
              </div>
            ) : (
              <div className="divide-y divide-gray-100">
                {sales.map((sale) => (
                  <div key={sale.id} className="p-4 flex items-center justify-between">
                    <div>
                      <p className="font-medium text-gray-900">{sale.product?.name || `Product #${sale.product_id}`}</p>
                      <p className="text-sm text-gray-500">
                        {new Date(sale.created_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold text-gray-900">{formatCurrency(sale.total_amount)}</p>
                      <p className="text-sm text-gray-500">x{sale.quantity} • {sale.payment_method}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Cart Sidebar */}
      <div className="w-full lg:w-80">
        <div className="bg-white rounded-xl border border-gray-200 sticky top-20">
          <div className="p-4 border-b border-gray-100">
            <div className="flex items-center justify-between">
              <h2 className="font-semibold text-gray-900">Cart</h2>
              <div className="flex items-center gap-2">
                {!isOnline && (
                  <span className="px-2 py-1 bg-amber-100 text-amber-700 text-xs rounded-lg">Offline</span>
                )}
                <span className="px-2 py-1 bg-primary text-white text-xs rounded-lg">
                  {cartItems} items
                </span>
              </div>
            </div>
          </div>

          {cart.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              <svg className="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <p>Cart is empty</p>
              <p className="text-sm">Tap products to add</p>
            </div>
          ) : (
            <>
              <div className="p-4 space-y-3 max-h-64 overflow-y-auto">
                {cart.map((item) => (
                  <div key={item.product.id} className="flex items-center gap-3">
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-900 truncate">{item.product.name}</p>
                      <p className="text-sm text-gray-500">{formatCurrency(item.product.selling_price)} each</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => updateQuantity(item.product.id, item.quantity - 1)}
                        className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center"
                      >
                        -
                      </button>
                      <span className="w-8 text-center font-medium">{item.quantity}</span>
                      <button
                        onClick={() => updateQuantity(item.product.id, item.quantity + 1)}
                        className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center"
                      >
                        +
                      </button>
                    </div>
                    <button
                      onClick={() => removeFromCart(item.product.id)}
                      className="p-1 text-red-500 hover:bg-red-50 rounded"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                ))}
              </div>

              <div className="p-4 border-t border-gray-100">
                <div className="flex items-center justify-between mb-4">
                  <span className="font-semibold text-gray-900">Total</span>
                  <span className="text-xl font-bold text-primary">{formatCurrency(cartTotal)}</span>
                </div>

                <button
                  onClick={() => setShowCheckout(true)}
                  className="w-full py-3 bg-primary text-white rounded-xl font-semibold hover:bg-primary-dark transition-all"
                >
                  Checkout
                </button>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Checkout Modal */}
      {showCheckout && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-lg font-bold text-gray-900">Checkout</h3>
            </div>
            <div className="p-6 space-y-4">
              {error && (
                <div className="p-3 bg-red-50 text-red-600 rounded-xl text-sm">
                  {error}
                </div>
              )}

              <div className="text-center py-4">
                <p className="text-gray-500">Total Amount</p>
                <p className="text-3xl font-bold text-gray-900">{formatCurrency(cartTotal)}</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Payment Method</label>
                <div className="grid grid-cols-4 gap-2">
                  {[
                    { value: 'cash', label: 'Cash', icon: '💵' },
                    { value: 'mpesa', label: 'M-Pesa', icon: '📱' },
                    { value: 'card', label: 'Card', icon: '💳' },
                    { value: 'bank', label: 'Bank', icon: '🏦' }
                  ].map((method) => (
                    <button
                      key={method.value}
                      onClick={() => setPaymentMethod(method.value as 'cash' | 'mpesa' | 'card' | 'bank')}
                      className={`p-3 rounded-xl border-2 transition ${
                        paymentMethod === method.value
                          ? 'border-primary bg-primary-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <span className="text-2xl block mb-1">{method.icon}</span>
                      <span className="text-sm font-medium">{method.label}</span>
                    </button>
                  ))}
                </div>
              </div>

              {paymentMethod === 'bank' && (
                <div className="bg-surface-50 rounded-xl p-4">
                  <div className="text-center">
                    <p className="text-sm text-surface-600 mb-3">Scan customer QR to receive payment</p>
                    {isMobile ? (
                      <button
                        onClick={startQRScanPayment}
                        disabled={isQRScaning}
                        className="w-full py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition flex items-center justify-center gap-2 disabled:opacity-50"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
                        </svg>
                        {isQRScaning ? 'Scanning...' : 'Scan QR Code'}
                      </button>
                    ) : (
                      <div className="text-sm text-surface-500">
                        QR scanning available on mobile app
                      </div>
                    )}
                  </div>
                </div>
              )}

              {paymentMethod === 'mpesa' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
                  <input
                    type="tel"
                    value={mpesaPhone}
                    onChange={(e) => setMpesaPhone(e.target.value)}
                    placeholder="254712345678"
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                  />
                </div>
              )}

              <div className="flex gap-3 pt-4">
                <button
                  onClick={() => setShowCheckout(false)}
                  className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50 transition-all"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCheckout}
                  disabled={isProcessing || (paymentMethod === 'mpesa' && !mpesaPhone)}
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isProcessing ? (
                    <>
                      <span className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                      Processing...
                    </>
                  ) : (
                    `Pay ${formatCurrency(cartTotal)}`
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Success Modal */}
      {showSuccess && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-sm rounded-2xl shadow-2xl p-8 text-center animate-fade-in-up">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">Sale Complete!</h3>
            <p className="text-gray-500 mb-4">Successfully processed</p>
            <p className="text-2xl font-bold text-primary">{formatCurrency(lastSaleTotal)}</p>
          </div>
        </div>
      )}
    </div>
  )
}
