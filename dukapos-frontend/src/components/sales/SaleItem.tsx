import { Link } from 'react-router-dom'
import type { Sale } from '@/api/types'

interface SaleItemProps {
  sale: Sale
  showProduct?: boolean
}

export function SaleItem({ sale, showProduct = true }: SaleItemProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const getPaymentIcon = (method: string) => {
    switch (method) {
      case 'mpesa':
        return (
          <svg className="w-4 h-4 text-yellow-600" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
          </svg>
        )
      case 'card':
        return (
          <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
          </svg>
        )
      default:
        return (
          <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
        )
    }
  }

  return (
    <div className="flex items-center justify-between p-4 hover:bg-gray-50 transition">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 bg-primary-50 rounded-xl flex items-center justify-center">
          <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
          </svg>
        </div>
        <div>
          {showProduct && (
            <Link to={`/products/${sale.product_id}`} className="font-medium text-gray-900 hover:text-primary">
              {sale.product?.name || `Product #${sale.product_id}`}
            </Link>
          )}
          <p className="text-sm text-gray-500">
            {new Date(sale.created_at).toLocaleString('en-KE', {
              day: 'numeric',
              month: 'short',
              hour: '2-digit',
              minute: '2-digit'
            })}
          </p>
        </div>
      </div>
      <div className="text-right">
        <p className="font-semibold text-gray-900">{formatCurrency(sale.total_amount)}</p>
        <div className="flex items-center gap-1 justify-end">
          {getPaymentIcon(sale.payment_method)}
          <span className="text-xs text-gray-500">x{sale.quantity}</span>
        </div>
      </div>
    </div>
  )
}

interface CartItemProps {
  productName: string
  unitPrice: number
  quantity: number
  onQuantityChange: (quantity: number) => void
  onRemove: () => void
}

export function CartItem({ productName, unitPrice, quantity, onQuantityChange, onRemove }: CartItemProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  return (
    <div className="flex items-center gap-3 p-3 border-b border-gray-100 last:border-0">
      <div className="flex-1 min-w-0">
        <p className="font-medium text-gray-900 truncate">{productName}</p>
        <p className="text-sm text-gray-500">{formatCurrency(unitPrice)} each</p>
      </div>
      <div className="flex items-center gap-2">
        <button
          onClick={() => onQuantityChange(quantity - 1)}
          className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition"
        >
          -
        </button>
        <span className="w-8 text-center font-medium">{quantity}</span>
        <button
          onClick={() => onQuantityChange(quantity + 1)}
          className="w-8 h-8 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition"
        >
          +
        </button>
      </div>
      <p className="font-semibold text-gray-900 w-20 text-right">
        {formatCurrency(unitPrice * quantity)}
      </p>
      <button
        onClick={onRemove}
        className="p-1 text-red-500 hover:bg-red-50 rounded transition"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  )
}

interface EmptyCartProps {
  onBrowse?: () => void
}

export function EmptyCart({ onBrowse }: EmptyCartProps) {
  return (
    <div className="p-8 text-center text-gray-500">
      <svg className="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
      </svg>
      <p>Cart is empty</p>
      <p className="text-sm">Tap products to add</p>
      {onBrowse && (
        <button
          onClick={onBrowse}
          className="mt-4 text-primary hover:underline"
        >
          Browse Products
        </button>
      )}
    </div>
  )
}
