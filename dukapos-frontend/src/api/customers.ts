import { api } from './client'
import type { Customer, PaginatedResponse } from './types'

interface CustomerFilters {
  shop_id?: number
  search?: string
  page?: number
  limit?: number
}

interface CreateCustomerData {
  name: string
  phone: string
  email?: string
}

export const customersApi = {
  list: async (filters: CustomerFilters = {}): Promise<PaginatedResponse<Customer>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.search) params.append('search', filters.search)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Customer>>(`/v1/customers?${params}`)
    return response.data
  },

  get: async (id: number): Promise<Customer> => {
    const response = await api.get<Customer>(`/v1/customers/${id}`)
    return response.data
  },

  create: async (data: CreateCustomerData): Promise<Customer> => {
    const response = await api.post<Customer>('/v1/customers', data)
    return response.data
  },

  update: async (id: number, data: Partial<CreateCustomerData>): Promise<Customer> => {
    const response = await api.put<Customer>(`/v1/customers/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/v1/customers/${id}`)
  }
}
