import { authApi } from './client'
import type { AuthResponse, Account } from './types'

interface LoginData {
  phone: string
  password: string
}

interface RegisterData {
  name: string
  phone: string
  email: string
  password: string
}

interface TwoFactorSetupResponse {
  message: string
  secret: string
  qr_code_url: string
  account_name: string
  issuer: string
}

interface TwoFactorVerifyResponse {
  message: string
  valid: boolean
}

interface TwoFactorBackupCodesResponse {
  backup_codes: string[]
  message: string
}

export const authApiService = {
  login: async (data: LoginData): Promise<AuthResponse> => {
    const response = await authApi.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterData): Promise<AuthResponse> => {
    const response = await authApi.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  sendOTP: async (phone: string): Promise<void> => {
    await authApi.post('/auth/otp/send', { phone })
  },

  verifyOTP: async (phone: string, code: string): Promise<{ verified: boolean }> => {
    const response = await authApi.post('/auth/otp/verify', { phone, code })
    return response.data
  },

  getProfile: async (): Promise<Account> => {
    const response = await authApi.get<Account>('/v1/shop/profile')
    return response.data
  },

  updateProfile: async (data: Partial<Account>): Promise<Account> => {
    const response = await authApi.put<Account>('/v1/shop/profile', data)
    return response.data
  },

  // Two-Factor Authentication
  setupTwoFactor: async (accountId: number, accountName: string, phone: string): Promise<TwoFactorSetupResponse> => {
    const response = await authApi.post<TwoFactorSetupResponse>('/v1/twofactor/setup', {
      account_id: accountId,
      account_name: accountName,
      phone
    })
    return response.data
  },

  verifyTwoFactor: async (token: string, secret: string): Promise<TwoFactorVerifyResponse> => {
    const response = await authApi.post<TwoFactorVerifyResponse>('/v1/twofactor/verify', {
      token,
      secret
    })
    return response.data
  },

  disableTwoFactor: async (accountId: number, token: string, secret: string): Promise<{ message: string }> => {
    const response = await authApi.post<{ message: string }>('/v1/twofactor/disable', {
      account_id: accountId,
      token,
      secret
    })
    return response.data
  },

  generateBackupCodes: async (): Promise<TwoFactorBackupCodesResponse> => {
    const response = await authApi.post<TwoFactorBackupCodesResponse>('/v1/twofactor/backup-codes/generate')
    return response.data
  },

  verifyBackupCode: async (code: string, hash: string[]): Promise<{ message: string }> => {
    const response = await authApi.post<{ message: string }>('/v1/twofactor/backup-codes/verify', {
      code,
      hash
    })
    return response.data
  }
}
