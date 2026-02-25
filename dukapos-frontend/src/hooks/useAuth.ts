import { useCallback } from 'react'
import { useAuthStore } from '@/stores/authStore'

interface LoginCredentials {
  phone: string
  password: string
}

interface RegisterData {
  name: string
  phone: string
  email: string
  password: string
}

export function useAuth() {
  const { 
    user, 
    shop, 
    token, 
    isAuthenticated, 
    isAdmin,
    isLoading,
    login: storeLogin,
    register: storeRegister,
    logout: storeLogout,
    setShop: storeSetShop,
    initialize
  } = useAuthStore()

  const login = useCallback(async (credentials: LoginCredentials) => {
    try {
      await storeLogin(credentials.phone, credentials.password)
      return { success: true }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Login failed'
      return { success: false, error: message }
    }
  }, [storeLogin])

  const register = useCallback(async (data: RegisterData) => {
    try {
      await storeRegister(data.name, data.phone, data.email, data.password)
      return { success: true }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Registration failed'
      return { success: false, error: message }
    }
  }, [storeRegister])

  const logout = useCallback(async () => {
    try {
      await storeLogout()
      return { success: true }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Logout failed'
      return { success: false, error: message }
    }
  }, [storeLogout])

  const setShop = useCallback((shop: Parameters<typeof storeSetShop>[0]) => {
    storeSetShop(shop)
  }, [storeSetShop])

  const hasPermission = useCallback((permission: string): boolean => {
    if (!user) return false
    
    // Admin has all permissions
    if (user.is_admin) return true
    
    // Check role-based permissions
    const rolePermissions: Record<string, string[]> = {
      manager: ['products', 'sales', 'customers', 'reports', 'staff', 'settings'],
      staff: ['products', 'sales', 'customers'],
      viewer: ['products', 'sales']
    }
    
    const userRole = (user as { role?: string }).role || 'staff'
    return rolePermissions[userRole]?.includes(permission) || false
  }, [user])

  const isAuthenticatedAndReady = useCallback(() => {
    return isAuthenticated && !isLoading
  }, [isAuthenticated, isLoading])

  return {
    user,
    shop,
    token,
    isAuthenticated,
    isAdmin,
    isLoading,
    login,
    register,
    logout,
    setShop,
    initialize,
    hasPermission,
    isAuthenticatedAndReady
  }
}
