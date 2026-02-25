import { useState, useEffect, useMemo } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Skeleton } from '@/components/common/Skeleton'
import { Chart } from '@/components/common/Chart'

interface TopProduct {
  name: string
  revenue: number
  quantity: number
  percentage: number
}

interface SalesReport {
  total_sales: number
  total_profit: number
  transaction_count: number
  top_products: TopProduct[]
  daily_sales?: Array<{ date: string; sales: number; profit: number }>
  payment_breakdown?: Array<{ method: string; count: number; total: number }>
}

export default function Reports() {
  const shop = useAuthStore((state) => state.shop)
  const [period, setPeriod] = useState<'daily' | 'weekly' | 'monthly'>('daily')
  const [report, setReport] = useState<SalesReport | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [isExporting, setIsExporting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (shop?.id) {
      fetchReport()
    } else {
      setIsLoading(false)
    }
  }, [shop?.id, period, startDate, endDate])

  const fetchReport = async () => {
    if (!shop?.id) return
    setIsLoading(true)
    setError('')
    try {
      const params = new URLSearchParams()
      params.append('period', period)
      if (startDate) params.append('start_date', startDate)
      if (endDate) params.append('end_date', endDate)
      
      const response = await api.get(`/v1/reports?${params}`)
      const responseData = response.data
      const reportData = responseData?.data || responseData
      setReport(reportData || {
        total_sales: 0,
        total_profit: 0,
        transaction_count: 0,
        top_products: []
      })
    } catch (err) {
      console.error(err)
      setError('Unable to load report data')
      setReport({
        total_sales: 0,
        total_profit: 0,
        transaction_count: 0,
        top_products: []
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleExport = async (format: 'csv' | 'pdf') => {
    if (!shop?.id) return
    setIsExporting(true)
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      params.append('format', format)
      if (startDate) params.append('start_date', startDate)
      if (endDate) params.append('end_date', endDate)
      
      const response = await api.get(`/v1/export/sales?${params}`, {
        responseType: 'blob'
      })
      
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `sales_report_${new Date().toISOString().split('T')[0]}.${format}`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (err) {
      console.error('Export failed:', err)
      alert('Export failed. Please try again.')
    } finally {
      setIsExporting(false)
    }
  }

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES', minimumFractionDigits: 0 }).format(amount)

  // Mock daily sales data for demo (replace with actual API data)
  const mockDailySales = useMemo(() => {
    const days = period === 'daily' ? 7 : period === 'weekly' ? 4 : 12
    return Array.from({ length: days }, (_, i) => {
      const date = new Date()
      date.setDate(date.getDate() - (days - 1 - i))
      return {
        date: date.toISOString().split('T')[0],
        sales: Math.floor(Math.random() * 50000) + 10000,
        profit: Math.floor(Math.random() * 15000) + 3000
      }
    })
  }, [period])

  const chartDataWithMock = useMemo(() => ({
    sales: mockDailySales.map(d => ({
      label: new Date(d.date).toLocaleDateString('en-KE', { month: 'short', day: 'numeric' }),
      value: d.sales
    })),
    topProducts: report?.top_products?.slice(0, 5).map(p => ({
      label: p.name.length > 12 ? p.name.slice(0, 9) + '...' : p.name,
      value: p.revenue
    })) || [],
    payment: [
      { label: 'Cash', value: Math.floor((report?.total_sales || 0) * 0.4) },
      { label: 'M-Pesa', value: Math.floor((report?.total_sales || 0) * 0.5) },
      { label: 'Card', value: Math.floor((report?.total_sales || 0) * 0.1) }
    ].filter(p => p.value > 0)
  }), [report, mockDailySales])

  return (
    <div className="-mx-4 md:-mx-6">
      <div className="px-4 md:px-6 pb-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-surface-900">Reports</h1>
            <p className="text-surface-500 mt-1">Sales analytics and insights</p>
          </div>
          <div className="flex gap-2">
            <Button
              onClick={() => handleExport('csv')}
              disabled={isExporting || !shop}
              variant="secondary"
              leftIcon={
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              }
            >
              CSV
            </Button>
            <Button
              onClick={() => handleExport('pdf')}
              disabled={isExporting || !shop}
              leftIcon={
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              }
            >
              PDF
            </Button>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-xl">
            {error}
          </div>
        )}

        {/* Date Filters */}
        <Card className="mb-6">
          <div className="flex flex-wrap gap-4 items-end">
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">Period</label>
              <div className="flex gap-2">
                {(['daily', 'weekly', 'monthly'] as const).map((p) => (
                  <button 
                    key={p} 
                    onClick={() => setPeriod(p)} 
                    className={`px-4 py-2 rounded-xl font-medium transition ${
                      period === p 
                        ? 'bg-primary text-white shadow-md shadow-primary/25' 
                        : 'bg-surface-100 text-surface-600 hover:bg-surface-200'
                    }`}
                  >
                    {p.charAt(0).toUpperCase() + p.slice(1)}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">From</label>
              <input 
                type="date" 
                value={startDate} 
                onChange={(e) => setStartDate(e.target.value)} 
                className="px-4 py-2.5 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" 
              />
            </div>
            <div>
              <label className="block text-sm font-semibold text-surface-700 mb-2">To</label>
              <input 
                type="date" 
                value={endDate} 
                onChange={(e) => setEndDate(e.target.value)} 
                className="px-4 py-2.5 bg-surface-50 border border-surface-200 rounded-xl focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none" 
              />
            </div>
            {(startDate || endDate) && (
              <button 
                onClick={() => { setStartDate(''); setEndDate('') }} 
                className="px-4 py-2.5 text-red-600 hover:bg-red-50 rounded-xl transition"
              >
                Clear
              </button>
            )}
          </div>
        </Card>

        {!shop ? (
          <Card className="text-center py-12">
            <div className="w-16 h-16 bg-amber-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-surface-900 mb-2">No Shop Selected</h3>
            <p className="text-surface-500">Please select a shop to view reports</p>
          </Card>
        ) : isLoading ? (
          <div className="grid gap-4 md:grid-cols-3 mb-6">
            <Card><Skeleton className="h-16" /></Card>
            <Card><Skeleton className="h-16" /></Card>
            <Card><Skeleton className="h-16" /></Card>
          </div>
        ) : (
          <>
            <div className="grid gap-4 md:grid-cols-3 mb-6">
              <StatCard
                title="Total Sales"
                value={formatCurrency(report?.total_sales || 0)}
                variant="success"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                }
              />
              <StatCard
                title="Total Profit"
                value={formatCurrency(report?.total_profit || 0)}
                variant="info"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                  </svg>
                }
              />
              <StatCard
                title="Transactions"
                value={report?.transaction_count || 0}
                variant="default"
                icon={
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                }
              />
            </div>

            <Card padding="none">
              <div className="p-4 border-b border-surface-100 font-semibold text-surface-900">Top Products</div>
              {report?.top_products && report.top_products.length > 0 ? (
                <div className="divide-y divide-surface-100">
                  {report.top_products.map((p, i) => (
                    <div key={i} className="p-4 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 bg-primary/10 rounded-full flex items-center justify-center text-xs font-medium text-primary">{i + 1}</span>
                        <span className="font-medium text-surface-900">{p.name}</span>
                      </div>
                      <div className="text-right">
                        <p className="font-semibold text-surface-900">{formatCurrency(p.revenue)}</p>
                        <p className="text-sm text-surface-500">{p.quantity} sold ({p.percentage.toFixed(1)}%)</p>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="p-8 text-center text-surface-500">
                  <svg className="w-12 h-12 mx-auto mb-3 text-surface-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                  <p>No data for this period</p>
                </div>
              )}
            </Card>

            {/* Charts Section */}
            <div className="grid gap-4 md:grid-cols-2 mt-6">
              {/* Sales Trend Chart */}
              <Card>
                <h3 className="font-semibold text-surface-900 mb-4">Sales Trend</h3>
                <Chart 
                  data={chartDataWithMock.sales} 
                  type="bar" 
                  height={220}
                  showValues={true}
                  showLabels={true}
                />
              </Card>

              {/* Top Products Chart */}
              <Card>
                <h3 className="font-semibold text-surface-900 mb-4">Top Products</h3>
                <Chart 
                  data={chartDataWithMock.topProducts} 
                  type="bar" 
                  height={220}
                  showValues={true}
                  showLabels={true}
                />
              </Card>

              {/* Payment Methods */}
              <Card>
                <h3 className="font-semibold text-surface-900 mb-4">Payment Methods</h3>
                <Chart 
                  data={chartDataWithMock.payment} 
                  type="pie" 
                  height={220}
                  showValues={true}
                  showLabels={true}
                />
              </Card>

              {/* Profit vs Revenue */}
              <Card>
                <h3 className="font-semibold text-surface-900 mb-4">Profit Overview</h3>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-surface-600">Revenue</span>
                      <span className="font-medium text-surface-900">{formatCurrency(report?.total_sales || 0)}</span>
                    </div>
                    <div className="h-3 bg-surface-100 rounded-full overflow-hidden">
                      <div className="h-full bg-primary rounded-full" style={{ width: '100%' }} />
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-surface-600">Profit</span>
                      <span className="font-medium text-green-600">{formatCurrency(report?.total_profit || 0)}</span>
                    </div>
                    <div className="h-3 bg-surface-100 rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-green-500 rounded-full" 
                        style={{ width: `${report?.total_sales ? ((report.total_profit / report.total_sales) * 100) : 0}%` }} 
                      />
                    </div>
                  </div>
                  <div className="pt-2 text-center">
                    <span className="text-2xl font-bold text-green-600">
                      {report?.total_sales ? ((report.total_profit / report.total_sales) * 100).toFixed(1) : 0}%
                    </span>
                    <p className="text-sm text-surface-500">Profit Margin</p>
                  </div>
                </div>
              </Card>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
