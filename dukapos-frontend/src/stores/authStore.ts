import { create } from 'zustand'
import { Preferences } from '@capacitor/preferences'
import { Capacitor } from '@capacitor/core'
import { authApiService } from '@/api/auth'
import type { Account, Shop } from '@/api/types'

const isNative = typeof window !== 'undefined' && typeof Capacitor !== 'undefined' && Capacitor.isNativePlatform()

const storage = {
  get: async (key: string): Promise<string | null> => {
    if (isNative) {
      try {
        const { value } = await Preferences.get({ key })
        return value
      } catch {
        return null
      }
    }
    if (typeof window !== 'undefined') {
      return localStorage.getItem(key)
    }
    return null
  },
  set: async (key: string, value: string): Promise<void> => {
    if (isNative) {
      try {
        await Preferences.set({ key, value })
      } catch {
        // Ignore
      }
    } else if (typeof window !== 'undefined') {
      localStorage.setItem(key, value)
    }
  },
  remove: async (key: string): Promise<void> => {
    if (isNative) {
      try {
        await Preferences.remove({ key })
      } catch {
        // Ignore
      }
    } else if (typeof window !== 'undefined') {
      localStorage.removeItem(key)
    }
  }
}

interface AuthState {
  user: Account | null
  shop: Shop | null
  token: string | null
  isLoading: boolean
  isAuthenticated: boolean
  isAdmin: boolean
  
  initialize: () => Promise<void>
  login: (phone: string, password: string) => Promise<void>
  register: (name: string, phone: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setShop: (shop: Shop) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  shop: null,
  token: null,
  isLoading: true,
  isAuthenticated: false,
  isAdmin: false,

  initialize: async () => {
    try {
      const token = await storage.get('auth_token')
      const userJson = await storage.get('user_data')
      const shopJson = await storage.get('shop_data')
      
      if (token && userJson) {
        const user = JSON.parse(userJson)
        const shop = shopJson ? JSON.parse(shopJson) : null
        
        set({
          token,
          user,
          shop,
          isAuthenticated: true,
          isAdmin: user.is_admin || false,
          isLoading: false
        })
      } else {
        set({ isLoading: false })
      }
    } catch {
      set({ isLoading: false })
    }
  },

  login: async (phone: string, password: string) => {
    const response = await authApiService.login({ phone, password })
    
    const user = response.account || response.user
    if (!user) {
      throw new Error('Invalid response: no user data')
    }
    
    await storage.set('auth_token', response.token)
    await storage.set('user_data', JSON.stringify(user))
    
    if (response.shop) {
      await storage.set('shop_data', JSON.stringify(response.shop))
    }
    
    set({
      token: response.token,
      user,
      shop: response.shop || null,
      isAuthenticated: true,
      isAdmin: user.is_admin || false
    })
  },

  register: async (name: string, phone: string, email: string, password: string) => {
    const response = await authApiService.register({ name, phone, email, password })
    
    const user = response.account || response.user
    if (!user) {
      throw new Error('Invalid response: no user data')
    }
    
    await storage.set('auth_token', response.token)
    await storage.set('user_data', JSON.stringify(user))
    
    if (response.shop) {
      await storage.set('shop_data', JSON.stringify(response.shop))
    }
    
    set({
      token: response.token,
      user,
      shop: response.shop || null,
      isAuthenticated: true,
      isAdmin: user.is_admin || false
    })
  },

  logout: async () => {
    await storage.remove('auth_token')
    await storage.remove('user_data')
    await storage.remove('shop_data')
    
    set({
      token: null,
      user: null,
      shop: null,
      isAuthenticated: false,
      isAdmin: false
    })
  },

  setShop: (shop: Shop) => {
    storage.set('shop_data', JSON.stringify(shop))
    set({ shop })
  }
}))
