import { api } from './client'
import type { Supplier, Order, PaginatedResponse } from './types'

interface SupplierFilters {
  shop_id?: number
  search?: string
  page?: number
  limit?: number
}

interface CreateSupplierData {
  name: string
  phone?: string
  email?: string
  address?: string
}

interface OrderFilters {
  supplier_id?: number
  status?: string
  page?: number
  limit?: number
}

interface CreateOrderData {
  supplier_id: number
  products: Array<{
    product_id: number
    quantity: number
    unit_price: number
  }>
  notes?: string
}

export const suppliersApi = {
  list: async (filters: SupplierFilters = {}): Promise<PaginatedResponse<Supplier>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.search) params.append('search', filters.search)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Supplier>>(`/v1/suppliers?${params}`)
    return response.data
  },

  get: async (id: number): Promise<Supplier> => {
    const response = await api.get<Supplier>(`/v1/suppliers/${id}`)
    return response.data
  },

  create: async (data: CreateSupplierData): Promise<Supplier> => {
    const response = await api.post<Supplier>('/v1/suppliers', data)
    return response.data
  },

  update: async (id: number, data: Partial<CreateSupplierData>): Promise<Supplier> => {
    const response = await api.put<Supplier>(`/v1/suppliers/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/v1/suppliers/${id}`)
  },

  listOrders: async (filters: OrderFilters = {}): Promise<PaginatedResponse<Order>> => {
    const params = new URLSearchParams()
    
    if (filters.supplier_id) params.append('supplier_id', filters.supplier_id.toString())
    if (filters.status) params.append('status', filters.status)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Order>>(`/v1/orders?${params}`)
    return response.data
  },

  createOrder: async (data: CreateOrderData): Promise<Order> => {
    const response = await api.post<Order>('/v1/orders', data)
    return response.data
  },

  getOrder: async (id: number): Promise<Order> => {
    const response = await api.get<Order>(`/v1/orders/${id}`)
    return response.data
  },

  updateOrderStatus: async (id: number, status: string): Promise<Order> => {
    const response = await api.put<Order>(`/v1/orders/${id}/status`, { status })
    return response.data
  },

  deleteOrder: async (id: number): Promise<void> => {
    await api.delete(`/v1/orders/${id}`)
  }
}
