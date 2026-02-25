import { api } from './client'
import type { Shop, DashboardData } from './types'

interface UpdateShopData {
  name?: string
  phone?: string
  owner_name?: string
  address?: string
  email?: string
  mpesa_shortcode?: string
}

export const shopApi = {
  getProfile: async (): Promise<Shop> => {
    const response = await api.get<Shop>('/v1/shop/profile')
    return response.data
  },

  updateProfile: async (data: UpdateShopData): Promise<Shop> => {
    const response = await api.put<Shop>('/v1/shop/profile', data)
    return response.data
  },

  getDashboard: async (): Promise<DashboardData> => {
    const response = await api.get<DashboardData>('/v1/shop/dashboard')
    return response.data
  },

  getAccount: async (): Promise<{ id: number; shops: Shop[] }> => {
    const response = await api.get<{ id: number; shops: Shop[] }>('/v1/shop/account')
    return response.data
  }
}
