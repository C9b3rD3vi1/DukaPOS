import { useSyncStore } from '@/stores/syncStore'
import { useMemo } from 'react'

interface SyncStatusProps {
  showDetails?: boolean
}

export function SyncStatus({ showDetails = false }: SyncStatusProps) {
  const { 
    isOnline, 
    pendingCount, 
    isSyncing, 
    lastSyncTime, 
    syncError,
    syncNow 
  } = useSyncStore()

  const lastSyncText = useMemo(() => {
    if (!lastSyncTime) return 'Never'
    const now = Date.now()
    const diff = now - lastSyncTime.getTime()
    if (diff < 60000) return 'Just now'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    return `${Math.floor(diff / 3600000)}h ago`
  }, [lastSyncTime])

  if (!isOnline && pendingCount === 0) return null

  return (
    <div className="flex items-center gap-3">
      <div className="flex items-center gap-2">
        {isSyncing ? (
          <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
        ) : pendingCount > 0 ? (
          <div className="relative">
            <div className="w-5 h-5 bg-amber-100 rounded-full flex items-center justify-center">
              <svg className="w-3 h-3 text-amber-600" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
              </svg>
            </div>
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-amber-500 text-white text-xs rounded-full flex items-center justify-center">
              {pendingCount}
            </span>
          </div>
        ) : (
          <div className="w-5 h-5 bg-green-100 rounded-full flex items-center justify-center">
            <svg className="w-3 h-3 text-green-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        )}
        
        <span className="text-sm text-gray-600">
          {isSyncing ? 'Syncing...' : pendingCount > 0 ? `${pendingCount} pending` : 'Synced'}
        </span>
      </div>

      {showDetails && (
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-400">
            Last: {lastSyncText}
          </span>
          
          {isOnline && pendingCount > 0 && !isSyncing && (
            <button
              onClick={() => syncNow()}
              className="text-xs text-primary hover:underline"
            >
              Sync Now
            </button>
          )}
        </div>
      )}

      {syncError && (
        <span className="text-xs text-red-500">{syncError}</span>
      )}
    </div>
  )
}
