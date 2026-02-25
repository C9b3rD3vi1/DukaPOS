import type { Product } from '@/api/types'

interface CartItem {
  product: Product
  quantity: number
}

interface CartSidebarProps {
  cart: CartItem[]
  onUpdateQuantity: (productId: number, quantity: number) => void
  onRemove: (productId: number) => void
  onCheckout: () => void
  cartTotal: number
  cartItems: number
  isOnline: boolean
}

export function CartSidebar({
  cart,
  onUpdateQuantity,
  onRemove,
  onCheckout,
  cartTotal,
  cartItems,
  isOnline
}: CartSidebarProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  return (
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
                      onClick={() => onUpdateQuantity(item.product.id, item.quantity - 1)}
                      className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center"
                    >
                      -
                    </button>
                    <span className="w-8 text-center font-medium">{item.quantity}</span>
                    <button
                      onClick={() => onUpdateQuantity(item.product.id, item.quantity + 1)}
                      className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center"
                    >
                      +
                    </button>
                  </div>
                  <button
                    onClick={() => onRemove(item.product.id)}
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
                onClick={onCheckout}
                className="w-full py-3 bg-primary text-white rounded-xl font-semibold hover:bg-primary-dark transition-all"
              >
                Checkout
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
