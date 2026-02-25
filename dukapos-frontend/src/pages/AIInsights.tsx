import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard } from '@/components/common/Card'
import { Skeleton, SkeletonList } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { AIPredictions } from '@/api/types'

interface Trend {
  date: string
  sales: number
  revenue: number
}

interface InventoryValue {
  total_value: number
  total_items: number
}

export default function AIInsights() {
  const shop = useAuthStore((state) => state.shop)
  const [predictions, setPredictions] = useState<AIPredictions[]>([])
  const [trends, setTrends] = useState<Trend[]>([])
  const [inventoryValue, setInventoryValue] = useState<InventoryValue | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'predictions' | 'trends' | 'value'>('predictions')

  useEffect(() => {
    if (shop?.id) {
      fetchAI()
    } else {
      setIsLoading(false)
    }
  }, [shop?.id])

  const fetchAI = async () => {
    if (!shop?.id) return
    setIsLoading(true)
    setError(null)
    try {
      const [predRes, trendRes, valueRes] = await Promise.all([
        api.get('/v1/ai/predictions'),
        api.get('/v1/ai/trends'),
        api.get('/v1/ai/inventory-value')
      ])
      setPredictions((predRes.data as AIPredictions[]) || [])
      setTrends((trendRes.data as Trend[]) || [])
      setInventoryValue((valueRes.data as InventoryValue) || null)
    } catch (err) {
      console.error(err)
      setError('Failed to load AI insights')
    } finally {
      setIsLoading(false)
    }
  }

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES', minimumFractionDigits: 0 }).format(amount)

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <Skeleton className="h-8 w-32 mb-2" />
          <Skeleton className="h-4 w-48" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
        <SkeletonList items={5} />
      </div>
    )
  }

  if (!shop) {
    return (
      <Card className="text-center py-12">
        <EmptyState
          variant="generic"
          title="No Shop Selected"
          description="Please select a shop to view AI insights"
        />
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="text-center py-12">
        <EmptyState
          variant="generic"
          title="Error Loading Data"
          description={error}
          action={{
            label: 'Try Again',
            onClick: () => fetchAI()
          }}
        />
      </Card>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">AI Insights</h1>
        <p className="text-gray-500 mt-1">Smart analytics for your business</p>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <StatCard
          title="Inventory Value"
          value={formatCurrency(inventoryValue?.total_value || 0)}
          variant="info"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          }
        />
        <StatCard
          title="Total Items"
          value={inventoryValue?.total_items || 0}
          variant="default"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          }
        />
        <StatCard
          title="Predicted Restocks"
          value={predictions.length}
          variant="warning"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          }
        />
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6">
        {(['predictions', 'trends', 'value'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-xl font-medium text-sm transition ${
              activeTab === tab ? 'bg-primary text-white' : 'bg-white text-gray-600 border border-gray-200 hover:bg-gray-50'
            }`}
          >
            {tab === 'predictions' ? 'Stock Predictions' : tab === 'trends' ? 'Sales Trends' : 'Inventory Value'}
          </button>
        ))}
      </div>

      {activeTab === 'predictions' && (
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm">
          <div className="p-5 border-b border-gray-100">
            <h2 className="font-semibold text-gray-900">Restock Recommendations</h2>
            <p className="text-sm text-gray-500">AI-powered predictions based on sales velocity</p>
          </div>
          {predictions.length > 0 ? (
            <div className="divide-y divide-gray-100">
              {predictions.map((item) => (
                <div key={item.product_id} className="p-4 flex items-center justify-between hover:bg-gray-50">
                  <div className="flex items-center gap-4">
                    <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                      item.trend === 'up' ? 'bg-green-100' : item.trend === 'down' ? 'bg-red-100' : 'bg-gray-100'
                    }`}>
                      <span className={`text-lg ${item.trend === 'up' ? '📈' : item.trend === 'down' ? '📉' : '➡️'}`}>
                        {item.trend === 'up' ? '↑' : item.trend === 'down' ? '↓' : '→'}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{item.product_name}</p>
                      <p className="text-sm text-gray-500">
                        Stock: {item.current_stock} • Avg daily: {item.avg_daily_sales}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-semibold text-gray-900">{item.days_until_stockout} days</p>
                    <p className="text-sm text-green-600">Order {item.recommended_order} units</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-8 text-center text-gray-500">
              No prediction data available
            </div>
          )}
        </div>
      )}

      {activeTab === 'trends' && (
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5">
          <h2 className="font-semibold text-gray-900 mb-4">Sales Trends</h2>
          {trends.length > 0 ? (
            <div className="h-64 flex items-end gap-2">
              {trends.map((item, i) => {
                const maxRevenue = Math.max(...trends.map(t => t.revenue), 1)
                const heightPercent = item.revenue > 0 ? Math.max((item.revenue / maxRevenue) * 100, 5) : 5
                return (
                  <div key={i} className="flex-1 flex flex-col items-center">
                    <div 
                      className="w-full bg-primary rounded-t-lg transition-all hover:bg-primary-dark"
                      style={{ height: `${heightPercent}%` }}
                      title={formatCurrency(item.revenue)}
                    ></div>
                    <span className="text-xs text-gray-500 mt-2">{item.date.slice(5)}</span>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="text-center text-gray-500 py-8">No trend data available</div>
          )}
        </div>
      )}

      {activeTab === 'value' && (
        <div className="grid md:grid-cols-2 gap-4">
          <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
            <p className="text-sm text-gray-500 mb-1">Total Inventory Value</p>
            <p className="text-3xl font-bold text-gray-900">{formatCurrency(inventoryValue?.total_value || 0)}</p>
          </div>
          <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
            <p className="text-sm text-gray-500 mb-1">Total Products</p>
            <p className="text-3xl font-bold text-gray-900">{inventoryValue?.total_items || 0}</p>
          </div>
        </div>
      )}
    </div>
  )
}
