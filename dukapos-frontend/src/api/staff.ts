import { api } from './client'
import type { Staff, PaginatedResponse } from './types'

interface StaffFilters {
  shop_id?: number
  role?: string
  is_active?: boolean
  page?: number
  limit?: number
}

interface CreateStaffData {
  name: string
  phone: string
  role: string
  pin?: string
}

interface UpdateStaffData {
  name?: string
  phone?: string
  role?: string
  is_active?: boolean
}

export const staffApi = {
  list: async (filters: StaffFilters = {}): Promise<PaginatedResponse<Staff>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.role) params.append('role', filters.role)
    if (filters.is_active !== undefined) params.append('is_active', filters.is_active.toString())
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Staff>>(`/v1/staff?${params}`)
    return response.data
  },

  get: async (id: number): Promise<Staff> => {
    const response = await api.get<Staff>(`/v1/staff/${id}`)
    return response.data
  },

  create: async (data: CreateStaffData): Promise<Staff> => {
    const response = await api.post<Staff>('/v1/staff', data)
    return response.data
  },

  update: async (id: number, data: UpdateStaffData): Promise<Staff> => {
    const response = await api.put<Staff>(`/v1/staff/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/v1/staff/${id}`)
  },

  updatePin: async (id: number, pin: string): Promise<{ success: boolean }> => {
    const response = await api.put<{ success: boolean }>(`/v1/staff/${id}/pin`, { pin })
    return response.data
  }
}
