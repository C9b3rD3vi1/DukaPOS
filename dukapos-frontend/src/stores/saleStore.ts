import { create } from 'zustand'
import { api } from '@/api/client'
import { dbSales, db, type DBSale } from '@/db/db'
import { syncEngine } from '@/db/sync'
import { useSyncStore } from './syncStore'
import type { Sale } from '@/api/types'

interface CartItem {
  productId: number
  productName: string
  quantity: number
  unitPrice: number
  totalPrice: number
}

interface SaleState {
  sales: Sale[]
  currentSale: Sale | null
  cart: CartItem[]
  isLoading: boolean
  isProcessing: boolean
  error: string | null
  paymentMethod: 'cash' | 'mpesa' | 'card' | 'bank'
  mpesaPhone: string

  fetchSales: (shopId: number, limit?: number) => Promise<void>
  fetchSaleById: (id: number) => Promise<Sale | null>
  createSale: (shopId: number, items: CartItem[], paymentMethod: string, mpesaPhone?: string) => Promise<Sale | null>
  addToCart: (productId: number, productName: string, quantity: number, unitPrice: number) => void
  removeFromCart: (productId: number) => void
  updateCartQuantity: (productId: number, quantity: number) => void
  clearCart: () => void
  getCartTotal: () => number
  getCartItemCount: () => number
  setPaymentMethod: (method: 'cash' | 'mpesa' | 'card' | 'bank') => void
  setMpesaPhone: (phone: string) => void
  syncPendingSales: () => Promise<void>
  getRecentSales: (shopId: number) => Promise<Sale[]>
}

export const useSaleStore = create<SaleState>((set, get) => ({
  sales: [],
  currentSale: null,
  cart: [],
  isLoading: false,
  isProcessing: false,
  error: null,
  paymentMethod: 'cash',
  mpesaPhone: '',

  fetchSales: async (shopId: number, limit = 50) => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.get<{ data: Sale[] }>(`/v1/sales?shop_id=${shopId}&limit=${limit}`)
      const sales = (response.data as unknown as Sale[]) || []
      set({ sales, isLoading: false })
      
      // Cache to IndexedDB
      await cacheSalesToDB(sales)
    } catch (error) {
      // Try loading from cache
      const cached = await loadCachedSales()
      if (cached.length > 0) {
        set({ sales: cached, isLoading: false })
      } else {
        set({ error: 'Failed to fetch sales', isLoading: false })
      }
    }
  },

  fetchSaleById: async (id: number) => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.get<{ data: Sale }>(`/v1/sales/${id}`)
      const sale = response.data as unknown as Sale
      set({ currentSale: sale, isLoading: false })
      return sale
    } catch (error) {
      set({ error: 'Failed to fetch sale', isLoading: false })
      return null
    }
  },

  createSale: async (shopId: number, items: CartItem[], paymentMethod: string, mpesaPhone?: string) => {
    set({ isProcessing: true, error: null })
    
    // Save to local DB first (offline-first)
    const localSaleIds: number[] = []
    for (const item of items) {
      const id = await dbSales.add({
        productId: item.productId,
        productName: item.productName,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        totalAmount: item.totalPrice,
        paymentMethod: paymentMethod as 'cash' | 'mpesa' | 'card' | 'bank',
        mpesaReceipt: mpesaPhone,
        createdAt: new Date(),
        synced: false
      })
      localSaleIds.push(id)
    }
    
    // Try to sync immediately if online
    const syncStore = useSyncStore.getState()
    
    if (syncStore.isOnline) {
      try {
        // Create sales for each item on server
        const createdSales: Sale[] = []
        
        for (const item of items) {
          const response = await api.post<{ data: Sale }>('/v1/sales', {
            shop_id: shopId,
            product_id: item.productId,
            quantity: item.quantity,
            unit_price: item.unitPrice,
            payment_method: paymentMethod,
            mpesa_phone: mpesaPhone
          })
          const newSale = response.data as unknown as Sale
          createdSales.push(newSale)
          
          // Mark as synced in local DB
          const localSale = await dbSales.getUnsynced()
          const unsynced = localSale.find(s => s.productId === item.productId && s.quantity === item.quantity && !s.serverId)
          if (unsynced?.id) {
            await dbSales.markSynced(unsynced.id, newSale.id)
          }
        }

        set(state => ({
          sales: [...createdSales, ...state.sales],
          isProcessing: false
        }))

        return createdSales[0]
      } catch (error) {
        // Queue for offline sync
        console.log('Online but failed to sync, queuing for later:', error)
        await queueOfflineSale(shopId, items, paymentMethod, mpesaPhone)
        set({ error: 'Sale saved offline. Will sync when online.', isProcessing: false })
        return null
      }
    } else {
      // Offline - queue for later sync
      await queueOfflineSale(shopId, items, paymentMethod, mpesaPhone)
      set({ error: 'Sale saved offline. Will sync when online.', isProcessing: false })
      return null
    }
  },

  addToCart: (productId: number, productName: string, quantity: number, unitPrice: number) => {
    const { cart } = get()
    const existing = cart.find(item => item.productId === productId)
    
    if (existing) {
      set({
        cart: cart.map(item =>
          item.productId === productId
            ? { ...item, quantity: item.quantity + quantity, totalPrice: (item.quantity + quantity) * item.unitPrice }
            : item
        )
      })
    } else {
      set({
        cart: [...cart, { productId, productName, quantity, unitPrice, totalPrice: quantity * unitPrice }]
      })
    }
  },

  removeFromCart: (productId: number) => {
    set(state => ({
      cart: state.cart.filter(item => item.productId !== productId)
    }))
  },

  updateCartQuantity: (productId: number, quantity: number) => {
    if (quantity <= 0) {
      get().removeFromCart(productId)
      return
    }
    
    set(state => ({
      cart: state.cart.map(item =>
        item.productId === productId
          ? { ...item, quantity, totalPrice: quantity * item.unitPrice }
          : item
      )
    }))
  },

  clearCart: () => {
    set({ cart: [], mpesaPhone: '' })
  },

  getCartTotal: () => {
    return get().cart.reduce((sum, item) => sum + item.totalPrice, 0)
  },

  getCartItemCount: () => {
    return get().cart.reduce((sum, item) => sum + item.quantity, 0)
  },

  setPaymentMethod: (method: 'cash' | 'mpesa' | 'card' | 'bank') => {
    set({ paymentMethod: method })
  },

  setMpesaPhone: (phone: string) => {
    set({ mpesaPhone: phone })
  },

  syncPendingSales: async () => {
    // Use the syncEngine for better retry logic
    const syncStore = useSyncStore.getState()
    await syncStore.syncNow()
  },

  getRecentSales: async (shopId: number) => {
    try {
      const response = await api.get<{ data: Sale[] }>(`/v1/sales?shop_id=${shopId}&limit=20`)
      const sales = (response.data as unknown as Sale[]) || []
      
      // Cache recent sales
      for (const sale of sales) {
        await dbSales.add({
          serverId: sale.id,
          productId: sale.product_id,
          productName: sale.product?.name || '',
          quantity: sale.quantity,
          unitPrice: sale.unit_price,
          totalAmount: sale.total_amount,
          paymentMethod: sale.payment_method,
          mpesaReceipt: sale.mpesa_receipt,
          createdAt: new Date(sale.created_at),
          synced: true
        })
      }
      
      return sales
    } catch {
      const cached = await loadCachedSales()
      return cached.slice(0, 20)
    }
  }
}))

async function cacheSalesToDB(sales: Sale[]): Promise<void> {
  await dbSales.clear()
  for (const sale of sales) {
    await dbSales.add({
      serverId: sale.id,
      productId: sale.product_id,
      productName: sale.product?.name || '',
      quantity: sale.quantity,
      unitPrice: sale.unit_price,
      totalAmount: sale.total_amount,
      paymentMethod: sale.payment_method,
      mpesaReceipt: sale.mpesa_receipt,
      staffId: sale.staff_id,
      notes: sale.notes,
      createdAt: new Date(sale.created_at),
      synced: true
    })
  }
}

async function loadCachedSales(): Promise<Sale[]> {
  const allSales = await db.sales.toArray()
  return allSales.map((s: DBSale) => ({
    id: s.serverId || 0,
    shop_id: 0,
    product_id: s.productId,
    quantity: s.quantity,
    unit_price: s.unitPrice,
    total_amount: s.totalAmount,
    cost_amount: 0,
    profit: 0,
    payment_method: s.paymentMethod,
    mpesa_receipt: s.mpesaReceipt,
    created_at: s.createdAt.toISOString()
  }))
}

async function queueOfflineSale(shopId: number, items: CartItem[], paymentMethod: string, mpesaPhone?: string): Promise<void> {
  // Use syncEngine to queue for sync
  for (const item of items) {
    await syncEngine.queueSaleForSync({
      productId: item.productId,
      productName: item.productName,
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      totalAmount: item.totalPrice,
      paymentMethod: paymentMethod as 'cash' | 'mpesa' | 'card' | 'bank',
      mpesaReceipt: mpesaPhone,
      shopId: shopId
    })
  }
}
