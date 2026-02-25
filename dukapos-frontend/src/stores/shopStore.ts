import { create } from 'zustand'
import { api } from '@/api/client'
import type { Shop } from '@/api/types'

interface ShopState {
  shops: Shop[]
  currentShop: Shop | null
  isLoading: boolean
  error: string | null

  fetchShops: () => Promise<void>
  fetchShopById: (id: number) => Promise<Shop | null>
  createShop: (data: Partial<Shop>) => Promise<Shop | null>
  updateShop: (id: number, data: Partial<Shop>) => Promise<void>
  deleteShop: (id: number) => Promise<void>
  setCurrentShop: (shop: Shop) => void
}

export const useShopStore = create<ShopState>((set) => ({
  shops: [],
  currentShop: null,
  isLoading: false,
  error: null,

  fetchShops: async () => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.get<{ data: Shop[] }>('/v1/shops')
      const shops = (response.data as unknown as Shop[]) || []
      set({ shops, isLoading: false })
    } catch (error) {
      set({ error: 'Failed to fetch shops', isLoading: false })
    }
  },

  fetchShopById: async (id: number) => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.get<{ data: Shop }>(`/v1/shops/${id}`)
      const shop = response.data as unknown as Shop
      set({ currentShop: shop, isLoading: false })
      return shop
    } catch (error) {
      set({ error: 'Failed to fetch shop', isLoading: false })
      return null
    }
  },

  createShop: async (data: Partial<Shop>) => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.post<{ data: Shop }>('/v1/shops', data)
      const newShop = response.data as unknown as Shop
      set(state => ({
        shops: [...state.shops, newShop],
        isLoading: false
      }))
      return newShop
    } catch (error) {
      set({ error: 'Failed to create shop', isLoading: false })
      return null
    }
  },

  updateShop: async (id: number, data: Partial<Shop>) => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.put<{ data: Shop }>(`/v1/shops/${id}`, data)
      const updatedShop = response.data as unknown as Shop
      set(state => ({
        shops: state.shops.map(s => s.id === id ? updatedShop : s),
        currentShop: state.currentShop?.id === id ? updatedShop : state.currentShop,
        isLoading: false
      }))
    } catch (error) {
      set({ error: 'Failed to update shop', isLoading: false })
    }
  },

  deleteShop: async (id: number) => {
    set({ isLoading: true, error: null })
    try {
      await api.delete(`/v1/shops/${id}`)
      set(state => ({
        shops: state.shops.filter(s => s.id !== id),
        currentShop: state.currentShop?.id === id ? null : state.currentShop,
        isLoading: false
      }))
    } catch (error) {
      set({ error: 'Failed to delete shop', isLoading: false })
    }
  },

  setCurrentShop: (shop: Shop) => {
    set({ currentShop: shop })
  }
}))
