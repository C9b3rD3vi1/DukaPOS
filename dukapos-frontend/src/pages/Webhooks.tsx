import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card, StatCard } from '@/components/common/Card'
import { Skeleton } from '@/components/common/Skeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { Webhook } from '@/api/types'

export default function Webhooks() {
  const shop = useAuthStore((state) => state.shop)
  const [webhooks, setWebhooks] = useState<Webhook[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [testing, setTesting] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', url: '', events: [] as string[] })

  const availableEvents = ['sale.created', 'sale.updated', 'product.low_stock', 'customer.created', 'payment.received']

  useEffect(() => { fetchWebhooks() }, [shop?.id])

  const fetchWebhooks = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get<{ data: any[] }>('/v1/webhooks')
      const webhooksData = response.data?.data || []
      // Convert events string to array if needed
      const processed = webhooksData.map((w: any) => ({
        ...w,
        events: typeof w.events === 'string' ? w.events.split(',').map((e: string) => e.trim()) : w.events
      }))
      setWebhooks(processed)
    } catch (err) { console.error(err) }
    finally { setIsLoading(false) }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/v1/webhooks', form)
      setShowCreateModal(false)
      setForm({ name: '', url: '', events: [] })
      fetchWebhooks()
    } catch (err) { console.error(err) }
  }

  const handleToggle = async (id: number, isActive: boolean) => {
    try {
      await api.put(`/v1/webhooks/${id}`, { is_active: !isActive })
      fetchWebhooks()
    } catch (err) { console.error(err) }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this webhook?')) return
    try {
      await api.delete(`/v1/webhooks/${id}`)
      fetchWebhooks()
    } catch (err) { console.error(err) }
  }

  const handleTest = async (id: number) => {
    setTesting(id)
    try {
      await api.post(`/v1/webhooks/${id}/test`)
      alert('Test event sent!')
    } catch (err) { console.error(err) }
    finally { setTesting(null) }
  }

  const totalWebhooks = webhooks.length
  const activeWebhooks = webhooks.filter(w => w.is_active).length

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <Skeleton className="h-8 w-32 mb-2" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="space-y-3">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Webhooks</h1>
          <p className="text-gray-500 mt-1">Configure webhook endpoints for real-time events</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-xl hover:bg-primary-dark transition"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Webhook
        </button>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 mb-6">
        <div className="flex items-start gap-3">
          <svg className="w-5 h-5 text-blue-600 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <p className="font-medium text-blue-800">What are webhooks?</p>
            <p className="text-sm text-blue-700">Webhooks allow external services to receive real-time notifications when events occur in your store.</p>
          </div>
        </div>
      </div>

      {/* Stats */}
      {shop && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
          <StatCard
            title="Total Webhooks"
            value={totalWebhooks}
            variant="default"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            }
          />
          <StatCard
            title="Active"
            value={activeWebhooks}
            variant="success"
            icon={
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
        </div>
      )}

      {webhooks.length === 0 ? (
        <Card className="text-center py-12">
          <EmptyState
            variant="generic"
            title={!shop ? 'No Shop Selected' : 'No Webhooks Configured'}
            description={!shop ? 'Please select a shop to view webhooks' : 'Add a webhook to receive real-time notifications'}
            action={shop ? {
              label: 'Add Webhook',
              onClick: () => setShowCreateModal(true),
            } : undefined}
          />
        </Card>
      ) : (
        <div className="space-y-4">
          {webhooks.map((webhook) => (
            <div key={webhook.id} className="bg-white rounded-2xl border border-gray-100 shadow-sm p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className={`w-3 h-3 rounded-full ${webhook.is_active ? 'bg-green-500' : 'bg-gray-300'}`}></div>
                  <div>
                    <p className="font-medium text-gray-900">{webhook.url}</p>
                    <p className="text-sm text-gray-500">{webhook.events.join(', ')}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleTest(webhook.id)}
                    disabled={testing === webhook.id}
                    className="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition"
                  >
                    {testing === webhook.id ? 'Testing...' : 'Test'}
                  </button>
                  <button
                    onClick={() => handleToggle(webhook.id, webhook.is_active)}
                    className={`px-3 py-1.5 text-sm rounded-lg transition ${
                      webhook.is_active ? 'text-amber-600 hover:bg-amber-50' : 'text-green-600 hover:bg-green-50'
                    }`}
                  >
                    {webhook.is_active ? 'Disable' : 'Enable'}
                  </button>
                  <button
                    onClick={() => handleDelete(webhook.id)}
                    className="p-1.5 text-red-600 hover:bg-red-50 rounded-lg transition"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl">
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-lg font-bold text-gray-900">Add Webhook</h3>
            </div>
            <form onSubmit={handleCreate} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="My Webhook"
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Webhook URL</label>
                <input
                  type="url"
                  value={form.url}
                  onChange={(e) => setForm({ ...form, url: e.target.value })}
                  placeholder="https://your-server.com/webhook"
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Events</label>
                <div className="space-y-2">
                  {availableEvents.map((event) => (
                    <label key={event} className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={form.events.includes(event)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setForm({ ...form, events: [...form.events, event] })
                          } else {
                            setForm({ ...form, events: form.events.filter(e => e !== event) })
                          }
                        }}
                        className="w-4 h-4 text-primary rounded"
                      />
                      <span className="text-sm text-gray-600">{event}</span>
                    </label>
                  ))}
                </div>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark"
                >
                  Create Webhook
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
