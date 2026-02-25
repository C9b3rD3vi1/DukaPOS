import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { Preferences } from '@capacitor/preferences'
import { Capacitor } from '@capacitor/core'

const isNative = typeof Capacitor !== 'undefined' && Capacitor.isNativePlatform()

export interface ApiError {
  message: string
  code?: string
  status?: number
}

const getToken = async (): Promise<string | null> => {
  if (isNative) {
    try {
      const { value } = await Preferences.get({ key: 'auth_token' })
      return value
    } catch {
      return null
    }
  }
  return localStorage.getItem('auth_token')
}

const removeToken = async (): Promise<void> => {
  if (isNative) {
    try {
      await Preferences.remove({ key: 'auth_token' })
    } catch {
      // Ignore
    }
  } else {
    localStorage.removeItem('auth_token')
  }
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 
  (isNative
    ? 'https://api.dukapos.com' 
    : 'http://localhost:8080')

export function extractApiData<T>(response: unknown): T {
  if (!response) return [] as T
  const res = response as { data?: unknown }
  if (res.data !== undefined) {
    if (Array.isArray(res.data)) return res.data as T
    const inner = res.data as { data?: unknown }
    if (inner.data !== undefined) return inner.data as T
    return res.data as T
  }
  return response as T
}

export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{ message?: string; error?: string }>
    
    if (axiosError.response) {
      // Server responded with error
      const data = axiosError.response.data
      if (data?.message) return data.message
      if (data?.error) return data.error
      
      // Status-based messages
      switch (axiosError.response.status) {
        case 400: return 'Invalid request. Please check your input.'
        case 401: return 'Session expired. Please login again.'
        case 403: return 'You do not have permission to perform this action.'
        case 404: return 'Resource not found.'
        case 422: return 'Validation error. Please check your input.'
        case 429: return 'Too many requests. Please wait a moment.'
        case 500: return 'Server error. Please try again later.'
        case 502: return 'Service unavailable. Please try again later.'
        case 503: return 'Service temporarily unavailable.'
        default: return `Request failed (${axiosError.response.status})`
      }
    } else if (axiosError.request) {
      // Request made but no response
      return 'Network error. Please check your connection.'
    }
  }
  
  // Non-axios errors
  if (error instanceof Error) {
    return error.message
  }
  
  return 'An unexpected error occurred.'
}

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

export const authApi = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

export const adminApi = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

let errorHandler: ((message: string, type: 'error' | 'warning') => void) | null = null

export function setErrorHandler(handler: (message: string, type: 'error' | 'warning') => void) {
  errorHandler = handler
}

export function showError(message: string) {
  if (errorHandler) {
    errorHandler(message, 'error')
  } else {
    console.error(message)
  }
}

export function showWarning(message: string) {
  if (errorHandler) {
    errorHandler(message, 'warning')
  } else {
    console.warn(message)
  }
}

api.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  try {
    const token = await getToken()
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  } catch {
    // Storage not available, continue without token
  }
  
  return config
})

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const message = getErrorMessage(error)
    
    if (error.response?.status === 401) {
      console.warn('API returned 401 - redirecting to login')
      window.location.href = '/login'
    } else {
      showError(message)
    }
    
    return Promise.reject(error)
  }
)

authApi.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  try {
    const token = await getToken()
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  } catch {
    // Storage not available, continue without token
  }
  
  return config
})

authApi.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const message = getErrorMessage(error)
    
    if (error.response?.status === 401) {
      try {
        await removeToken()
      } catch {
        // Ignore errors
      }
      window.location.href = '/login'
    } else {
      showError(message)
    }
    
    return Promise.reject(error)
  }
)

adminApi.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  try {
    const token = await getToken()
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  } catch {
    // Storage not available, continue without token
  }
  
  return config
})

adminApi.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const message = getErrorMessage(error)
    
    if (error.response?.status === 401) {
      try {
        await removeToken()
      } catch {
        // Ignore errors
      }
      window.location.href = '/admin/login'
    } else {
      showError(message)
    }
    
    return Promise.reject(error)
  }
)

export default api
