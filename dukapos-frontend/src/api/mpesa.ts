import { api } from './client'
import type { MpesaPayment, PaginatedResponse } from './types'

interface MpesaFilters {
  shop_id?: number
  status?: string
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
}

interface STKPushData {
  phone: string
  amount: number
  account_ref?: string
  description?: string
}

interface STKPushResponse {
  checkout_request_id: string
  response_code: string
  response_description: string
}

export const mpesaApi = {
  stkPush: async (data: STKPushData): Promise<STKPushResponse> => {
    const response = await api.post<STKPushResponse>('/v1/mpesa/stk-push', data)
    return response.data
  },

  getStatus: async (checkoutId: string): Promise<{ status: string; amount?: number; mpesa_receipt?: string }> => {
    const response = await api.get<{ status: string; amount?: number; mpesa_receipt?: string }>(`/v1/mpesa/status/${checkoutId}`)
    return response.data
  },

  listPayments: async (filters: MpesaFilters = {}): Promise<PaginatedResponse<MpesaPayment>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.status) params.append('status', filters.status)
    if (filters.start_date) params.append('start_date', filters.start_date)
    if (filters.end_date) params.append('end_date', filters.end_date)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<MpesaPayment>>(`/v1/mpesa/payments?${params}`)
    return response.data
  },

  retryPayment: async (id: number): Promise<{ success: boolean }> => {
    const response = await api.post<{ success: boolean }>(`/v1/mpesa/payments/${id}/retry`)
    return response.data
  },

  getTransactions: async (filters: MpesaFilters = {}): Promise<{ transactions: unknown[]; total: number }> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.start_date) params.append('start_date', filters.start_date)
    if (filters.end_date) params.append('end_date', filters.end_date)
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<{ transactions: unknown[]; total: number }>(`/v1/mpesa/transactions?${params}`)
    return response.data
  },

  getBalance: async (): Promise<{ balance: number; currency: string }> => {
    const response = await api.get<{ balance: number; currency: string }>('/v1/mpesa/balance')
    return response.data
  },

  b2cSend: async (phone: string, amount: number, remarks: string): Promise<{ success: boolean; conversation_id?: string }> => {
    const response = await api.post<{ success: boolean; conversation_id?: string }>('/v1/mpesa/b2c', { phone, amount, remarks })
    return response.data
  }
}
