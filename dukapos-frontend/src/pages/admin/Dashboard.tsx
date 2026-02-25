import { useState, useEffect } from 'react'
import { adminApi } from '@/api/client'
import { Skeleton } from '@/components/common/Skeleton'

interface DashboardData {
  total_accounts: number
  total_shops: number
  total_revenue: number
  active_accounts: number
  new_accounts_today: number
  new_shops_today: number
}

interface RevenueData {
  daily: Array<{ date: string; revenue: number }>
  monthly: Array<{ month: string; revenue: number }>
  total: number
}

export default function AdminDashboard() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [revenue, setRevenue] = useState<RevenueData | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    try {
      const [dashRes, revRes] = await Promise.all([
        adminApi.get('/admin/dashboard'),
        adminApi.get('/admin/revenue')
      ])
      setData(dashRes.data as unknown as DashboardData)
      setRevenue(revRes.data as unknown as RevenueData)
    } catch (err) { console.error(err) }
    finally { setIsLoading(false) }
  }

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES', minimumFractionDigits: 0 }).format(amount)

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <Skeleton className="h-8 w2" />
         -48 mb- <Skeleton className="h-4 w-64" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="text-gray-500 mt-1">System overview</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm">
          <p className="text-sm text-gray-500">Total Accounts</p>
          <p className="text-2xl font-bold text-gray-900">{data?.total_accounts || 0}</p>
        </div>
        <div className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm">
          <p className="text-sm text-gray-500">Total Shops</p>
          <p className="text-2xl font-bold text-gray-900">{data?.total_shops || 0}</p>
        </div>
        <div className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm">
          <p className="text-sm text-gray-500">Active Accounts</p>
          <p className="text-2xl font-bold text-green-600">{data?.active_accounts || 0}</p>
        </div>
        <div className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm">
          <p className="text-sm text-gray-500">New Today</p>
          <p className="text-2xl font-bold text-primary">{data?.new_accounts_today || 0}</p>
        </div>
      </div>

      {/* Revenue Chart */}
      <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
        <h2 className="font-semibold text-gray-900 mb-4">Monthly Revenue</h2>
        {revenue?.monthly && revenue.monthly.length > 0 ? (
          <div className="h-48 flex items-end gap-2">
            {revenue.monthly.map((item, i) => (
              <div key={i} className="flex-1 flex flex-col items-center">
                <div 
                  className="w-full bg-primary rounded-t-lg"
                  style={{ height: `${(item.revenue / (Math.max(...revenue.monthly.map(m => m.revenue)) || 1)) * 100}%`, minHeight: '8px' }}
                  title={formatCurrency(item.revenue)}
                ></div>
                <span className="text-xs text-gray-500 mt-2">{item.month}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center text-gray-500 py-8">No revenue data</div>
        )}
        <div className="mt-4 pt-4 border-t border-gray-100">
          <p className="text-lg font-semibold">Total Revenue: {formatCurrency(revenue?.total || 0)}</p>
        </div>
      </div>
    </div>
  )
}
