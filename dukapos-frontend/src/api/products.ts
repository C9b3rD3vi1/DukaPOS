import { api } from './client'
import type { Product, PaginatedResponse } from './types'

interface ProductFilters {
  shop_id?: number
  category?: string
  search?: string
  low_stock?: boolean
  page?: number
  limit?: number
}

interface CreateProductData {
  name: string
  category?: string
  unit?: string
  cost_price: number
  selling_price: number
  current_stock: number
  low_stock_threshold?: number
  barcode?: string
  image_url?: string
}

export const productsApi = {
  list: async (filters: ProductFilters = {}): Promise<PaginatedResponse<Product>> => {
    const params = new URLSearchParams()
    
    if (filters.shop_id) params.append('shop_id', filters.shop_id.toString())
    if (filters.category) params.append('category', filters.category)
    if (filters.search) params.append('search', filters.search)
    if (filters.low_stock) params.append('low_stock', 'true')
    if (filters.page) params.append('page', filters.page.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    
    const response = await api.get<PaginatedResponse<Product>>(`/v1/products?${params}`)
    return response.data
  },

  get: async (id: number): Promise<Product> => {
    const response = await api.get<Product>(`/v1/products/${id}`)
    return response.data
  },

  create: async (data: CreateProductData): Promise<Product> => {
    const response = await api.post<Product>('/v1/products', data)
    return response.data
  },

  update: async (id: number, data: Partial<CreateProductData>): Promise<Product> => {
    const response = await api.put<Product>(`/v1/products/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/v1/products/${id}`)
  },

  bulkCreate: async (products: CreateProductData[]): Promise<Product[]> => {
    const response = await api.post<Product[]>('/v1/products/bulk', { products })
    return response.data
  },

  categories: async (): Promise<string[]> => {
    const response = await api.get<string[]>('/v1/products/categories')
    return response.data
  },

  createCategory: async (name: string): Promise<{ id: number }> => {
    const response = await api.post<{ id: number }>('/v1/products/categories', { name })
    return response.data
  },

  deleteCategory: async (id: number): Promise<void> => {
    await api.delete(`/v1/products/categories/${id}`)
  }
}
