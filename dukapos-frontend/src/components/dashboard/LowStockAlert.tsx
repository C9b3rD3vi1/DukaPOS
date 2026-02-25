import { Link } from 'react-router-dom'
import type { Product } from '@/api/types'
import { Card } from '@/components/common'

interface LowStockAlertProps {
  products: Product[]
  onRestock?: (product: Product) => void
  maxItems?: number
}

export function LowStockAlert({ products, onRestock, maxItems = 5 }: LowStockAlertProps) {
  const lowStockProducts = products
    .filter(p => p.current_stock <= p.low_stock_threshold)
    .sort((a, b) => a.current_stock - b.current_stock)
    .slice(0, maxItems)

  if (lowStockProducts.length === 0) {
    return null
  }

  return (
    <Card padding="none" className="overflow-hidden">
      <div className="p-4 border-b border-gray-100 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center">
            <svg className="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">Low Stock Alert</h3>
            <p className="text-xs text-gray-500">{lowStockProducts.length} items need attention</p>
          </div>
        </div>
        <Link
          to="/products?filter=low-stock"
          className="text-sm text-primary hover:text-primary/80 font-medium"
        >
          View All
        </Link>
      </div>

      <div className="divide-y divide-gray-100">
        {lowStockProducts.map(product => {
          const percentage = Math.round((product.current_stock / product.low_stock_threshold) * 100)
          const isCritical = product.current_stock === 0

          return (
            <div
              key={product.id}
              className="p-4 hover:bg-gray-50 transition"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                    isCritical ? 'bg-red-100' : 'bg-amber-100'
                  }`}>
                    <svg className={`w-5 h-5 ${isCritical ? 'text-red-600' : 'text-amber-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{product.name}</p>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className={`text-xs font-medium ${isCritical ? 'text-red-600' : 'text-amber-600'}`}>
                        {product.current_stock} left
                      </span>
                      <span className="text-xs text-gray-400">• Min: {product.low_stock_threshold}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {/* Progress bar */}
                  <div className="w-16 h-2 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${isCritical ? 'bg-red-500' : 'bg-amber-500'}`}
                      style={{ width: `${Math.min(percentage, 100)}%` }}
                    />
                  </div>

                  {onRestock && (
                    <button
                      onClick={() => onRestock(product)}
                      className="p-2 text-gray-400 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                      title="Restock"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                    </button>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {products.filter(p => p.current_stock <= p.low_stock_threshold).length > maxItems && (
        <div className="p-3 bg-gray-50 text-center">
          <Link
            to="/products?filter=low-stock"
            className="text-sm text-primary hover:text-primary/80 font-medium"
          >
            +{products.filter(p => p.current_stock <= p.low_stock_threshold).length - maxItems} more items
          </Link>
        </div>
      )}
    </Card>
  )
}

export default LowStockAlert
