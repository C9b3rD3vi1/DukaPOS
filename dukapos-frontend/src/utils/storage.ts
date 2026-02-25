import { Capacitor } from '@capacitor/core'
import { Preferences } from '@capacitor/preferences'

const isNative = typeof Capacitor !== 'undefined' && Capacitor.isNativePlatform()

class Storage {
  private prefix = 'dukapos_'

  private getKey(key: string): string {
    return `${this.prefix}${key}`
  }

  async get(key: string): Promise<unknown> {
    try {
      if (isNative) {
        const { value } = await Preferences.get({ key: this.getKey(key) })
        if (value === null) return null
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      } else {
        const value = localStorage.getItem(this.getKey(key))
        if (value === null) return null
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      }
    } catch (error) {
      console.error('Storage get error:', error)
      return null
    }
  }

  async set(key: string, value: unknown): Promise<void> {
    try {
      const serialized = typeof value === 'string' ? value : JSON.stringify(value)
      if (isNative) {
        await Preferences.set({ key: this.getKey(key), value: serialized })
      } else {
        localStorage.setItem(this.getKey(key), serialized)
      }
    } catch (error) {
      console.error('Storage set error:', error)
    }
  }

  async remove(key: string): Promise<void> {
    try {
      if (isNative) {
        await Preferences.remove({ key: this.getKey(key) })
      } else {
        localStorage.removeItem(this.getKey(key))
      }
    } catch (error) {
      console.error('Storage remove error:', error)
    }
  }

  async clear(): Promise<void> {
    try {
      if (isNative) {
        const { keys } = await Preferences.keys()
        for (const key of keys) {
          if (key.startsWith(this.prefix)) {
            await Preferences.remove({ key })
          }
        }
      } else {
        const keys = Object.keys(localStorage)
        for (const key of keys) {
          if (key.startsWith(this.prefix)) {
            localStorage.removeItem(key)
          }
        }
      }
    } catch (error) {
      console.error('Storage clear error:', error)
    }
  }
}

export const storage = new Storage()

export const storageKeys = {
  authToken: 'auth_token',
  userData: 'user_data',
  shopData: 'shop_data',
  lastSync: 'last_sync',
  settings: 'settings',
  theme: 'theme',
  language: 'language',
  onboardingComplete: 'onboarding_complete'
} as const

export async function getAuthToken(): Promise<string | null> {
  return storage.get(storageKeys.authToken) as Promise<string | null>
}

export async function setAuthToken(token: string): Promise<void> {
  return storage.set(storageKeys.authToken, token)
}

export async function removeAuthToken(): Promise<void> {
  return storage.remove(storageKeys.authToken)
}

export async function getUserData<T = unknown>(): Promise<T | null> {
  return storage.get(storageKeys.userData) as Promise<T | null>
}

export async function setUserData<T = unknown>(data: T): Promise<void> {
  return storage.set(storageKeys.userData, data)
}

export async function getShopData<T = unknown>(): Promise<T | null> {
  return storage.get(storageKeys.shopData) as Promise<T | null>
}

export async function setShopData<T = unknown>(data: T): Promise<void> {
  return storage.set(storageKeys.shopData, data)
}

export async function getSettings<T = unknown>(): Promise<T | null> {
  return storage.get(storageKeys.settings) as Promise<T | null>
}

export async function setSettings<T = unknown>(settings: T): Promise<void> {
  return storage.set(storageKeys.settings, settings)
}

export async function clearAllData(): Promise<void> {
  return storage.clear()
}
