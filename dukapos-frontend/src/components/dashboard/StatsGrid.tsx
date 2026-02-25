import { Link } from 'react-router-dom'
import type { Product } from '@/api/types'

interface StatsGridProps {
  totalSales: number
  profit: number
  transactions: number
  products: number
  lowStockCount?: number
}

export function StatsGrid({ totalSales, profit, transactions, products, lowStockCount = 0 }: StatsGridProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const stats = [
    {
      title: "Today's Sales",
      value: formatCurrency(totalSales),
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
      ),
      variant: 'success' as const
    },
    {
      title: 'Profit',
      value: formatCurrency(profit),
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
      variant: 'success' as const
    },
    {
      title: 'Transactions',
      value: transactions,
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z" />
        </svg>
      ),
      variant: 'default' as const
    },
    {
      title: 'Products',
      value: products,
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
      ),
      variant: (lowStockCount > 0 ? 'warning' : 'default') as 'default' | 'success' | 'warning'
    }
  ]

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {stats.map((stat, index) => (
        <div 
          key={index}
          className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm"
        >
          <div className="flex items-center justify-between mb-3">
            <div className={`w-12 h-12 bg-gradient-to-br ${
              stat.variant === 'success' ? 'from-green-400 to-green-600' :
              stat.variant === 'warning' ? 'from-amber-400 to-amber-600' :
              'from-blue-400 to-blue-600'
            } rounded-xl flex items-center justify-center text-white`}>
              {stat.icon}
            </div>
            {stat.variant === 'warning' && lowStockCount > 0 && (
              <span className="text-xs font-semibold text-amber-600 bg-amber-50 px-2 py-1 rounded-full">
                {lowStockCount} low
              </span>
            )}
          </div>
          <p className="text-sm text-gray-500 font-medium">{stat.title}</p>
          <p className="text-xl md:text-2xl font-bold text-gray-900 mt-1">{stat.value}</p>
        </div>
      ))}
    </div>
  )
}

interface LowStockAlertProps {
  products: Product[]
}

export function LowStockAlert({ products }: LowStockAlertProps) {
  if (products.length === 0) return null

  return (
    <div className="bg-white rounded-2xl border border-amber-200 shadow-sm overflow-hidden">
      <div className="p-5 border-b border-amber-100 flex items-center justify-between bg-amber-50">
        <h2 className="font-semibold text-amber-800 flex items-center gap-2">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          Low Stock Alert
        </h2>
        <Link to="/products?low_stock=true" className="text-amber-600 text-sm font-medium hover:underline">
          View All
        </Link>
      </div>
      <div className="p-4">
        <div className="flex gap-2 overflow-x-auto pb-2">
          {products.slice(0, 5).map((product) => (
            <div key={product.id} className="flex-shrink-0 bg-amber-50 rounded-xl p-3 min-w-[140px]">
              <p className="font-medium text-gray-900 text-sm truncate">{product.name}</p>
              <p className="text-amber-600 font-bold">{product.current_stock} left</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

interface RecentSalesProps {
  sales: Array<{
    id: number
    product?: Product
    product_id: number
    quantity: number
    total_amount: number
    payment_method: string
    created_at: string
  }>
}

export function RecentSales({ sales }: RecentSalesProps) {
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const getPaymentIcon = (method: string) => {
    const iconClass = "w-4 h-4"
    switch (method) {
      case 'mpesa':
        return <span className={iconClass + ' text-yellow-600'}>📱</span>
      case 'card':
        return <span className={iconClass + ' text-blue-600'}>💳</span>
      default:
        return <span className={iconClass + ' text-green-600'}>💵</span>
    }
  }

  return (
    <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden">
      <div className="p-5 border-b border-gray-100 flex items-center justify-between">
        <h2 className="font-semibold text-gray-900 flex items-center gap-2">
          <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Recent Sales
        </h2>
        <Link to="/sales" className="text-primary text-sm font-medium hover:underline">
          See All
        </Link>
      </div>
      
      {sales && sales.length > 0 ? (
        <div className="divide-y divide-gray-100">
          {sales.slice(0, 5).map((sale) => (
            <div key={sale.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-primary-50 rounded-xl flex items-center justify-center">
                  <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                  </svg>
                </div>
                <div>
                  <p className="font-medium text-gray-900">
                    {sale.product?.name || `Product #${sale.product_id}`}
                  </p>
                  <p className="text-sm text-gray-500">
                    {new Date(sale.created_at).toLocaleTimeString('en-KE', { hour: '2-digit', minute: '2-digit' })}
                  </p>
                </div>
              </div>
              <div className="text-right">
                <p className="font-semibold text-gray-900">
                  {formatCurrency(sale.total_amount)}
                </p>
                <div className="flex items-center gap-1 justify-end">
                  {getPaymentIcon(sale.payment_method)}
                  <span className="text-xs text-gray-500">x{sale.quantity}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="p-8 text-center text-gray-500">
          No sales yet. Start selling!
        </div>
      )}
    </div>
  )
}
