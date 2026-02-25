import { create } from 'zustand'
import { api } from '@/api/client'
import { dbSuppliers, dbSyncQueue, db, type DBSupplier } from '@/db/db'
import { useSyncStore } from './syncStore'
import type { Supplier } from '@/api/types'

interface SupplierState {
  suppliers: Supplier[]
  isLoading: boolean
  error: string | null
  searchQuery: string

  fetchSuppliers: (shopId: number) => Promise<void>
  createSupplier: (data: Partial<Supplier>) => Promise<Supplier | null>
  updateSupplier: (id: number, data: Partial<Supplier>) => Promise<void>
  deleteSupplier: (id: number) => Promise<void>
  getSupplierById: (id: number) => Supplier | undefined
  setSearchQuery: (query: string) => void
}

export const useSupplierStore = create<SupplierState>((set, get) => ({
  suppliers: [],
  isLoading: false,
  error: null,
  searchQuery: '',

  fetchSuppliers: async (shopId: number) => {
    set({ isLoading: true, error: null })
    try {
      const { searchQuery } = get()
      const params = new URLSearchParams()
      params.append('shop_id', shopId.toString())
      if (searchQuery) params.append('search', searchQuery)

      const response = await api.get<{ data: Supplier[] }>(`/v1/suppliers?${params}`)
      const suppliers = (response.data as unknown as Supplier[]) || []
      set({ suppliers, isLoading: false })

      await cacheSuppliers(suppliers)
    } catch (error) {
      const cached = await loadCachedSuppliers()
      if (cached.length > 0) {
        set({ suppliers: cached, isLoading: false })
      } else {
        set({ error: 'Failed to fetch suppliers', isLoading: false })
      }
    }
  },

  createSupplier: async (data: Partial<Supplier>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const localId = await dbSuppliers.add({
      serverId: 0,
      name: data.name || '',
      phone: data.phone,
      email: data.email,
      address: data.address,
      updatedAt: new Date(),
      synced: false
    })

    if (syncStore.isOnline) {
      try {
        const response = await api.post<{ data: Supplier }>('/v1/suppliers', data)
        const newSupplier = response.data as unknown as Supplier
        
        const existing = await db.suppliers.get(localId)
        if (existing?.id) {
          await dbSuppliers.update({ ...existing, serverId: newSupplier.id, synced: true })
        }
        
        set(state => ({
          suppliers: [...state.suppliers, newSupplier],
          isLoading: false
        }))
        return newSupplier
      } catch (error) {
        await queueOfflineSupplier('create', { ...data, id: localId }, 0)
        set({ error: 'Supplier saved offline. Will sync when online.', isLoading: false })
        return null
      }
    } else {
      await queueOfflineSupplier('create', data, 0)
      set({ error: 'Supplier saved offline. Will sync when online.', isLoading: false })
      return null
    }
  },

  updateSupplier: async (id: number, data: Partial<Supplier>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const existing = await dbSuppliers.getByServerId(id)
    if (existing?.id) {
      await dbSuppliers.update({
        ...existing,
        name: data.name || existing.name,
        phone: data.phone ?? existing.phone,
        email: data.email ?? existing.email,
        address: data.address ?? existing.address,
        updatedAt: new Date(),
        synced: false
      })
    }

    if (syncStore.isOnline) {
      try {
        const response = await api.put<{ data: Supplier }>(`/v1/suppliers/${id}`, data)
        const updatedSupplier = response.data as unknown as Supplier
        
        if (existing?.id) {
          await dbSuppliers.update({ ...existing, synced: true })
        }
        
        set(state => ({
          suppliers: state.suppliers.map(s => s.id === id ? updatedSupplier : s),
          isLoading: false
        }))
      } catch (error) {
        await queueOfflineSupplier('update', { ...data, id }, id)
        set({ error: 'Supplier updated offline. Will sync when online.', isLoading: false })
      }
    } else {
      await queueOfflineSupplier('update', { ...data, id }, id)
      set({ error: 'Supplier updated offline. Will sync when online.', isLoading: false })
    }
  },

  deleteSupplier: async (id: number) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const existing = await dbSuppliers.getByServerId(id)
    if (existing?.id) {
      await dbSuppliers.delete(existing.id)
    }

    set(state => ({
      suppliers: state.suppliers.filter(s => s.id !== id),
      isLoading: false
    }))

    if (syncStore.isOnline) {
      try {
        await api.delete(`/v1/suppliers/${id}`)
      } catch (error) {
        await queueOfflineSupplier('delete', { id }, id)
      }
    } else {
      await queueOfflineSupplier('delete', { id }, id)
    }
  },

  getSupplierById: (id: number) => {
    return get().suppliers.find(s => s.id === id)
  },

  setSearchQuery: (query: string) => {
    set({ searchQuery: query })
  }
}))

async function cacheSuppliers(suppliers: Supplier[]): Promise<void> {
  const dbSuppliersData: DBSupplier[] = suppliers.map(s => ({
    serverId: s.id,
    name: s.name,
    phone: s.phone,
    email: s.email,
    address: s.address,
    updatedAt: new Date(s.created_at),
    synced: true
  }))

  await dbSuppliers.clear()
  for (const supplier of dbSuppliersData) {
    await dbSuppliers.add(supplier)
  }
}

async function loadCachedSuppliers(): Promise<Supplier[]> {
  const cached = await dbSuppliers.getAll()
  return cached.map(s => ({
    id: s.serverId,
    shop_id: 0,
    name: s.name,
    phone: s.phone,
    email: s.email,
    address: s.address,
    created_at: s.updatedAt.toISOString()
  }))
}

async function queueOfflineSupplier(action: 'create' | 'update' | 'delete', data: Partial<Supplier>, serverId: number): Promise<void> {
  await dbSyncQueue.add({
    type: 'supplier',
    action,
    data: {
      ...data,
      serverId,
      updatedAt: new Date().toISOString()
    },
    attempts: 0,
    createdAt: new Date()
  })
}
