import { create } from 'zustand'
import { dbSyncQueue, dbSales, type DBSyncQueue, type DBSale } from '@/db/db'
import { syncEngine, type SyncResult, type ConflictInfo } from '@/db/sync'

interface SyncState {
  isOnline: boolean
  isSyncing: boolean
  pendingCount: number
  lastSyncTime: Date | null
  syncError: string | null
  autoSyncEnabled: boolean
  isInitialized: boolean
  conflicts: ConflictInfo[]

  initialize: () => void
  setOnline: (online: boolean) => void
  addToQueue: (item: Omit<DBSyncQueue, 'id'>) => Promise<void>
  syncNow: () => Promise<SyncResult>
  getPendingCount: () => Promise<void>
  queueSale: (sale: Omit<DBSale, 'id' | 'synced' | 'serverId'>) => Promise<void>
  enableAutoSync: (intervalMs?: number) => void
  disableAutoSync: () => void
  resolveConflict: (type: string, id: number, resolution: 'local' | 'server') => Promise<void>
  clearConflicts: () => void
}

export const useSyncStore = create<SyncState>((set, get) => ({
  isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
  isSyncing: false,
  pendingCount: 0,
  lastSyncTime: null,
  syncError: null,
  autoSyncEnabled: false,
  isInitialized: false,
  conflicts: [],

  initialize: () => {
    if (get().isInitialized) return
    
    if (typeof window !== 'undefined') {
      try {
        window.addEventListener('online', () => {
          useSyncStore.getState().setOnline(true)
          useSyncStore.getState().syncNow()
        })
        
        window.addEventListener('offline', () => {
          useSyncStore.getState().setOnline(false)
        })

        useSyncStore.getState().getPendingCount().catch(console.error)
        useSyncStore.getState().enableAutoSync(30000)
        set({ isInitialized: true })
      } catch (e) {
        console.error('Sync store init error:', e)
        set({ isInitialized: true })
      }
    }
  },

  setOnline: (online) => set({ isOnline: online }),

  addToQueue: async (item) => {
    try {
      await dbSyncQueue.add({
        ...item,
        attempts: 0,
        createdAt: new Date()
      })
      await get().getPendingCount()
    } catch (e) {
      console.error('Add to queue error:', e)
    }
  },

  syncNow: async () => {
    const { isSyncing, isOnline } = get()
    if (isSyncing || !isOnline) {
      return { success: false, synced: 0, failed: 0, errors: ['Sync already in progress or offline'], conflicts: [] }
    }
    
    set({ isSyncing: true, syncError: null })
    
    try {
      const result = await syncEngine.syncAll()
      
      set({ 
        lastSyncTime: new Date(),
        syncError: result.failed > 0 ? `${result.failed} items failed to sync` : null,
        conflicts: result.conflicts
      })
      
      await get().getPendingCount()
      return result
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Sync failed'
      set({ syncError: errorMessage })
      return { success: false, synced: 0, failed: 0, errors: [errorMessage], conflicts: [] }
    } finally {
      set({ isSyncing: false })
    }
  },

  getPendingCount: async () => {
    try {
      const count = await dbSyncQueue.count()
      set({ pendingCount: count })
    } catch (e) {
      console.error('Get pending count error:', e)
      set({ pendingCount: 0 })
    }
  },

  queueSale: async (sale) => {
    try {
      await dbSales.add({
        ...sale,
        synced: false,
        createdAt: new Date()
      })
      await get().getPendingCount()
      
      if (get().isOnline) {
        await get().syncNow()
      }
    } catch (e) {
      console.error('Queue sale error:', e)
    }
  },

  enableAutoSync: (intervalMs = 30000) => {
    try {
      syncEngine.startAutoSync(intervalMs)
      set({ autoSyncEnabled: true })
    } catch (e) {
      console.error('Enable auto sync error:', e)
    }
  },

  disableAutoSync: () => {
    try {
      syncEngine.stopAutoSync()
      set({ autoSyncEnabled: false })
    } catch (e) {
      console.error('Disable auto sync error:', e)
    }
  },

  resolveConflict: async (type, id, resolution) => {
    try {
      await syncEngine.resolveConflictManual(type, id, resolution)
      const conflicts = get().conflicts.filter(c => !(c.id === id && c.type === type))
      set({ conflicts })
      await get().syncNow()
    } catch (e) {
      console.error('Resolve conflict error:', e)
    }
  },

  clearConflicts: () => {
    set({ conflicts: [] })
  }
}))
