import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, CardHeader, StatCard, StatGrid } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import type { DashboardData, Sale } from '@/api/types'

interface WeeklyDataPoint {
  day: string
  sales: number
}

function getWeekDateRange(): { start: string; end: string } {
  const now = new Date()
  const dayOfWeek = now.getDay()
  const startOfWeek = new Date(now)
  startOfWeek.setDate(now.getDate() - dayOfWeek)
  const endOfWeek = new Date(startOfWeek)
  endOfWeek.setDate(startOfWeek.getDate() + 6)
  
  return {
    start: startOfWeek.toISOString().split('T')[0],
    end: endOfWeek.toISOString().split('T')[0]
  }
}

function aggregateSalesByDay(sales: Sale[]): WeeklyDataPoint[] {
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const dailyTotals: Record<string, number> = {}
  
  days.forEach(day => {
    dailyTotals[day] = 0
  })
  
  sales.forEach(sale => {
    const date = new Date(sale.created_at)
    const day = days[date.getDay()]
    dailyTotals[day] += sale.total_amount
  })

  return days.map(day => ({
    day,
    sales: dailyTotals[day]
  }))
}

export default function Dashboard() {
  const shop = useAuthStore((state) => state.shop)
  const user = useAuthStore((state) => state.user)
  const [data, setData] = useState<DashboardData | null>(null)
  const [weeklyData, setWeeklyData] = useState<WeeklyDataPoint[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  const maxSales = weeklyData.length > 0 ? Math.max(...weeklyData.map(d => d.sales), 1) : 1

  useEffect(() => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    
    const fetchDashboard = async () => {
      try {
        const [dashboardResponse] = await Promise.all([
          api.get(`/v1/shop/dashboard`)
        ])
        setData(dashboardResponse.data as unknown as DashboardData)
      } catch (err) {
        setError('Failed to load dashboard data')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }

    fetchDashboard()
  }, [shop?.id])

  useEffect(() => {
    if (!shop?.id) return
    
    const fetchWeeklyData = async () => {
      try {
        const { start, end } = getWeekDateRange()
        const response = await api.get<{ data: Sale[] }>(
          `/v1/sales?start_date=${start}&end_date=${end}&limit=1000`
        )
        const salesData = response.data?.data || response.data || []
        const sales = Array.isArray(salesData) ? salesData : []
        const aggregated = aggregateSalesByDay(sales)
        setWeeklyData(aggregated)
      } catch (err) {
        console.error('Failed to fetch weekly data:', err)
        setWeeklyData([
          { day: 'Mon', sales: 0 },
          { day: 'Tue', sales: 0 },
          { day: 'Wed', sales: 0 },
          { day: 'Thu', sales: 0 },
          { day: 'Fri', sales: 0 },
          { day: 'Sat', sales: 0 },
          { day: 'Sun', sales: 0 },
        ])
      }
    }

    fetchWeeklyData()
  }, [shop?.id])

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

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    )
  }

  if (!shop) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">No Shop Found</h3>
          <p className="text-gray-500">Please contact support</p>
        </div>
      </div>
    )
  }

  const today = new Date().toLocaleDateString('en-KE', { 
    weekday: 'long', 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  })

  return (
    <div className="space-y-6 -mx-4 px-4 md:-mx-6 md:px-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-surface-900">
            Good {new Date().getHours() < 12 ? 'morning' : new Date().getHours() < 17 ? 'afternoon' : 'evening'}{user?.name ? `, ${user.name}` : ''}!
          </h1>
          <p className="text-surface-500 mt-1">{today}</p>
        </div>
        
        {/* Quick Actions */}
        <div className="flex gap-3">
          <Link to="/sales/new">
            <Button variant="primary" size="lg" leftIcon={
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            }>
              New Sale
            </Button>
          </Link>
          <Link to="/products/new">
            <Button variant="secondary" size="lg" leftIcon={
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
              </svg>
            }>
              Add Product
            </Button>
          </Link>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-red-50 text-red-600 rounded-xl">
          {error}
        </div>
      )}

      {/* Today's Hero Card */}
      <Card variant="elevated" className="relative overflow-hidden bg-gradient-to-br from-primary via-primary to-primary-dark text-white">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4wNSI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
        <div className="relative">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
            <div>
              <p className="text-white/80 font-medium mb-2">Today's Revenue</p>
              <p className="text-4xl md:text-5xl font-bold">
                {formatCurrency(data?.total_sales || 0)}
              </p>
              <div className="flex items-center gap-4 mt-4">
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2z" />
                    </svg>
                  </div>
                  <span className="text-white/90">{data?.transaction_count || 0} transactions</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
                    </svg>
                  </div>
                  <span className="text-white/90">Profit: {formatCurrency(data?.total_profit || 0)}</span>
                </div>
              </div>
            </div>
            <div className="hidden lg:block">
              <div className="w-32 h-32 bg-white/10 rounded-full flex items-center justify-center animate-pulse">
                <svg className="w-16 h-16 text-white/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Stats Grid */}
      <StatGrid columns={4}>
        <StatCard
          title="Products"
          value={data?.product_count || 0}
          variant="info"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          }
          subtitle={`${data?.low_stock_count || 0} low stock`}
          onClick={() => {}}
        />
        <StatCard
          title="Avg. Sale"
          value={formatCurrency((data?.transaction_count || 0) > 0 ? ((data?.total_sales || 0) / (data?.transaction_count || 1)) : 0)}
          variant="success"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
          }
        />
        <StatCard
          title="This Week"
          value={formatCurrency(weeklyData.reduce((acc, d) => acc + d.sales, 0))}
          variant="warning"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          }
        />
      </StatGrid>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Weekly Sales Chart */}
        <Card className="lg:col-span-2">
          <CardHeader
            title="Weekly Sales"
            subtitle="Your sales performance this week"
            icon={
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            }
          />
          <div className="flex items-end justify-between gap-2 h-48 mt-4">
            {weeklyData.map((item) => (
              <div key={item.day} className="flex-1 flex flex-col items-center group">
                <div 
                  className="w-full bg-gradient-to-t from-primary to-primary-light rounded-t-lg transition-all duration-500 group-hover:from-primary-dark group-hover:to-primary relative overflow-hidden"
                  style={{ height: `${maxSales > 0 ? (item.sales / maxSales) * 100 : 0}%`, minHeight: item.sales > 0 ? '8px' : '2px' }}
                >
                  <div className="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300" />
                </div>
                <span className="text-xs text-surface-500 mt-3 font-medium">{item.day}</span>
              </div>
            ))}
          </div>
          {weeklyData.every(d => d.sales === 0) && (
            <p className="text-center text-surface-400 text-sm mt-4">No sales data for this week</p>
          )}
        </Card>

        {/* Low Stock Alert */}
        <Card variant={data?.low_stock && data.low_stock.length > 0 ? 'bordered' : 'default'} className={data?.low_stock && data.low_stock.length > 0 ? 'border-amber-300' : ''}>
          <CardHeader
            title="Low Stock"
            subtitle={`${data?.low_stock_count || 0} products need attention`}
            icon={
              <svg className="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            }
            action={
              <Link to="/products?low_stock=true">
                <Button variant="ghost" size="sm">View All</Button>
              </Link>
            }
          />
          <div className="space-y-3 mt-4">
            {data?.low_stock && data.low_stock.length > 0 ? (
              data.low_stock.slice(0, 4).map((product) => (
                <div key={product.id} className="flex items-center gap-3 p-3 bg-amber-50 rounded-xl hover:bg-amber-100 transition-colors cursor-pointer">
                  <div className="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center">
                    <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-surface-900 truncate">{product.name}</p>
                    <p className="text-sm text-amber-600 font-semibold">{product.current_stock} left</p>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-surface-400">
                <svg className="w-12 h-12 mx-auto mb-2 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                <p>All products well stocked!</p>
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Recent Sales & Top Products */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Sales */}
        <Card>
          <CardHeader
            title="Recent Sales"
            subtitle="Latest transactions"
            icon={
              <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
            action={
              <Link to="/sales">
                <Button variant="ghost" size="sm">See All</Button>
              </Link>
            }
          />
          
          {data?.recent_sales && data.recent_sales.length > 0 ? (
            <div className="space-y-1 mt-2">
              {data.recent_sales.slice(0, 5).map((sale) => (
                <div key={sale.id} className="flex items-center justify-between p-3 hover:bg-surface-50 rounded-xl transition-colors">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-primary-50 rounded-xl flex items-center justify-center">
                      <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                      </svg>
                    </div>
                    <div>
                      <p className="font-medium text-surface-900">
                        {sale.product?.name || `Product #${sale.product_id}`}
                      </p>
                      <p className="text-sm text-surface-500">
                        {new Date(sale.created_at).toLocaleTimeString('en-KE', { hour: '2-digit', minute: '2-digit' })}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-semibold text-surface-900">
                      {formatCurrency(sale.total_amount)}
                    </p>
                    <div className="flex items-center gap-1 justify-end">
                      {getPaymentIcon(sale.payment_method)}
                      <span className="text-xs text-surface-500">x{sale.quantity}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-8 text-center text-surface-400">
              <svg className="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
              <p className="font-medium">No sales yet</p>
              <Link to="/sales/new">
                <Button variant="primary" size="sm" className="mt-4">
                  Start Selling
                </Button>
              </Link>
            </div>
          )}
        </Card>

        {/* Top Products */}
        <Card>
          <CardHeader
            title="Top Products"
            subtitle="Best sellers this week"
            icon={
              <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z" />
              </svg>
            }
            action={
              <Link to="/products">
                <Button variant="ghost" size="sm">View All</Button>
              </Link>
            }
          />
          <div className="space-y-3 mt-2">
            {data?.top_products && data.top_products.length > 0 ? (
              data.top_products.slice(0, 5).map((product, idx) => (
                <div key={product.id} className="flex items-center gap-3 p-3 hover:bg-surface-50 rounded-xl transition-colors">
                  <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                    idx === 0 ? 'bg-amber-100 text-amber-700' :
                    idx === 1 ? 'bg-surface-100 text-surface-700' :
                    idx === 2 ? 'bg-orange-100 text-orange-700' :
                    'bg-surface-50 text-surface-500'
                  }`}>
                    {idx + 1}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-surface-900 truncate">{product.name}</p>
                    <p className="text-sm text-surface-500">{product.current_stock} in stock</p>
                  </div>
                  <p className="font-semibold text-surface-900">
                    {formatCurrency(product.selling_price)}
                  </p>
                </div>
              ))
            ) : (
              <div className="p-8 text-center text-surface-400">
                <svg className="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                </svg>
                <p className="font-medium">No products yet</p>
                <Link to="/products/new">
                  <Button variant="primary" size="sm" className="mt-4">
                    Add Products
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  )
}
