import { useEffect, useRef, useCallback, useState } from 'react'
import { useSyncStore } from '@/stores/syncStore'

interface UseBackgroundSyncOptions {
  enabled?: boolean
  intervalMs?: number
  onSync?: (result: unknown) => void
  onError?: (error: Error) => void
}

export function useBackgroundSync(options: UseBackgroundSyncOptions = {}) {
  const { enabled = true, intervalMs = 60000, onSync, onError } = options
  const { isOnline, isSyncing, syncNow, pendingCount, lastSyncTime } = useSyncStore()
  const [lastSyncResult, setLastSyncResult] = useState<unknown>(null)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [_syncTimeoutRef] = useState<ReturnType<typeof setTimeout> | null>(null)
  const syncIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const performSync = useCallback(async () => {
    if (!isOnline || isSyncing) return
    
    try {
      const result = await syncNow()
      setLastSyncResult(result)
      onSync?.(result)
    } catch (error) {
      onError?.(error as Error)
    }
  }, [isOnline, isSyncing, syncNow, onSync, onError])

  useEffect(() => {
    if (!enabled || !isOnline) return

    syncIntervalRef.current = setInterval(() => {
      if (pendingCount > 0) {
        performSync()
      }
    }, intervalMs)

    return () => {
      if (syncIntervalRef.current) {
        clearInterval(syncIntervalRef.current)
      }
    }
  }, [enabled, isOnline, intervalMs, pendingCount, performSync])

  useEffect(() => {
    if (!enabled || !isOnline) return

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && pendingCount > 0) {
        performSync()
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [enabled, isOnline, pendingCount, performSync])

  useEffect(() => {
    if (!enabled || !isOnline) return

    const handleOnline = () => {
      if (pendingCount > 0) {
        performSync()
      }
    }

    window.addEventListener('online', handleOnline)
    return () => {
      window.removeEventListener('online', handleOnline)
    }
  }, [enabled, isOnline, pendingCount, performSync])

  const triggerSync = useCallback(() => {
    if (isOnline && !isSyncing) {
      performSync()
    }
  }, [isOnline, isSyncing, performSync])

  return {
    triggerSync,
    lastSyncResult,
    isOnline,
    isSyncing,
    pendingCount,
    lastSyncTime
  }
}

export function useOfflineDetector() {
  const [isOffline, setIsOffline] = useState(
    typeof navigator !== 'undefined' ? !navigator.onLine : false
  )
  const [wasOffline, setWasOffline] = useState(false)

  useEffect(() => {
    const handleOnline = () => {
      setIsOffline(false)
      setWasOffline(true)
    }

    const handleOffline = () => {
      setIsOffline(true)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return { isOffline, wasOffline, setWasOffline }
}

export function SyncIndicator() {
  const { isOnline, pendingCount, lastSyncTime, isSyncing } = useSyncStore()
  
  if (isOnline && pendingCount === 0) return null
  
  return (
    <div className={`flex items-center gap-2 text-sm ${isOnline ? 'text-yellow-600' : 'text-red-600'}`}>
      {isSyncing ? (
        <>
          <span className="animate-spin">⟳</span>
          <span>Syncing...</span>
        </>
      ) : isOnline ? (
        <>
          <span>⚠</span>
          <span>{pendingCount} pending</span>
        </>
      ) : (
        <>
          <span>○</span>
          <span>Offline</span>
        </>
      )}
      {lastSyncTime && (
        <span className="text-gray-500">
          Last: {new Date(lastSyncTime).toLocaleTimeString()}
        </span>
      )}
    </div>
  )
}
