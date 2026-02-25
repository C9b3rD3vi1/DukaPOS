import { api } from './client'
import type { Sale, PaginatedResponse, SalesReport } from './types'

interface SaleFilters {
  shop_id?: number
  product_id?: number
  customer_id?: number
  payment_method?: string
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
}

interface CreateSaleData {
  product_id: number
  quantity: number
  unit_price: number
  payment_method: 'cash' | 'mpesa' | 'card' | 'bank'
  mpesa_phone?: string
  staff_id?: number
  notes?: string
}

export const salesApi = {
  list: async (filters: SaleFilters = {}): Promise<PaginatedResponse<Sale>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.product_id) params.append('product_id', filters.product_id.toString())
    if (filters.customer_id) params.append('customer_id', filters.customer_id.toString())
    if (filters.payment_method) params.append('payment_method', filters.payment_method)
    if (filters.start_date) params.append('start_date', filters.start_date)
    if (filters.end_date) params.append('end_date', filters.end_date)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Sale>>(`/v1/sales?${params}`)
    return response.data
  },

  get: async (id: number): Promise<Sale> => {
    const response = await api.get<Sale>(`/v1/sales/${id}`)
    return response.data
  },

  create: async (data: CreateSaleData): Promise<Sale> => {
    const response = await api.post<Sale>('/v1/sales', data)
    return response.data
  },

  getReport: async (shopId: number, period: 'daily' | 'weekly' | 'monthly'): Promise<SalesReport> => {
    const response = await api.get<SalesReport>(`/v1/reports/${shopId}?period=${period}`)
    return response.data
  },

  exportSales: async (shopId: number, format: 'csv' | 'pdf' | 'excel', startDate?: string, endDate?: string): Promise<Blob> => {
    const params = new URLSearchParams()
    params.append('shop_id', shopId.toString())
    params.append('format', format)
    if (startDate) params.append('start_date', startDate)
    if (endDate) params.append('end_date', endDate)
    
    const response = await api.get<Blob>(`/v1/export/sales?${params}`, {
      responseType: 'blob'
    })
    return response.data
  }
}
