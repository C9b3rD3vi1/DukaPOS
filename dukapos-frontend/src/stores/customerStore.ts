import { create } from 'zustand'
import { api } from '@/api/client'
import { dbCustomers, dbSyncQueue, type DBCustomer } from '@/db/db'
import { useSyncStore } from './syncStore'
import type { Customer } from '@/api/types'

interface CustomerState {
  customers: Customer[]
  isLoading: boolean
  error: string | null
  searchQuery: string

  fetchCustomers: (shopId: number) => Promise<void>
  createCustomer: (data: Partial<Customer>) => Promise<Customer | null>
  updateCustomer: (id: number, data: Partial<Customer>) => Promise<void>
  deleteCustomer: (id: number) => Promise<void>
  getCustomerById: (id: number) => Customer | undefined
  getCustomerByPhone: (phone: string) => Customer | undefined
  setSearchQuery: (query: string) => void
  addLoyaltyPoints: (id: number, points: number) => Promise<void>
  deductLoyaltyPoints: (id: number, points: number) => Promise<void>
}

export const useCustomerStore = create<CustomerState>((set, get) => ({
  customers: [],
  isLoading: false,
  error: null,
  searchQuery: '',

  fetchCustomers: async (shopId: number) => {
    set({ isLoading: true, error: null })
    try {
      const { searchQuery } = get()
      const params = new URLSearchParams()
      params.append('shop_id', shopId.toString())
      if (searchQuery) params.append('search', searchQuery)

      const response = await api.get<{ data: Customer[] }>(`/v1/customers?${params}`)
      const customers = (response.data as unknown as Customer[]) || []
      set({ customers, isLoading: false })

      // Cache to IndexedDB for offline
      await cacheCustomers(customers)
    } catch (error) {
      // Try loading from cache
      const cached = await loadCachedCustomers()
      if (cached.length > 0) {
        set({ customers: cached, isLoading: false })
      } else {
        set({ error: 'Failed to fetch customers', isLoading: false })
      }
    }
  },

  createCustomer: async (data: Partial<Customer>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    // Save to local IndexedDB first
    const localId = await dbCustomers.add({
      serverId: 0,
      name: data.name || '',
      phone: data.phone || '',
      email: data.email,
      loyaltyPoints: data.loyalty_points || 0,
      totalPurchases: data.total_purchases || 0,
      updatedAt: new Date(),
      synced: false
    })

    if (syncStore.isOnline) {
      try {
        const response = await api.post<{ data: Customer }>('/v1/customers', data)
        const newCustomer = response.data as unknown as Customer
        
        // Update local with server ID
        const existing = await dbCustomers.get(localId)
        if (existing?.id) {
          await dbCustomers.update({ ...existing, serverId: newCustomer.id, synced: true })
        }
        
        set(state => ({
          customers: [...state.customers, newCustomer],
          isLoading: false
        }))
        return newCustomer
      } catch (error) {
        // Queue for offline sync
        await queueOfflineCustomer('create', { ...data, id: localId }, 0)
        set({ error: 'Customer saved offline. Will sync when online.', isLoading: false })
        return null
      }
    } else {
      // Offline - queue for later sync
      await queueOfflineCustomer('create', data, 0)
      set({ error: 'Customer saved offline. Will sync when online.', isLoading: false })
      return null
    }
  },

  updateCustomer: async (id: number, data: Partial<Customer>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    // Update local IndexedDB
    const existing = await dbCustomers.getByServerId(id)
    if (existing?.id) {
      await dbCustomers.update({
        ...existing,
        name: data.name || existing.name,
        phone: data.phone || existing.phone,
        email: data.email ?? existing.email,
        loyaltyPoints: data.loyalty_points ?? existing.loyaltyPoints,
        totalPurchases: data.total_purchases ?? existing.totalPurchases,
        updatedAt: new Date(),
        synced: false
      })
    }

    if (syncStore.isOnline) {
      try {
        const response = await api.put<{ data: Customer }>(`/v1/customers/${id}`, data)
        const updatedCustomer = response.data as unknown as Customer
        
        if (existing?.id) {
          await dbCustomers.update({ ...existing, synced: true })
        }
        
        set(state => ({
          customers: state.customers.map(c => c.id === id ? updatedCustomer : c),
          isLoading: false
        }))
      } catch (error) {
        await queueOfflineCustomer('update', { ...data, id }, id)
        set({ error: 'Customer updated offline. Will sync when online.', isLoading: false })
      }
    } else {
      await queueOfflineCustomer('update', { ...data, id }, id)
      set({ error: 'Customer updated offline. Will sync when online.', isLoading: false })
    }
  },

  deleteCustomer: async (id: number) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const existing = await dbCustomers.getByServerId(id)
    if (existing?.id) {
      await dbCustomers.delete(existing.id)
    }

    set(state => ({
      customers: state.customers.filter(c => c.id !== id),
      isLoading: false
    }))

    if (syncStore.isOnline) {
      try {
        await api.delete(`/v1/customers/${id}`)
      } catch (error) {
        await queueOfflineCustomer('delete', { id }, id)
      }
    } else {
      await queueOfflineCustomer('delete', { id }, id)
    }
  },

  getCustomerById: (id: number) => {
    return get().customers.find(c => c.id === id)
  },

  getCustomerByPhone: (phone: string) => {
    return get().customers.find(c => c.phone === phone)
  },

  setSearchQuery: (query: string) => {
    set({ searchQuery: query })
  },

  addLoyaltyPoints: async (id: number, points: number) => {
    const customer = get().getCustomerById(id)
    if (customer) {
      const newPoints = (customer.loyalty_points || 0) + points
      await get().updateCustomer(id, { loyalty_points: newPoints })
    }
  },

  deductLoyaltyPoints: async (id: number, points: number) => {
    const customer = get().getCustomerById(id)
    if (customer) {
      const newPoints = Math.max(0, (customer.loyalty_points || 0) - points)
      await get().updateCustomer(id, { loyalty_points: newPoints })
    }
  }
}))

async function cacheCustomers(customers: Customer[]): Promise<void> {
  const dbCustomersData: DBCustomer[] = customers.map(c => ({
    serverId: c.id,
    name: c.name,
    phone: c.phone,
    email: c.email,
    loyaltyPoints: c.loyalty_points,
    totalPurchases: c.total_purchases,
    updatedAt: new Date((c as { updated_at?: string }).updated_at || c.created_at),
    synced: true
  }))

  await dbCustomers.clear()
  for (const customer of dbCustomersData) {
    await dbCustomers.add(customer)
  }
}

async function loadCachedCustomers(): Promise<Customer[]> {
  const cached = await dbCustomers.getAll()
  return cached.map((c: DBCustomer) => ({
    id: c.serverId,
    shop_id: 0,
    name: c.name,
    phone: c.phone,
    email: c.email,
    loyalty_points: c.loyaltyPoints,
    total_purchases: c.totalPurchases,
    created_at: c.updatedAt.toISOString()
  }))
}

async function queueOfflineCustomer(action: 'create' | 'update' | 'delete', data: Partial<Customer>, serverId: number): Promise<void> {
  await dbSyncQueue.add({
    type: 'customer',
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
