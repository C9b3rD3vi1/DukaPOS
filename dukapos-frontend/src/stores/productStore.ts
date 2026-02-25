import { create } from 'zustand'
import { api } from '@/api/client'
import { dbProducts, dbSyncQueue, db, type DBDProduct } from '@/db/db'
import { useSyncStore } from './syncStore'
import type { Product } from '@/api/types'

interface ProductState {
  products: Product[]
  categories: string[]
  isLoading: boolean
  error: string | null
  searchQuery: string
  selectedCategory: string

  fetchProducts: (shopId: number) => Promise<void>
  fetchCategories: () => Promise<void>
  createProduct: (data: Partial<Product>) => Promise<Product | null>
  updateProduct: (id: number, data: Partial<Product>) => Promise<void>
  deleteProduct: (id: number) => Promise<void>
  getProductById: (id: number) => Product | undefined
  getProductByBarcode: (barcode: string) => Product | undefined
  getLowStockProducts: () => Product[]
  setSearchQuery: (query: string) => void
  setSelectedCategory: (category: string) => void
  syncFromServer: (shopId: number) => Promise<void>
}

export const useProductStore = create<ProductState>((set, get) => ({
  products: [],
  categories: [],
  isLoading: false,
  error: null,
  searchQuery: '',
  selectedCategory: '',

  fetchProducts: async (shopId: number) => {
    set({ isLoading: true, error: null })
    try {
      const { searchQuery, selectedCategory } = get()
      const params = new URLSearchParams()
      params.append('shop_id', shopId.toString())
      if (searchQuery) params.append('search', searchQuery)
      if (selectedCategory) params.append('category', selectedCategory)

      const response = await api.get<{ data: Product[] }>(`/v1/products?${params}`)
      const products = (response.data as unknown as Product[]) || []
      set({ products, isLoading: false })

      // Cache to IndexedDB for offline
      await cacheProducts(products)
    } catch (error) {
      // Try loading from cache
      const cached = await loadCachedProducts()
      if (cached.length > 0) {
        set({ products: cached, isLoading: false })
      } else {
        set({ error: 'Failed to fetch products', isLoading: false })
      }
    }
  },

  fetchCategories: async () => {
    try {
      const response = await api.get<{ data: string[] }>('/v1/products/categories')
      const categories = (response.data as unknown as string[]) || []
      set({ categories })
    } catch (error) {
      console.error('Failed to fetch categories:', error)
    }
  },

  createProduct: async (data: Partial<Product>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    // Save to local IndexedDB first
    const localId = await dbProducts.add({
      serverId: 0,
      name: data.name || '',
      category: data.category || '',
      unit: data.unit || 'pcs',
      costPrice: data.cost_price || 0,
      sellingPrice: data.selling_price || 0,
      currency: data.currency || 'KES',
      currentStock: data.current_stock || 0,
      lowStockThreshold: data.low_stock_threshold || 10,
      barcode: data.barcode,
      imageUrl: data.image_url,
      isActive: true,
      updatedAt: new Date(),
      synced: false
    })

    if (syncStore.isOnline) {
      try {
        const response = await api.post<{ data: Product }>('/v1/products', data)
        const newProduct = response.data as unknown as Product
        
        // Update local with server ID
        const existing = await db.products.get(localId)
        if (existing) {
          await dbProducts.update({ ...existing, serverId: newProduct.id, synced: true })
        }
        
        set(state => ({
          products: [...state.products, newProduct],
          isLoading: false
        }))
        return newProduct
      } catch (error) {
        // Queue for offline sync
        await queueOfflineProduct('create', { ...data, id: localId }, 0)
        set({ error: 'Product saved offline. Will sync when online.', isLoading: false })
        return null
      }
    } else {
      // Offline - queue for later sync
      await queueOfflineProduct('create', data, 0)
      set({ error: 'Product saved offline. Will sync when online.', isLoading: false })
      return null
    }
  },

  updateProduct: async (id: number, data: Partial<Product>) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    // Update local IndexedDB
    const existing = await dbProducts.getByServerId(id)
    if (existing?.id) {
      await dbProducts.update({
        ...existing,
        name: data.name || existing.name,
        category: data.category || existing.category,
        unit: data.unit || existing.unit,
        costPrice: data.cost_price ?? existing.costPrice,
        sellingPrice: data.selling_price ?? existing.sellingPrice,
        currentStock: data.current_stock ?? existing.currentStock,
        lowStockThreshold: data.low_stock_threshold ?? existing.lowStockThreshold,
        barcode: data.barcode ?? existing.barcode,
        updatedAt: new Date(),
        synced: false
      })
    }

    if (syncStore.isOnline) {
      try {
        const response = await api.put<{ data: Product }>(`/v1/products/${id}`, data)
        const updatedProduct = response.data as unknown as Product
        
        // Mark as synced
        if (existing?.id) {
          await dbProducts.update({ ...existing, synced: true })
        }
        
        set(state => ({
          products: state.products.map(p => p.id === id ? updatedProduct : p),
          isLoading: false
        }))
      } catch (error) {
        // Queue for offline sync
        await queueOfflineProduct('update', { ...data, id }, id)
        set({ error: 'Product updated offline. Will sync when online.', isLoading: false })
      }
    } else {
      // Offline - queue for later sync
      await queueOfflineProduct('update', { ...data, id }, id)
      set({ error: 'Product updated offline. Will sync when online.', isLoading: false })
    }
  },

  deleteProduct: async (id: number) => {
    set({ isLoading: true, error: null })
    const syncStore = useSyncStore.getState()
    
    // Remove from local IndexedDB
    const existing = await dbProducts.getByServerId(id)
    if (existing?.id) {
      await dbProducts.delete(existing.id)
    }

    // Optimistic update
    set(state => ({
      products: state.products.filter(p => p.id !== id),
      isLoading: false
    }))

    if (syncStore.isOnline) {
      try {
        await api.delete(`/v1/products/${id}`)
      } catch (error) {
        // Queue for offline sync
        await queueOfflineProduct('delete', { id }, id)
      }
    } else {
      // Offline - queue for later sync
      await queueOfflineProduct('delete', { id }, id)
    }
  },

  getProductById: (id: number) => {
    return get().products.find(p => p.id === id)
  },

  getProductByBarcode: (barcode: string) => {
    return get().products.find(p => p.barcode === barcode)
  },

  getLowStockProducts: () => {
    return get().products.filter(p => p.current_stock <= p.low_stock_threshold)
  },

  setSearchQuery: (query: string) => {
    set({ searchQuery: query })
  },

  setSelectedCategory: (category: string) => {
    set({ selectedCategory: category })
  },

  syncFromServer: async (shopId: number) => {
    await get().fetchProducts(shopId)
  }
}))

async function cacheProducts(products: Product[]): Promise<void> {
  const dbProductsData: DBDProduct[] = products.map(p => ({
    serverId: p.id,
    name: p.name,
    category: p.category || '',
    unit: p.unit,
    costPrice: p.cost_price,
    sellingPrice: p.selling_price,
    currency: p.currency,
    currentStock: p.current_stock,
    lowStockThreshold: p.low_stock_threshold,
    barcode: p.barcode,
    imageUrl: p.image_url,
    isActive: p.is_active,
    updatedAt: new Date(p.updated_at),
    synced: true
  }))

  await dbProducts.clear()
  for (const product of dbProductsData) {
    await dbProducts.add(product)
  }
}

async function loadCachedProducts(): Promise<Product[]> {
  const cached = await dbProducts.getAll()
  return cached.map(p => ({
    id: p.serverId,
    shop_id: 0,
    name: p.name,
    category: p.category,
    unit: p.unit,
    cost_price: p.costPrice,
    selling_price: p.sellingPrice,
    currency: p.currency,
    current_stock: p.currentStock,
    low_stock_threshold: p.lowStockThreshold,
    barcode: p.barcode,
    image_url: p.imageUrl,
    is_active: p.isActive,
    created_at: p.updatedAt.toISOString(),
    updated_at: p.updatedAt.toISOString()
  }))
}

async function queueOfflineProduct(action: 'create' | 'update' | 'delete', data: Partial<Product>, serverId: number): Promise<void> {
  await dbSyncQueue.add({
    type: 'product',
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
