import { useCallback, useEffect, useState } from 'react'
import { useSyncStore } from '@/stores/syncStore'
import { useOnline } from './useOnline'
import type { SyncResult } from '@/db/sync'

interface UseSyncOptions {
  autoSync?: boolean
  autoSyncInterval?: number
  onSyncComplete?: (result: SyncResult) => void
  onError?: (error: string) => void
}

export function useSync(options: UseSyncOptions = {}) {
  const {
    autoSync = true,
    autoSyncInterval = 30000,
    onSyncComplete,
    onError
  } = options

  const { 
    isOnline, 
    isSyncing, 
    pendingCount, 
    lastSyncTime,
    syncError,
    syncNow,
    getPendingCount,
    enableAutoSync,
    disableAutoSync
  } = useSyncStore()

  const networkStatus = useOnline()
  const [syncResult, setSyncResult] = useState<SyncResult | null>(null)

  // Enable/disable auto sync based on online status
  useEffect(() => {
    if (autoSync && isOnline) {
      enableAutoSync(autoSyncInterval)
    } else {
      disableAutoSync()
    }

    return () => {
      disableAutoSync()
    }
  }, [autoSync, isOnline, autoSyncInterval, enableAutoSync, disableAutoSync])

  // Sync when coming back online
  useEffect(() => {
    if (networkStatus.wasOffline && isOnline) {
      handleSync()
    }
  }, [networkStatus.wasOffline, isOnline])

  const handleSync = useCallback(async () => {
    if (!isOnline) {
      onError?.('Cannot sync while offline')
      return
    }

    try {
      const result = await syncNow()
      setSyncResult(result)
      onSyncComplete?.(result)
      
      if (!result.success && result.errors.length > 0) {
        result.errors.forEach(err => onError?.(err))
      }
      
      return result
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Sync failed'
      onError?.(message)
      return { success: false, synced: 0, failed: 0, errors: [message] }
    }
  }, [isOnline, syncNow, onSyncComplete, onError])

  const handleGetPendingCount = useCallback(async () => {
    await getPendingCount()
  }, [getPendingCount])

  const isReady = isOnline && !isSyncing

  const canSync = isOnline && !isSyncing && pendingCount > 0

  const getLastSyncText = useCallback(() => {
    if (!lastSyncTime) return 'Never'
    const now = new Date()
    const diff = now.getTime() - lastSyncTime.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)

    if (minutes < 1) return 'Just now'
    if (minutes < 60) return `${minutes}m ago`
    if (hours < 24) return `${hours}h ago`
    return lastSyncTime.toLocaleDateString()
  }, [lastSyncTime])

  return {
    // State
    isOnline,
    isOffline: !isOnline,
    isSyncing,
    pendingCount,
    lastSyncTime,
    syncError,
    syncResult,
    
    // Computed
    isReady,
    canSync,
    lastSyncText: getLastSyncText(),
    
    // Actions
    sync: handleSync,
    getPendingCount: handleGetPendingCount,
    
    // Formatted helpers
    pendingCountText: pendingCount === 0 ? 'All synced' : `${pendingCount} pending`,
    statusText: isSyncing ? 'Syncing...' : !isOnline ? 'Offline' : pendingCount > 0 ? `${pendingCount} pending` : 'Synced'
  }
}

export function useSyncStatus() {
  const { isOnline, isSyncing, pendingCount, lastSyncTime, syncNow, getPendingCount } = useSyncStore()
  
  const getSyncStatus = useCallback(() => {
    if (isSyncing) return 'syncing'
    if (!isOnline) return 'offline'
    if (pendingCount > 0) return 'pending'
    return 'synced'
  }, [isSyncing, isOnline, pendingCount])

  const getStatusColor = useCallback(() => {
    switch (getSyncStatus()) {
      case 'syncing': return 'text-blue-500'
      case 'offline': return 'text-amber-500'
      case 'pending': return 'text-amber-500'
      case 'synced': return 'text-green-500'
      default: return 'text-gray-500'
    }
  }, [getSyncStatus])

  const getStatusIcon = useCallback(() => {
    switch (getSyncStatus()) {
      case 'syncing': return '↻'
      case 'offline': return '○'
      case 'pending': return '◔'
      case 'synced': return '✓'
      default: return '?'
    }
  }, [getSyncStatus])

  return {
    isOnline,
    isSyncing,
    pendingCount,
    lastSyncTime,
    syncNow,
    getPendingCount,
    status: getSyncStatus(),
    statusColor: getStatusColor(),
    statusIcon: getStatusIcon()
  }
}
