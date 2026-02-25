import { Link } from 'react-router-dom'
import type { Product } from '@/api/types'

interface ProductCardProps {
  product: Product
  onClick?: () => void
  onAddToCart?: (product: Product) => void
  showStock?: boolean
  compact?: boolean
}

export function ProductCard({ product, onClick, onAddToCart, showStock = true, compact = false }: ProductCardProps) {
  const isLowStock = product.current_stock <= product.low_stock_threshold

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const content = (
    <div className={`bg-white rounded-xl border border-gray-200 hover:border-primary hover:shadow-md transition-all ${compact ? 'p-3' : 'p-4'} ${product.current_stock === 0 ? 'opacity-60' : ''}`}>
      <div className={`${compact ? 'h-16 mb-2' : 'w-full h-20 mb-3'} bg-gray-100 rounded-lg flex items-center justify-center overflow-hidden`}>
        {product.image_url ? (
          <img src={product.image_url} alt={product.name} className="w-full h-full object-cover rounded-lg" />
        ) : (
          <svg className={`${compact ? 'w-6 h-6' : 'w-8 h-8'} text-gray-400`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
        )}
      </div>
      <p className="font-medium text-gray-900 truncate">{product.name}</p>
      <p className={`text-sm ${compact ? '' : 'mt-1'} ${isLowStock ? 'text-red-500' : 'text-gray-500'}`}>
        {formatCurrency(product.selling_price)}
      </p>
      {showStock && (
        <p className={`text-xs mt-1 ${isLowStock ? 'text-red-500 font-medium' : 'text-green-500'}`}>
          {product.current_stock} in stock
        </p>
      )}
    </div>
  )

  if (onClick) {
    return <div onClick={onClick}>{content}</div>
  }

  if (onAddToCart) {
    return (
      <button
        onClick={() => onAddToCart(product)}
        disabled={product.current_stock === 0}
        className="w-full text-left"
      >
        {content}
      </button>
    )
  }

  return (
    <Link to={`/products/${product.id}`}>
      {content}
    </Link>
  )
}

interface ProductListItemProps {
  product: Product
  onEdit?: (product: Product) => void
  onDelete?: (product: Product) => void
}

export function ProductListItem({ product, onEdit, onDelete }: ProductListItemProps) {
  const isLowStock = product.current_stock <= product.low_stock_threshold

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  return (
    <div className="flex items-center gap-4 p-4 hover:bg-gray-50 transition">
      <div className="w-12 h-12 bg-primary-50 rounded-xl flex items-center justify-center overflow-hidden flex-shrink-0">
        {product.image_url ? (
          <img src={product.image_url} alt={product.name} className="w-full h-full object-cover" />
        ) : (
          <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <p className="font-medium text-gray-900 truncate">{product.name}</p>
        <div className="flex items-center gap-3 mt-1">
          {product.category && (
            <span className="text-xs text-gray-500">{product.category}</span>
          )}
          {product.barcode && (
            <span className="text-xs text-gray-400">#{product.barcode}</span>
          )}
        </div>
      </div>
      <div className="text-right flex-shrink-0">
        <p className="font-medium text-gray-900">{formatCurrency(product.selling_price)}</p>
        <span className={`text-xs ${isLowStock ? 'text-red-500 bg-red-50' : 'text-green-500 bg-green-50'} px-2 py-0.5 rounded-full`}>
          {product.current_stock} {product.unit}
        </span>
      </div>
      {(onEdit || onDelete) && (
        <div className="flex items-center gap-1 flex-shrink-0">
          {onEdit && (
            <button
              onClick={() => onEdit(product)}
              className="p-2 text-gray-400 hover:text-primary hover:bg-primary-50 rounded-lg transition"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
            </button>
          )}
          {onDelete && (
            <button
              onClick={() => onDelete(product)}
              className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          )}
        </div>
      )}
    </div>
  )
}
