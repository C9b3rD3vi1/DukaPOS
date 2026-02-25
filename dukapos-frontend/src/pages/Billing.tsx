import { useState, useEffect } from 'react'
import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

interface SubscriptionPlan {
  id: number
  name: string
  description?: string
  price: number
  interval: string
  features: string[]
  product_limit: number
  shop_limit: number
  staff_limit: number
  is_popular: boolean
}

interface CurrentSubscription {
  id: number
  plan_id: number
  status: string
  current_period_start: string
  current_period_end: string
  cancel_at_period_end: boolean
  plan?: SubscriptionPlan
}

interface BillingHistoryItem {
  id: number
  amount: number
  currency: string
  status: string
  description: string
  created_at: string
  invoice_url?: string
}

export default function Billing() {
  const shop = useAuthStore((state) => state.shop)
  const token = useAuthStore((state) => state.token)
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [currentSubscription, setCurrentSubscription] = useState<CurrentSubscription | null>(null)
  const [billingHistory, setBillingHistory] = useState<BillingHistoryItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [upgrading, setUpgrading] = useState(false)
  const [error, setError] = useState('')
  const [showConfirmModal, setShowConfirmModal] = useState(false)
  const [selectedPlanId, setSelectedPlanId] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'plans' | 'history'>('plans')

  useEffect(() => {
    if (shop?.id && token) {
      fetchBillingData()
    } else {
      setIsLoading(false)
    }
  }, [shop?.id, token])

  const fetchBillingData = async () => {
    if (!token) return
    try {
      const headers = { Authorization: `Bearer ${token}` }
      
      const [plansRes, subscriptionRes, historyRes] = await Promise.all([
        axios.get(`${API_BASE_URL}/api/v1/subscriptions/plans`, { headers }),
        axios.get(`${API_BASE_URL}/api/v1/subscriptions/current`, { headers }).catch(() => ({ data: { data: null } })),
        axios.get(`${API_BASE_URL}/api/v1/billing/history`, { headers }).catch(() => ({ data: { data: [] } }))
      ])

      const plansData = plansRes.data?.plans || plansRes.data?.data || plansRes.data || []
      setPlans(Array.isArray(plansData) ? plansData : [])

      const subData = subscriptionRes.data?.data || subscriptionRes.data || null
      setCurrentSubscription(subData)

      const historyData = historyRes.data?.data || historyRes.data || []
      setBillingHistory(Array.isArray(historyData) ? historyData : [])
    } catch (err) {
      console.error(err)
      setError('Unable to load subscription information')
    } finally {
      setIsLoading(false)
    }
  }

  const handleUpgradeClick = (planId: number) => {
    const currentPlanId = currentSubscription?.plan?.id
    if (currentPlanId === planId) return
    setSelectedPlanId(planId)
    setShowConfirmModal(true)
  }

  const handleConfirmUpgrade = async () => {
    if (!token || !selectedPlanId) return
    setUpgrading(true)
    setError('')
    try {
      await axios.post(`${API_BASE_URL}/api/v1/subscriptions/upgrade`,
        { plan_id: selectedPlanId },
        { headers: { Authorization: `Bearer ${token}` }}
      )
      await fetchBillingData()
      setShowConfirmModal(false)
      setSelectedPlanId(null)
    } catch (err: any) {
      console.error(err)
      setError(err.response?.data?.error || 'Failed to upgrade subscription. Please try again.')
    } finally {
      setUpgrading(false)
    }
  }

  const handleCancelSubscription = async () => {
    if (!token || !confirm('Are you sure you want to cancel your subscription? You will lose access to premium features at the end of your billing period.')) return
    
    try {
      await axios.post(`${API_BASE_URL}/api/v1/subscriptions/cancel`,
        {},
        { headers: { Authorization: `Bearer ${token}` }}
      )
      await fetchBillingData()
      alert('Subscription will be cancelled at the end of the billing period.')
    } catch (err) {
      console.error(err)
      setError('Failed to cancel subscription')
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const getCurrentPlan = () => {
    if (currentSubscription?.plan) return currentSubscription.plan
    if (plans.length > 0) return plans[0]
    return null
  }

  const getCurrentPlanIndex = () => {
    const currentId = currentSubscription?.plan_id
    if (!currentId) return 0
    const index = plans.findIndex(p => p.id === currentId)
    return index >= 0 ? index : 0
  }

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'paid':
      case 'active':
        return 'bg-green-100 text-green-700'
      case 'pending':
        return 'bg-yellow-100 text-yellow-700'
      case 'failed':
      case 'cancelled':
        return 'bg-red-100 text-red-700'
      default:
        return 'bg-surface-100 text-surface-600'
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    )
  }

  const currentPlan = getCurrentPlan()
  const currentIndex = getCurrentPlanIndex()

  return (
    <div className="-mx-4 md:-mx-6">
      <div className="px-4 md:px-6 pb-6">
        <div className="mb-6">
          <h1 className="text-2xl md:text-3xl font-bold text-surface-900">Subscription & Billing</h1>
          <p className="text-surface-500 mt-1">Manage your subscription and view billing history</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          <button
            onClick={() => setActiveTab('plans')}
            className={`px-4 py-2 rounded-xl font-medium transition ${
              activeTab === 'plans'
                ? 'bg-primary text-white shadow-md shadow-primary/25'
                : 'bg-surface-100 text-surface-600 hover:bg-surface-200'
            }`}
          >
            Plans
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`px-4 py-2 rounded-xl font-medium transition ${
              activeTab === 'history'
                ? 'bg-primary text-white shadow-md shadow-primary/25'
                : 'bg-surface-100 text-surface-600 hover:bg-surface-200'
            }`}
          >
            Billing History
          </button>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 text-red-600 rounded-xl">
            {error}
          </div>
        )}

        {/* Plans Tab */}
        {activeTab === 'plans' && (
          <>
            {/* Current Plan Banner */}
            {currentSubscription && (
              <Card className="mb-6 bg-gradient-to-r from-primary to-primary-dark text-white">
                <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                  <div>
                    <p className="text-white/80 text-sm">Current Plan</p>
                    <p className="text-2xl font-bold">{currentPlan?.name || 'Starter'}</p>
                    <span className="inline-flex items-center gap-1 mt-1 px-2 py-0.5 bg-white/20 rounded-full text-xs capitalize">
                      <span className={`w-1.5 h-1.5 ${currentSubscription.status === 'active' ? 'bg-green-400' : 'bg-yellow-400'} rounded-full`}></span>
                      {currentSubscription.status}
                    </span>
                  </div>
                  <div className="text-right">
                    <p className="text-white/80 text-sm">Billing Period</p>
                    <p className="font-medium">
                      {new Date(currentSubscription.current_period_start).toLocaleDateString()} - {new Date(currentSubscription.current_period_end).toLocaleDateString()}
                    </p>
                    {currentSubscription.cancel_at_period_end && (
                      <span className="inline-block mt-1 px-2 py-0.5 bg-red-500/20 text-red-200 text-xs rounded">
                        Cancelling
                      </span>
                    )}
                  </div>
                </div>
                {currentSubscription.status === 'active' && !currentSubscription.cancel_at_period_end && (
                  <div className="mt-4 pt-4 border-t border-white/20">
                    <button
                      onClick={handleCancelSubscription}
                      className="text-sm text-white/70 hover:text-white underline"
                    >
                      Cancel subscription
                    </button>
                  </div>
                )}
              </Card>
            )}

            {/* Plans Grid */}
            {plans.length > 0 ? (
              <div className="grid md:grid-cols-3 gap-4 md:gap-6 mb-8">
                {plans.map((plan, index) => {
                  const currentPlanId = currentSubscription?.plan?.id || currentSubscription?.plan_id
                  const isCurrentPlan = currentPlanId === plan.id
                  const isPopular = plan.is_popular
                  
                  return (
                    <Card 
                      key={plan.id} 
                      className={`relative overflow-hidden ${isCurrentPlan ? 'ring-2 ring-primary' : ''}`}
                    >
                      {isPopular && (
                        <div className="absolute top-0 right-0 bg-gradient-to-r from-orange-500 to-amber-500 text-white text-xs font-bold px-3 py-1 rounded-bl-xl">
                          POPULAR
                        </div>
                      )}
                      {isCurrentPlan && (
                        <div className="absolute top-0 left-0 bg-green-500 text-white text-xs font-bold px-3 py-1 rounded-br-xl">
                          CURRENT
                        </div>
                      )}

                      <div className="p-6">
                        <h3 className="text-xl font-bold text-surface-900">{plan.name}</h3>
                        <p className="text-sm text-surface-500 mt-1">{plan.description}</p>
                        
                        <div className="mt-4 mb-6">
                          <span className="text-4xl font-bold text-surface-900">
                            {plan.price === 0 ? 'Free' : formatCurrency(plan.price)}
                          </span>
                          {plan.price > 0 && <span className="text-surface-500">/{plan.interval}</span>}
                        </div>

                        <ul className="space-y-3 mb-6">
                          {plan.features.map((feature, i) => (
                            <li key={i} className="flex items-start gap-2 text-sm text-surface-700">
                              <svg className="w-5 h-5 flex-shrink-0 mt-0.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                              </svg>
                              {feature}
                            </li>
                          ))}
                        </ul>

                        <Button
                          onClick={() => handleUpgradeClick(plan.id)}
                          disabled={upgrading || isCurrentPlan}
                          variant={isCurrentPlan ? 'secondary' : 'primary'}
                          className="w-full"
                        >
                          {isCurrentPlan ? 'Current Plan' : upgrading ? 'Processing...' : index > currentIndex ? 'Upgrade' : 'Select'}
                        </Button>
                      </div>
                    </Card>
                  )
                })}
              </div>
            ) : (
              <Card className="mb-6 text-center py-8">
                <div className="w-16 h-16 bg-surface-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-surface-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-lg font-semibold text-surface-900 mb-2">No Plans Available</h3>
                <p className="text-surface-500">Subscription plans are not currently available. Please contact support.</p>
              </Card>
            )}

            {/* Payment Methods */}
            <Card>
              <h3 className="text-lg font-semibold text-surface-900 mb-4">Payment Methods</h3>
              <div className="grid md:grid-cols-3 gap-4">
                <div className="flex items-center gap-3 p-4 bg-surface-50 rounded-xl">
                  <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                    <span className="text-green-600 font-bold text-sm">M</span>
                  </div>
                  <div>
                    <p className="font-medium text-surface-900">M-Pesa</p>
                    <p className="text-sm text-surface-500">Instant payment</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 p-4 bg-surface-50 rounded-xl">
                  <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                    <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-surface-900">Card</p>
                    <p className="text-sm text-surface-500">Visa, Mastercard</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 p-4 bg-surface-50 rounded-xl">
                  <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                    <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-surface-900">Bank Transfer</p>
                    <p className="text-sm text-surface-500">Direct deposit</p>
                  </div>
                </div>
              </div>
              
              <div className="mt-4 p-4 bg-surface-50 rounded-xl flex items-center gap-3">
                <svg className="w-5 h-5 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <p className="text-sm text-surface-600">
                  All payments are processed securely. Need help? Contact <a href="mailto:support@dukapos.com" className="text-primary font-medium">support@dukapos.com</a>
                </p>
              </div>
            </Card>
          </>
        )}

        {/* Billing History Tab */}
        {activeTab === 'history' && (
          <Card padding="none">
            <div className="p-4 border-b border-surface-100">
              <h2 className="font-semibold text-surface-900">Billing History</h2>
              <p className="text-sm text-surface-500">View your past invoices and payments</p>
            </div>
            {billingHistory.length === 0 ? (
              <div className="p-8 text-center text-surface-500">
                <svg className="w-12 h-12 mx-auto mb-3 text-surface-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="mb-2">No billing history yet</p>
                <p className="text-sm">Your payment history will appear here</p>
              </div>
            ) : (
              <div className="divide-y divide-surface-100">
                {billingHistory.map((item) => (
                  <div key={item.id} className="p-4 flex items-center justify-between hover:bg-surface-50">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 bg-surface-100 rounded-xl flex items-center justify-center">
                        <svg className="w-5 h-5 text-surface-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                      </div>
                      <div>
                        <p className="font-medium text-surface-900">{item.description}</p>
                        <p className="text-sm text-surface-500">{new Date(item.created_at).toLocaleDateString()} at {new Date(item.created_at).toLocaleTimeString()}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right">
                        <p className="font-semibold text-surface-900">{formatCurrency(item.amount)}</p>
                        <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full capitalize ${getStatusColor(item.status)}`}>
                          {item.status}
                        </span>
                      </div>
                      {item.invoice_url && (
                        <a
                          href={item.invoice_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="p-2 text-surface-500 hover:text-primary hover:bg-primary/10 rounded-lg transition"
                          title="View Invoice"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                          </svg>
                        </a>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        )}
      </div>

      {/* Confirmation Modal */}
      {showConfirmModal && selectedPlanId && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-end md:items-center justify-center">
          <div className="bg-white w-full max-w-md rounded-t-3xl md:rounded-3xl shadow-2xl">
            <div className="p-6 border-b border-surface-100">
              <h3 className="text-xl font-bold text-surface-900">Confirm Plan Change</h3>
            </div>
            <div className="p-6">
              <p className="text-surface-600 mb-4">
                You are about to switch to the <strong>{plans.find(p => p.id === selectedPlanId)?.name}</strong> plan.
              </p>
              {plans.find(p => p.id === selectedPlanId)?.price && plans.find(p => p.id === selectedPlanId)!.price > 0 && (
                <p className="text-surface-600 mb-4">
                  Your card will be charged <strong>{formatCurrency(plans.find(p => p.id === selectedPlanId)?.price || 0)}</strong> per {plans.find(p => p.id === selectedPlanId)?.interval}.
                </p>
              )}
              <p className="text-sm text-surface-500">
                You can cancel or change your plan at any time from this page.
              </p>
            </div>
            <div className="p-6 border-t border-surface-100 flex gap-3">
              <Button
                variant="secondary"
                onClick={() => { setShowConfirmModal(false); setSelectedPlanId(null); }}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={handleConfirmUpgrade}
                disabled={upgrading}
                className="flex-1"
              >
                {upgrading ? 'Processing...' : 'Confirm'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
