import { create } from 'zustand'
import { api } from '@/api/client'
import { dbOrders, dbSyncQueue, db, type DBOrder, type DBOrderItem } from '@/db/db'
import { useSyncStore } from './syncStore'
import type { Order } from '@/api/types'

interface OrderState {
  orders: Order[]
  isLoading: boolean
  error: string | null
  filterStatus: string

  fetchOrders: (shopId: number) => Promise<void>
  createOrder: (data: Partial<Order>) => Promise<Order | null>
  updateOrderStatus: (id: number, status: string) => Promise<void>
  deleteOrder: (id: number) => Promise<void>
  getOrderById: (id: number) => Order | undefined
  setFilterStatus: (status: string) => void
}

export const useOrderStore = create<OrderState>((set, get) => ({
  orders: [],
  isLoading: false,
  error: null,
  filterStatus: '',

  fetchOrders: async (shopId: number) => {
    set({ isLoading: true, error: null })
    try {
      const { filterStatus } = get()
      const params = new URLSearchParams()
      params.append('shop_id', shopId.toString())
      if (filterStatus) params.append('status', filterStatus)

      const response = await api.get<{ data: Order[] }>(`/v1/orders?${params}`)
      const orders = (response.data as unknown as Order[]) || []
      set({ orders, isLoading: false })

      await cacheOrders(orders)
    } catch (error) {
      const cached = await loadCachedOrders()
      if (cached.length > 0) {
        set({ orders: cached, isLoading: false })
      } else {
        set({ error: 'Failed to fetch orders', isLoading: false })
      }
    }
  },

  createOrder: async (data: Partial<Order>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const dbOrderItems: DBOrderItem[] = (data.items || []).map(item => ({
      productId: item.product_id,
      productName: item.product?.name || '',
      quantity: item.quantity,
      price: item.price,
      total: item.total
    }))

    const localId = await dbOrders.add({
      serverId: 0,
      supplierId: data.supplier_id || 0,
      supplierName: data.supplier?.name || '',
      status: 'pending',
      totalAmount: data.total_amount || 0,
      notes: data.notes,
      items: dbOrderItems,
      createdAt: new Date(),
      updatedAt: new Date(),
      synced: false
    })

    if (syncStore.isOnline) {
      try {
        const response = await api.post<{ data: Order }>('/v1/orders', data)
        const newOrder = response.data as unknown as Order
        
        const existing = await db.orders.get(localId)
        if (existing?.id) {
          await dbOrders.update({ ...existing, serverId: newOrder.id, synced: true })
        }
        
        set(state => ({
          orders: [...state.orders, newOrder],
          isLoading: false
        }))
        return newOrder
      } catch (error) {
        await queueOfflineOrder('create', { ...data, id: localId }, 0)
        set({ error: 'Order saved offline. Will sync when online.', isLoading: false })
        return null
      }
    } else {
      await queueOfflineOrder('create', data, 0)
      set({ error: 'Order saved offline. Will sync when online.', isLoading: false })
      return null
    }
  },

  updateOrderStatus: async (id: number, status: string) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const existing = await dbOrders.getByServerId(id)
    if (existing?.id) {
      await dbOrders.update({
        ...existing,
        status: status as 'pending' | 'approved' | 'received' | 'cancelled',
        updatedAt: new Date(),
        synced: false
      })
    }

    if (syncStore.isOnline) {
      try {
        const response = await api.put<{ data: Order }>(`/v1/orders/${id}`, { status })
        const updatedOrder = response.data as unknown as Order
        
        if (existing?.id) {
          await dbOrders.update({ ...existing, synced: true })
        }
        
        set(state => ({
          orders: state.orders.map(o => o.id === id ? updatedOrder : o),
          isLoading: false
        }))
      } catch (error) {
        await queueOfflineOrder('update', { id, status }, id)
        set({ error: 'Order updated offline. Will sync when online.', isLoading: false })
      }
    } else {
      await queueOfflineOrder('update', { id, status }, id)
      set({ error: 'Order updated offline. Will sync when online.', isLoading: false })
    }
  },

  deleteOrder: async (id: number) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    const existing = await dbOrders.getByServerId(id)
    if (existing?.id) {
      await dbOrders.delete(existing.id)
    }

    set(state => ({
      orders: state.orders.filter(o => o.id !== id),
      isLoading: false
    }))

    if (syncStore.isOnline) {
      try {
        await api.delete(`/v1/orders/${id}`)
      } catch (error) {
        await queueOfflineOrder('delete', { id }, id)
      }
    } else {
      await queueOfflineOrder('delete', { id }, id)
    }
  },

  getOrderById: (id: number) => {
    return get().orders.find(o => o.id === id)
  },

  setFilterStatus: (status: string) => {
    set({ filterStatus: status })
  }
}))

async function cacheOrders(orders: Order[]): Promise<void> {
  const dbOrdersData: DBOrder[] = orders.map(o => ({
    serverId: o.id,
    supplierId: o.supplier_id,
    supplierName: o.supplier?.name || '',
    status: o.status as 'pending' | 'approved' | 'received' | 'cancelled',
    totalAmount: o.total_amount,
    notes: o.notes,
    items: (o.items || []).map(item => ({
      productId: item.product_id,
      productName: item.product?.name || '',
      quantity: item.quantity,
      price: item.price,
      total: item.total
    })),
    createdAt: new Date(o.created_at),
    updatedAt: new Date(o.updated_at),
    synced: true
  }))

  await dbOrders.clear()
  for (const order of dbOrdersData) {
    await dbOrders.add(order)
  }
}

async function loadCachedOrders(): Promise<Order[]> {
  const cached = await dbOrders.getAll()
  return cached.map(o => ({
    id: o.serverId,
    shop_id: 0,
    supplier_id: o.supplierId,
    supplier: { id: o.supplierId, name: o.supplierName, shop_id: 0, created_at: '' },
    status: o.status,
    total_amount: o.totalAmount,
    notes: o.notes,
    items: o.items.map(item => ({
      product_id: item.productId,
      product: { id: item.productId, name: item.productName, shop_id: 0, unit: '', cost_price: 0, selling_price: 0, currency: '', current_stock: 0, low_stock_threshold: 0, is_active: true, created_at: '', updated_at: '' },
      quantity: item.quantity,
      price: item.price,
      total: item.total
    })),
    created_at: o.createdAt.toISOString(),
    updated_at: o.updatedAt.toISOString()
  }))
}

async function queueOfflineOrder(action: 'create' | 'update' | 'delete', data: Partial<Order>, serverId: number): Promise<void> {
  await dbSyncQueue.add({
    type: 'order',
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
