import { useEffect, useRef, useCallback, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'

export type WebSocketMessageType = 
  | 'new_sale'
  | 'low_stock'
  | 'payment_received'
  | 'order_update'
  | 'stock_sync'
  | 'pong'

export interface WebSocketMessage {
  type: WebSocketMessageType
  payload: unknown
  timestamp: number
}

export interface UseWebSocketOptions {
  onNewSale?: (data: { product: string; amount: number }) => void
  onLowStock?: (data: { product: string; current_stock: number }) => void
  onPaymentReceived?: (data: { amount: number; phone: string }) => void
  onOrderUpdate?: (data: { order_id: number; status: string }) => void
  onStockSync?: (data: { product_id: number; quantity: number }) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const isConnected = useRef(false)
  
  const token = useAuthStore((state) => state.token)
  const shop = useAuthStore((state) => state.shop)
  const syncStore = useSyncStore()

  const connect = useCallback(() => {
    if (!token || !shop?.id || isConnected.current) return

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?token=${token}&shop_id=${shop.id}`
    
    try {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        console.log('WebSocket connected')
        isConnected.current = true
        reconnectAttempts.current = 0
        syncStore.setOnline(true)
        options.onConnect?.()
        
        // Subscribe to shop updates
        ws.send(JSON.stringify({
          type: 'subscribe',
          payload: { shop_id: shop.id }
        }))
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          
          switch (message.type) {
            case 'new_sale':
              options.onNewSale?.(message.payload as { product: string; amount: number })
              break
            case 'low_stock':
              options.onLowStock?.(message.payload as { product: string; current_stock: number })
              break
            case 'payment_received':
              options.onPaymentReceived?.(message.payload as { amount: number; phone: string })
              break
            case 'order_update':
              options.onOrderUpdate?.(message.payload as { order_id: number; status: string })
              break
            case 'stock_sync':
              options.onStockSync?.(message.payload as { product_id: number; quantity: number })
              break
            case 'pong':
              // Heartbeat response
              break
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        options.onError?.(error)
      }

      ws.onclose = () => {
        console.log('WebSocket disconnected')
        isConnected.current = false
        syncStore.setOnline(false)
        options.onDisconnect?.()
        
        // Attempt to reconnect
        if (reconnectAttempts.current < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000)
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttempts.current++
            connect()
          }, delay)
        }
      }
    } catch (e) {
      console.error('Failed to create WebSocket:', e)
    }
  }, [token, shop?.id, options, syncStore])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    isConnected.current = false
  }, [])

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  const subscribe = useCallback((channel: string) => {
    send({ type: 'subscribe', payload: { channel } })
  }, [send])

  // Heartbeat
  useEffect(() => {
    const heartbeat = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)

    return () => clearInterval(heartbeat)
  }, [])

  // Auto-connect on mount
  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  return {
    isConnected: isConnected.current,
    connect,
    disconnect,
    send,
    subscribe
  }
}

// Toast/Web notification hook for real-time alerts
export function useRealTimeAlerts() {
  const [alerts, setAlerts] = useState<Array<{ id: string; type: string; message: string; timestamp: number }>>([])
  
  const ws = useWebSocket({
    onNewSale: (data) => {
      addAlert('sale', `New sale: ${data.product} - KES ${data.amount}`)
    },
    onLowStock: (data) => {
      addAlert('warning', `Low stock: ${data.product} (${data.current_stock} remaining)`)
    },
    onPaymentReceived: (data) => {
      addAlert('payment', `Payment received: KES ${data.amount} from ${data.phone}`)
    },
    onOrderUpdate: (data) => {
      addAlert('order', `Order #${data.order_id} status: ${data.status}`)
    }
  })

  const addAlert = (type: string, message: string) => {
    const id = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    setAlerts(prev => [...prev, { id, type, message, timestamp: Date.now() }])
    
    // Auto-remove after 10 seconds
    setTimeout(() => {
      setAlerts(prev => prev.filter(a => a.id !== id))
    }, 10000)
  }

  const dismissAlert = (id: string) => {
    setAlerts(prev => prev.filter(a => a.id !== id))
  }

  return { alerts, dismissAlert, ...ws }
}
