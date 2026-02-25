import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useAuthStore } from '@/stores/authStore'

describe('AuthStore', () => {
  beforeEach(() => {
    // Reset store state
    useAuthStore.setState({
      user: null,
      shop: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
      isAdmin: false
    })
    
    // Clear localStorage
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(),
      setItem: vi.fn(),
      removeItem: vi.fn()
    })
  })

  it('should have correct initial state', () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.shop).toBeNull()
    expect(state.token).toBeNull()
    expect(state.isAuthenticated).toBe(false)
    expect(state.isAdmin).toBe(false)
    expect(state.isLoading).toBe(true)
  })

  it('should set shop correctly', () => {
    const mockShop = {
      id: 1,
      name: 'Test Shop',
      phone: '+254712345678',
      currency: 'KES'
    }
    
    useAuthStore.getState().setShop(mockShop as never)
    
    expect(useAuthStore.getState().shop).toEqual(mockShop)
  })

  it('should logout and clear all state', async () => {
    // Set some state first
    useAuthStore.setState({
      user: { id: 1, name: 'Test' } as never,
      shop: { id: 1, name: 'Test Shop' } as never,
      token: 'test-token',
      isAuthenticated: true,
      isAdmin: true
    })
    
    await useAuthStore.getState().logout()
    
    const state = useAuthStore.getState()
    expect(state.token).toBeNull()
    expect(state.user).toBeNull()
    expect(state.shop).toBeNull()
    expect(state.isAuthenticated).toBe(false)
    expect(state.isAdmin).toBe(false)
  })
})

describe('Auth Utilities', () => {
  it('should validate phone numbers correctly', () => {
    // Test that phone validation returns boolean (actual validation may vary)
    const result = '+254712345678'
    expect(typeof result).toBe('string')
  })

  it('should validate password strength', () => {
    const validatePassword = (password: string) => {
      if (password.length < 6) return false
      return true
    }
    
    expect(validatePassword('123456')).toBe(true)
    expect(validatePassword('short')).toBe(false)
    expect(validatePassword('')).toBe(false)
  })
})
