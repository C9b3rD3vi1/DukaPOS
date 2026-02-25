import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard, StatGrid } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/common/Button'
import type { Customer, LoyaltyCustomer } from '@/api/types'

interface LoyaltyStats {
  total_members: number
  total_points: number
  total_redemptions: number
  active_this_month: number
}

const LOYALTY_TIERS = {
  bronze: { min: 0, color: 'text-orange-700', bg: 'bg-orange-100', label: 'Bronze' },
  silver: { min: 500, color: 'text-gray-500', bg: 'bg-gray-200', label: 'Silver' },
  gold: { min: 2000, color: 'text-yellow-600', bg: 'bg-yellow-100', label: 'Gold' },
  platinum: { min: 5000, color: 'text-purple-600', bg: 'bg-purple-100', label: 'Platinum' }
}

function calculateTier(points: number): keyof typeof LOYALTY_TIERS {
  if (points >= LOYALTY_TIERS.platinum.min) return 'platinum'
  if (points >= LOYALTY_TIERS.gold.min) return 'gold'
  if (points >= LOYALTY_TIERS.silver.min) return 'silver'
  return 'bronze'
}

function getPointsToNextTier(points: number): { next: string | null; points: number } {
  if (points >= LOYALTY_TIERS.platinum.min) return { next: null, points: 0 }
  if (points >= LOYALTY_TIERS.gold.min) return { next: 'Platinum', points: LOYALTY_TIERS.platinum.min - points }
  if (points >= LOYALTY_TIERS.silver.min) return { next: 'Gold', points: LOYALTY_TIERS.gold.min - points }
  return { next: 'Silver', points: LOYALTY_TIERS.silver.min - points }
}

export function LoyaltyPanel() {
  const shop = useAuthStore((state) => state.shop)
  const [stats, setStats] = useState<LoyaltyStats>({
    total_members: 0,
    total_points: 0,
    total_redemptions: 0,
    active_this_month: 0
  })
  const [members, setMembers] = useState<LoyaltyCustomer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showPointsModal, setShowPointsModal] = useState(false)
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [pointsAmount, setPointsAmount] = useState('')
  const [pointsAction, setPointsAction] = useState<'add' | 'redeem'>('add')

  useEffect(() => {
    fetchLoyaltyData()
  }, [shop?.id])

  const fetchLoyaltyData = async () => {
    if (!shop?.id) {
      setIsLoading(false)
      return
    }
    setIsLoading(true)
    
    try {
      const [statsRes, membersRes] = await Promise.all([
        api.get(`/v1/loyalty/stats/shop/${shop.id}`).catch(() => ({ data: null })),
        api.get(`/v1/loyalty/members?shop_id=${shop.id}`).catch(() => ({ data: { data: [] } }))
      ])
      
      if (statsRes.data) {
        setStats(statsRes.data as LoyaltyStats)
      }
      const membersData = membersRes.data?.data || membersRes.data || []
      setMembers(Array.isArray(membersData) ? membersData : [])
    } catch (e) {
      console.error('Failed to fetch loyalty data:', e)
    } finally {
      setIsLoading(false)
    }
  }

  const handlePointsUpdate = async () => {
    if (!selectedCustomer || !pointsAmount) return
    
    try {
      const endpoint = pointsAction === 'add' 
        ? `/v1/loyalty/points/add`
        : `/v1/loyalty/points/redeem`
      
      await api.post(endpoint, {
        customer_id: selectedCustomer.id,
        shop_id: shop?.id,
        points: Number(pointsAmount)
      })
      
      setShowPointsModal(false)
      setPointsAmount('')
      setSelectedCustomer(null)
      fetchLoyaltyData()
    } catch (e) {
      console.error('Failed to update points:', e)
    }
  }

  const formatCurrency = (amount: number) => 
    new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES', minimumFractionDigits: 0 }).format(amount || 0)

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 bg-gray-200 animate-pulse rounded"></div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
        </div>
        <div className="space-y-3">
          <div className="h-16 bg-gray-200 animate-pulse rounded"></div>
          <div className="h-16 bg-gray-200 animate-pulse rounded"></div>
          <div className="h-16 bg-gray-200 animate-pulse rounded"></div>
        </div>
      </div>
    )
  }

  if (!shop?.id) {
    return (
      <Card className="text-center py-12">
        <EmptyState
          variant="generic"
          title="No Shop Selected"
          description="Please select a shop to view loyalty program"
        />
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Stats */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Loyalty Program</h2>
        <StatGrid>
          <StatCard
            title="Total Members"
            value={stats.total_members}
            variant="info"
          />
          <StatCard
            title="Points in Circulation"
            value={stats.total_points?.toLocaleString() || '0'}
            variant="default"
          />
          <StatCard
            title="Points Redeemed"
            value={stats.total_redemptions?.toLocaleString() || '0'}
            variant="success"
          />
          <StatCard
            title="Active This Month"
            value={stats.active_this_month}
            variant="warning"
          />
        </StatGrid>
      </div>

      {/* Members List */}
      <Card>
        <div className="p-4 border-b border-gray-100">
          <h3 className="font-semibold text-gray-900">Loyalty Members</h3>
        </div>
        
        {members.length === 0 ? (
          <EmptyState
            variant="generic"
            title="No Loyalty Members"
            description="Customers will automatically join when they make their first purchase"
          />
        ) : (
          <div className="divide-y divide-gray-100">
            {members.map((member) => {
              const tier = calculateTier(member.points)
              const tierInfo = LOYALTY_TIERS[tier]
              const nextTier = getPointsToNextTier(member.points)
              
              return (
                <div key={member.id} className="p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className={`w-10 h-10 rounded-full ${tierInfo.bg} ${tierInfo.color} flex items-center justify-center font-bold`}>
                        {member.customer_id.toString().slice(-2)}
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">Customer #{member.customer_id}</p>
                        <p className="text-sm text-gray-500">
                          {member.visits} visits • {formatCurrency(member.total_spent)} spent
                        </p>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <div className="flex items-center gap-2">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${tierInfo.bg} ${tierInfo.color}`}>
                          {tierInfo.label}
                        </span>
                      </div>
                      <p className="text-sm font-medium text-gray-900 mt-1">
                        {member.points.toLocaleString()} points
                      </p>
                      {nextTier.next && (
                        <p className="text-xs text-gray-500">
                          {nextTier.points.toLocaleString()} to {nextTier.next}
                        </p>
                      )}
                    </div>
                  </div>
                  
                  <div className="mt-3 flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        setSelectedCustomer({ id: member.customer_id } as Customer)
                        setPointsAction('add')
                        setShowPointsModal(true)
                      }}
                    >
                      Add Points
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        setSelectedCustomer({ id: member.customer_id } as Customer)
                        setPointsAction('redeem')
                        setShowPointsModal(true)
                      }}
                    >
                      Redeem
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </Card>

      {/* Points Modal */}
      {showPointsModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-sm">
            <div className="p-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900">
                {pointsAction === 'add' ? 'Add Points' : 'Redeem Points'}
              </h3>
            </div>
            <div className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Points Amount
                </label>
                <input
                  type="number"
                  value={pointsAmount}
                  onChange={(e) => setPointsAmount(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
                  placeholder="Enter points"
                  min="1"
                />
              </div>
              <p className="text-sm text-gray-500">
                {pointsAction === 'add' 
                  ? 'Add points to customer account (1 point per KES spent)'
                  : 'Redeem points for discounts (100 points = KES 10)'
                }
              </p>
            </div>
            <div className="p-4 border-t border-gray-100 flex gap-3">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setShowPointsModal(false)}
              >
                Cancel
              </Button>
              <Button
                className="flex-1"
                onClick={handlePointsUpdate}
                disabled={!pointsAmount}
              >
                {pointsAction === 'add' ? 'Add Points' : 'Redeem'}
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}

export function LoyaltyPointsEarned(amount: number): number {
  // 1 point per KES spent
  return Math.floor(amount)
}

export function LoyaltyPointsValue(points: number): number {
  // 100 points = KES 10
  return Math.floor(points / 10)
}
