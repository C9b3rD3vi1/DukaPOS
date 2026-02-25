import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Badge } from '@/components/common/Badge'

interface WebhookEvent {
  id: number
  event_type: string
  payload: Record<string, unknown>
  status: 'pending' | 'success' | 'failed'
  response_code: number | null
  response_body: string | null
  attempts: number
  created_at: string
  delivered_at: string | null
}

export default function WebhookEvents() {
  const shop = useAuthStore((state) => state.shop)
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'success' | 'failed' | 'pending'>('all')
  const [selectedEvent, setSelectedEvent] = useState<WebhookEvent | null>(null)

  useEffect(() => {
    fetchEvents()
  }, [shop?.id, filter])

  const fetchEvents = async () => {
    if (!shop?.id) return
    setIsLoading(true)
    try {
      const params = new URLSearchParams()
      params.append('shop_id', shop.id.toString())
      if (filter !== 'all') params.append('status', filter)
      
      const response = await api.get<{ data: WebhookEvent[] }>(`/v1/webhooks/events?${params}`)
      setEvents(response.data?.data || response.data || [])
    } catch (error) {
      console.error('Failed to fetch webhook events:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const retryEvent = async (eventId: number) => {
    try {
      await api.post(`/v1/webhooks/events/${eventId}/retry`)
      fetchEvents()
    } catch (error) {
      console.error('Failed to retry webhook:', error)
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return <Badge variant="success">Delivered</Badge>
      case 'failed':
        return <Badge variant="danger">Failed</Badge>
      case 'pending':
        return <Badge variant="warning">Pending</Badge>
      default:
        return <Badge>{status}</Badge>
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('en-KE', {
      dateStyle: 'medium',
      timeStyle: 'short'
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Webhook Events</h1>
          <p className="text-gray-600 mt-1">Monitor webhook delivery and retry failed events</p>
        </div>
        <Button onClick={fetchEvents} variant="outline">
          Refresh
        </Button>
      </div>

      <div className="flex gap-2">
        {(['all', 'success', 'failed', 'pending'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition ${
              filter === f
                ? 'bg-primary text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Event</th>
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Status</th>
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Attempts</th>
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Response</th>
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Date</th>
                <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600">Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-gray-500">
                    Loading...
                  </td>
                </tr>
              ) : events.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-gray-500">
                    No webhook events found
                  </td>
                </tr>
              ) : (
                events.map((event) => (
                  <tr key={event.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="py-3 px-4">
                      <div>
                        <p className="font-medium text-gray-900">{event.event_type}</p>
                        <p className="text-xs text-gray-500">ID: {event.id}</p>
                      </div>
                    </td>
                    <td className="py-3 px-4">{getStatusBadge(event.status)}</td>
                    <td className="py-3 px-4 text-sm text-gray-600">{event.attempts}</td>
                    <td className="py-3 px-4">
                      {event.response_code ? (
                        <span className={`text-sm ${event.response_code >= 200 && event.response_code < 300 ? 'text-green-600' : 'text-red-600'}`}>
                          {event.response_code}
                        </span>
                      ) : (
                        <span className="text-sm text-gray-400">-</span>
                      )}
                    </td>
                    <td className="py-3 px-4 text-sm text-gray-600">
                      {formatDate(event.created_at)}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setSelectedEvent(event)}
                        >
                          View
                        </Button>
                        {event.status === 'failed' && (
                          <Button
                            size="sm"
                            onClick={() => retryEvent(event.id)}
                          >
                            Retry
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {selectedEvent && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-2xl max-h-[80vh] overflow-auto">
            <div className="flex justify-between items-start mb-4">
              <h2 className="text-lg font-bold">Event Details</h2>
              <button
                onClick={() => setSelectedEvent(null)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>
            
            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-semibold text-gray-600 mb-1">Event Type</h3>
                <p className="font-mono bg-gray-100 p-2 rounded">{selectedEvent.event_type}</p>
              </div>
              
              <div>
                <h3 className="text-sm font-semibold text-gray-600 mb-1">Payload</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded text-xs overflow-auto max-h-48">
                  {JSON.stringify(selectedEvent.payload, null, 2)}
                </pre>
              </div>
              
              {selectedEvent.response_body && (
                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-1">Response</h3>
                  <pre className="bg-gray-100 p-4 rounded text-xs overflow-auto max-h-48">
                    {selectedEvent.response_body}
                  </pre>
                </div>
              )}
              
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-500">Created:</span> {formatDate(selectedEvent.created_at)}
                </div>
                <div>
                  <span className="text-gray-500">Delivered:</span> {selectedEvent.delivered_at ? formatDate(selectedEvent.delivered_at) : '-'}
                </div>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
