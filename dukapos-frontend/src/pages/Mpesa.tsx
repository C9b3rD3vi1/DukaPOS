import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { StatCard } from '@/components/common/Card'
import { Skeleton } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { MpesaPayment } from '@/api/types'

export default function Mpesa() {
  const shop = useAuthStore((state) => state.shop)
  const user = useAuthStore((state) => state.user)
  const [payments, setPayments] = useState<MpesaPayment[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [phone, setPhone] = useState('')
  const [amount, setAmount] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)

  useEffect(() => { fetchPayments() }, [shop?.id])

  const fetchPayments = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get<{ data: MpesaPayment[] }>('/v1/mpesa/payments')
      const data = response.data?.data || []
      setPayments(data)
    } catch (err) { console.error(err) }
    finally { setIsLoading(false) }
  }

  const handleSTKPush = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsProcessing(true)
    try {
      await api.post('/v1/mpesa/stk-push', { phone, amount: parseFloat(amount) })
      alert('STK Push sent! Check your phone.')
      setPhone('')
      setAmount('')
      fetchPayments()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } } }
      alert(e.response?.data?.error || 'Failed to send STK push')
    } finally {
      setIsProcessing(false)
    }
  }

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES', minimumFractionDigits: 0 }).format(amount)

  const totalPayments = payments.length
  const totalReceived = payments.filter(p => p.status === 'completed').reduce((sum, p) => sum + p.amount, 0)
  const pending = payments.filter(p => p.status === 'pending').length

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">M-Pesa</h1>
        <p className="text-gray-500 mt-1">Manage M-Pesa payments</p>
      </div>

      {/* Stats */}
      {!isLoading && shop && user?.plan !== 'free' && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <StatCard
            title="Total Payments"
            value={totalPayments}
            variant="default"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
          <StatCard
            title="Total Received"
            value={formatCurrency(totalReceived)}
            variant="success"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
          <StatCard
            title="Pending"
            value={pending}
            variant="warning"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
        </div>
      )}

      {user?.plan === 'free' ? (
        <div className="bg-gradient-to-r from-yellow-400 to-yellow-500 rounded-xl p-6 text-gray-900 mb-6">
          <h3 className="font-bold text-lg mb-2">Upgrade to Pro</h3>
          <p>Get M-Pesa integration with STK Push, C2B, and B2C</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
          <h3 className="font-semibold mb-4">Request Payment (STK Push)</h3>
          <form onSubmit={handleSTKPush} className="flex gap-3">
            <input type="tel" placeholder="Phone (2547...)" value={phone} onChange={(e) => setPhone(e.target.value)} className="flex-1 px-4 py-3 border rounded-xl" required />
            <input type="number" placeholder="Amount" value={amount} onChange={(e) => setAmount(e.target.value)} className="w-32 px-4 py-3 border rounded-xl" required />
            <button type="submit" disabled={isProcessing} className="px-6 py-3 bg-yellow-500 text-gray-900 font-semibold rounded-xl hover:bg-yellow-400 disabled:opacity-50">
              {isProcessing ? 'Sending...' : 'Send'}
            </button>
          </form>
        </div>
      )}

      <div className="bg-white rounded-xl border border-gray-200">
        <div className="p-4 border-b font-semibold">Payment History</div>
        {isLoading ? (
          <div className="p-8">
            <Skeleton className="h-16 mb-2" />
            <Skeleton className="h-16 mb-2" />
            <Skeleton className="h-16" />
          </div>
        ) : !shop ? (
          <div className="p-8">
            <EmptyState
              variant="generic"
              title="No Shop Selected"
              description="Please select a shop to view M-Pesa payments"
            />
          </div>
        ) : payments.length === 0 ? (
          <div className="p-8">
            <EmptyState
              variant="generic"
              title="No Payments Yet"
              description="STK Push payments will appear here"
            />
          </div>
        ) : (
          <div className="divide-y">
            {payments.map((p) => (
              <div key={p.id} className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium">{p.phone}</p>
                  <p className="text-sm text-gray-500">{new Date(p.created_at).toLocaleString()}</p>
                </div>
                <div className="text-right">
                  <p className="font-semibold">{formatCurrency(p.amount)}</p>
                  <span className={`text-xs px-2 py-1 rounded ${p.status === 'completed' ? 'bg-green-100 text-green-600' : 'bg-yellow-100 text-yellow-600'}`}>
                    {p.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
