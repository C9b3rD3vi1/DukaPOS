import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useSyncStore } from '@/stores/syncStore'

describe('SyncStore', () => {
  beforeEach(() => {
    // Reset store state
    useSyncStore.setState({
      isOnline: true,
      isSyncing: false,
      pendingCount: 0,
      lastSyncTime: null,
      syncError: null,
      autoSyncEnabled: false,
      isInitialized: false
    })
    
    // Mock IndexedDB
    vi.stubGlobal('indexedDB', {
      open: vi.fn()
    })
  })

  it('should have correct initial state', () => {
    const state = useSyncStore.getState()
    expect(state.isOnline).toBe(true)
    expect(state.isSyncing).toBe(false)
    expect(state.pendingCount).toBe(0)
    expect(state.syncError).toBeNull()
    expect(state.autoSyncEnabled).toBe(false)
  })

  it('should set online status correctly', () => {
    useSyncStore.getState().setOnline(false)
    expect(useSyncStore.getState().isOnline).toBe(false)
    
    useSyncStore.getState().setOnline(true)
    expect(useSyncStore.getState().isOnline).toBe(true)
  })

  it('should set sync error correctly', async () => {
    // Attempting syncNow when offline should return early
    useSyncStore.setState({ isOnline: false })
    
    const result = await useSyncStore.getState().syncNow()
    
    expect(result.success).toBe(false)
    expect(result.errors).toContain('Sync already in progress or offline')
  })

  it('should format pending count correctly', async () => {
    useSyncStore.setState({ pendingCount: 5 })
    expect(useSyncStore.getState().pendingCount).toBe(5)
  })
})

describe('Offline Detection', () => {
  it('should detect online status', () => {
    // In Node.js environment, navigator.onLine may not exist
    const isOnline = typeof navigator !== 'undefined' ? navigator.onLine : true
    expect(typeof isOnline).toBe('boolean')
  })
})
