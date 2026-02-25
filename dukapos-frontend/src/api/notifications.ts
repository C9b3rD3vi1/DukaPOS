import { api } from './client'

export interface DeviceRegistration {
  device_token: string
  platform: 'ios' | 'android' | 'web'
}

export const notificationsApi = {
  registerDevice: async (data: DeviceRegistration) => {
    const response = await api.post('/v1/notifications/register-device', data)
    return response.data
  },

  unregisterDevice: async (deviceToken: string) => {
    const response = await api.delete(`/v1/notifications/devices/${deviceToken}`)
    return response.data
  },

  updateNotificationPreferences: async (preferences: {
    low_stock_alerts: boolean
    daily_reports: boolean
    order_updates: boolean
    marketing: boolean
  }) => {
    const response = await api.put('/v1/notifications/preferences', preferences)
    return response.data
  },

  getNotificationPreferences: async () => {
    const response = await api.get('/v1/notifications/preferences')
    return response.data
  }
}
