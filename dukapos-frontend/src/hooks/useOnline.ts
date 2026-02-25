import { useState, useEffect, useCallback } from 'react'

interface UseOnlineReturn {
  isOnline: boolean
  wasOffline: boolean
  isOffline: boolean
}

export function useOnline(): UseOnlineReturn {
  const [isOnline, setIsOnline] = useState<boolean>(true)
  const [wasOffline, setWasOffline] = useState(false)

  useEffect(() => {
    // Check initial state
    if (typeof window !== 'undefined') {
      setIsOnline(navigator.onLine)
    }

    const handleOnline = () => {
      setIsOnline(true)
      setWasOffline(true)
    }

    const handleOffline = () => {
      setIsOnline(false)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return {
    isOnline,
    wasOffline,
    isOffline: !isOnline
  }
}

export function useNetworkStatus(callbacks?: {
  onOnline?: () => void
  onOffline?: () => void
}) {
  const [status, setStatus] = useState<'online' | 'offline'>(
    typeof window !== 'undefined' && navigator.onLine ? 'online' : 'offline'
  )

  useEffect(() => {
    const handleOnline = () => {
      setStatus('online')
      callbacks?.onOnline?.()
    }

    const handleOffline = () => {
      setStatus('offline')
      callbacks?.onOffline?.()
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [callbacks?.onOnline, callbacks?.onOffline])

  return {
    status,
    isOnline: status === 'online',
    isOffline: status === 'offline'
  }
}

export function useRetryConnection(maxRetries = 3, intervalMs = 2000) {
  const [isRetrying, setIsRetrying] = useState(false)
  const [retryCount, setRetryCount] = useState(0)
  const [lastRetryTime, setLastRetryTime] = useState<Date | null>(null)

  const retry = useCallback(async (): Promise<boolean> => {
    if (isRetrying) return false
    
    setIsRetrying(true)
    setRetryCount(prev => prev + 1)
    setLastRetryTime(new Date())

    for (let i = 0; i < maxRetries; i++) {
      if (navigator.onLine) {
        setIsRetrying(false)
        return true
      }
      await new Promise(resolve => setTimeout(resolve, intervalMs))
    }

    setIsRetrying(false)
    return navigator.onLine
  }, [isRetrying, maxRetries, intervalMs])

  const reset = useCallback(() => {
    setRetryCount(0)
    setLastRetryTime(null)
  }, [])

  return {
    isRetrying,
    retryCount,
    lastRetryTime,
    retry,
    reset
  }
}
