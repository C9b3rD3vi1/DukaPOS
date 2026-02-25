import { useState, useEffect } from 'react'
import { api, authApi } from '@/api/client'
import { authApiService } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'
import { useTranslation } from '@/utils/i18n'
import type { NotificationPreferences } from '@/api/types'

export default function Settings() {
  const user = useAuthStore((state) => state.user)
  const shop = useAuthStore((state) => state.shop)
  const setShop = useAuthStore((state) => state.setShop)
  const { isOnline, pendingCount, syncNow, isSyncing } = useSyncStore()
  const { locale, setLocale, availableLocales, t } = useTranslation()
  const [formData, setFormData] = useState({ name: '', phone: '', address: '', email: '' })
  const [isSaving, setIsSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [notificationsEnabled, setNotificationsEnabled] = useState(false)
  const [exportLoading, setExportLoading] = useState('')
  const [notificationPrefs, setNotificationPrefs] = useState<NotificationPreferences>({
    low_stock_alerts: true,
    daily_reports: true,
    order_updates: true,
    marketing: false
  })
  const [currency, setCurrency] = useState('KES')
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false)
  const [twoFactorSetup, setTwoFactorSetup] = useState(false)
  const [twoFactorSecret, setTwoFactorSecret] = useState('')
  const [twoFactorQRCode, setTwoFactorQRCode] = useState('')
  const [twoFactorToken, setTwoFactorToken] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [twoFactorLoading, setTwoFactorLoading] = useState(false)

  useEffect(() => {
    if (shop) {
      setFormData({ name: shop.name, phone: shop.phone, address: shop.address || '', email: shop.email || '' })
      setCurrency((shop as { currency?: string }).currency || 'KES')
    }
    checkNotificationPermission()
    loadNotificationPreferences()
    checkTwoFactorStatus()
  }, [shop])

  const loadNotificationPreferences = async () => {
    try {
      const response = await api.get('/v1/notifications/preferences')
      if (response.data) {
        setNotificationPrefs(response.data)
      }
    } catch (e) {
      console.error('Failed to load preferences:', e)
    }
  }

  const saveNotificationPreferences = async (key: keyof NotificationPreferences, value: boolean) => {
    const newPrefs = { ...notificationPrefs, [key]: value }
    setNotificationPrefs(newPrefs)
    try {
      await api.put('/v1/notifications/preferences', newPrefs)
      setMessage('Preferences saved!')
    } catch (e) {
      console.error('Failed to save preferences:', e)
    }
    setTimeout(() => setMessage(''), 3000)
  }

  const checkNotificationPermission = async () => {
    if ('Notification' in window) {
      setNotificationsEnabled(Notification.permission === 'granted')
    }
  }

  const checkTwoFactorStatus = async () => {
    try {
      const response = await authApi.get('/v1/auth/2fa/status')
      setTwoFactorEnabled(response.data.enabled || false)
    } catch (e) {
      console.log('2FA status check failed - may not be enabled')
    }
  }

  const setupTwoFactor = async () => {
    if (!user?.id || !shop?.name) return
    setTwoFactorLoading(true)
    try {
      const response = await authApiService.setupTwoFactor(user.id, shop.name, shop.phone || '')
      setTwoFactorSecret(response.secret)
      setTwoFactorQRCode(response.qr_code_url)
      setTwoFactorSetup(true)
    } catch (e) {
      setMessage('Failed to setup 2FA')
    } finally {
      setTwoFactorLoading(false)
    }
  }

  const verifyAndEnableTwoFactor = async () => {
    if (!twoFactorToken || !twoFactorSecret) {
      setMessage('Please enter the verification code')
      return
    }
    setTwoFactorLoading(true)
    try {
      const response = await authApiService.verifyTwoFactor(twoFactorToken, twoFactorSecret)
      if (response.valid) {
        const backupResponse = await authApiService.generateBackupCodes()
        setBackupCodes(backupResponse.backup_codes)
        setTwoFactorEnabled(true)
        setTwoFactorSetup(false)
        setMessage('2FA enabled successfully! Save your backup codes.')
      } else {
        setMessage('Invalid verification code')
      }
    } catch (e) {
      setMessage('Failed to verify 2FA')
    } finally {
      setTwoFactorLoading(false)
    }
  }

  const disableTwoFactor = async () => {
    if (!twoFactorToken || !twoFactorSecret) {
      setMessage('Please enter your verification code to disable 2FA')
      return
    }
    setTwoFactorLoading(true)
    try {
      await authApiService.disableTwoFactor(user?.id || 0, twoFactorToken, twoFactorSecret)
      setTwoFactorEnabled(false)
      setTwoFactorSetup(false)
      setTwoFactorSecret('')
      setTwoFactorQRCode('')
      setTwoFactorToken('')
      setBackupCodes([])
      setMessage('2FA disabled successfully')
    } catch (e) {
      setMessage('Failed to disable 2FA - invalid token')
    } finally {
      setTwoFactorLoading(false)
    }
  }

  const requestNotifications = async () => {
    if (!('Notification' in window)) {
      setMessage('Notifications not supported')
      return
    }
    
    const permission = await Notification.requestPermission()
    if (permission === 'granted') {
      setNotificationsEnabled(true)
      setMessage('Notifications enabled!')
      
      // Subscribe to push notifications
      if ('serviceWorker' in navigator && 'PushManager' in window) {
        try {
          const registration = await navigator.serviceWorker.ready
          const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array('BEl62iUYgUivxIkv69yViEuiBIa-Ib9-SkvMeAtA3LFgDzkrxZJjSgSnfckjBJuBkr3qBUYIHBQFLXYp5Nksh8U')
          })
          
          // Send subscription to backend
          await api.post('/v1/push/subscribe', subscription)
        } catch (e) {
          console.error('Push subscription failed:', e)
        }
      }
    } else {
      setMessage('Notifications denied')
    }
    setTimeout(() => setMessage(''), 3000)
  }

  const urlBase64ToUint8Array = (base64String: string) => {
    const padding = '='.repeat((4 - base64String.length % 4) % 4)
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
    const rawData = window.atob(base64)
    const outputArray = new Uint8Array(rawData.length)
    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i)
    }
    return outputArray
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSaving(true)
    try {
      const response = await api.put('/v1/shop/profile', formData)
      setShop(response.data)
      setMessage('Settings saved!')
      setTimeout(() => setMessage(''), 3000)
    } catch (err) {
      setMessage('Failed to save')
    } finally {
      setIsSaving(false)
    }
  }

  const handleExport = async (type: string, format: string) => {
    if (!shop?.id) return
    setExportLoading(type)
    try {
      const response = await api.get(`/v1/export/${type}?format=${format}&shop_id=${shop.id}`, {
        responseType: 'blob'
      })
      
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `${type}_${new Date().toISOString().split('T')[0]}.${format}`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      setMessage(`${type} exported successfully!`)
    } catch (err) {
      setMessage('Export failed')
    } finally {
      setExportLoading('')
      setTimeout(() => setMessage(''), 3000)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500 mt-1">Manage your account and preferences</p>
      </div>

      {message && (
        <div className={`p-4 rounded-xl mb-6 ${message.includes('Failed') ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'}`}>
          {message}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Shop Settings */}
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h3 className="font-semibold mb-4">Shop Information</h3>
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Shop Name</label>
              <input type="text" value={formData.name} onChange={(e) => setFormData({...formData, name: e.target.value})} className="w-full px-4 py-3 border rounded-xl" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
              <input type="tel" value={formData.phone} onChange={(e) => setFormData({...formData, phone: e.target.value})} className="w-full px-4 py-3 border rounded-xl" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input type="email" value={formData.email} onChange={(e) => setFormData({...formData, email: e.target.value})} className="w-full px-4 py-3 border rounded-xl" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Address</label>
              <textarea value={formData.address} onChange={(e) => setFormData({...formData, address: e.target.value})} className="w-full px-4 py-3 border rounded-xl" rows={2} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Currency</label>
              <select 
                value={currency} 
                onChange={(e) => setCurrency(e.target.value)}
                className="w-full px-4 py-3 border rounded-xl"
              >
                <option value="KES">KES - Kenyan Shilling</option>
                <option value="USD">USD - US Dollar</option>
                <option value="TZS">TZS - Tanzanian Shilling</option>
                <option value="UGX">UGX - Ugandan Shilling</option>
              </select>
            </div>
            <button type="submit" disabled={isSaving} className="w-full py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50">
              {isSaving ? 'Saving...' : 'Save Changes'}
            </button>
          </form>
        </div>

        {/* Right Column */}
        <div className="space-y-6">
          {/* Language */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">{t('settings.language') || 'Language'}</h3>
            <div className="flex gap-2">
              {availableLocales.map((loc) => (
                <button
                  key={loc.code}
                  onClick={async () => {
                    setLocale(loc.code)
                    try {
                      await api.put('/v1/users/settings', { language: loc.code })
                    } catch (err) {
                      console.error('Failed to save language preference')
                    }
                  }}
                  className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition ${
                    locale === loc.code
                      ? 'bg-primary text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {loc.name}
                </button>
              ))}
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Language preference is saved to your account and synced across devices.
            </p>
          </div>

          {/* Notifications */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">Notifications</h3>
            
            {/* Push Notification Toggle */}
            <div className="flex items-center justify-between mb-6 pb-6 border-b border-gray-100">
              <div>
                <p className="font-medium">Push Notifications</p>
                <p className="text-sm text-gray-500">Enable device notifications</p>
              </div>
              {notificationsEnabled ? (
                <span className="px-3 py-1 bg-green-100 text-green-700 rounded-lg text-sm font-medium">Enabled</span>
              ) : (
                <button onClick={requestNotifications} className="px-4 py-2 bg-primary text-white rounded-lg text-sm font-medium">
                  Enable
                </button>
              )}
            </div>
            
            {/* Notification Preferences */}
            <div className="space-y-4">
              <p className="font-medium text-sm text-gray-500 mb-3">Notification Types</p>
              
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Low Stock Alerts</p>
                  <p className="text-sm text-gray-500">Get notified when products are running low</p>
                </div>
                <button
                  onClick={() => saveNotificationPreferences('low_stock_alerts', !notificationPrefs.low_stock_alerts)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    notificationPrefs.low_stock_alerts ? 'bg-primary' : 'bg-gray-200'
                  }`}
                >
                  <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    notificationPrefs.low_stock_alerts ? 'translate-x-6' : 'translate-x-1'
                  }`} />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Daily Reports</p>
                  <p className="text-sm text-gray-500">Receive daily sales summary</p>
                </div>
                <button
                  onClick={() => saveNotificationPreferences('daily_reports', !notificationPrefs.daily_reports)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    notificationPrefs.daily_reports ? 'bg-primary' : 'bg-gray-200'
                  }`}
                >
                  <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    notificationPrefs.daily_reports ? 'translate-x-6' : 'translate-x-1'
                  }`} />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Order Updates</p>
                  <p className="text-sm text-gray-500">Supplier order status notifications</p>
                </div>
                <button
                  onClick={() => saveNotificationPreferences('order_updates', !notificationPrefs.order_updates)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    notificationPrefs.order_updates ? 'bg-primary' : 'bg-gray-200'
                  }`}
                >
                  <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    notificationPrefs.order_updates ? 'translate-x-6' : 'translate-x-1'
                  }`} />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Marketing</p>
                  <p className="text-sm text-gray-500">Promotions and product updates</p>
                </div>
                <button
                  onClick={() => saveNotificationPreferences('marketing', !notificationPrefs.marketing)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    notificationPrefs.marketing ? 'bg-primary' : 'bg-gray-200'
                  }`}
                >
                  <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    notificationPrefs.marketing ? 'translate-x-6' : 'translate-x-1'
                  }`} />
                </button>
              </div>
            </div>
          </div>

          {/* Sync Status */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">Sync Status</h3>
            <div className="space-y-3 mb-4">
              <div className="flex justify-between"><span className="text-gray-500">Status</span><span className={`font-medium ${isOnline ? 'text-green-600' : 'text-amber-600'}`}>{isOnline ? 'Online' : 'Offline'}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Pending</span><span className="font-medium">{pendingCount} items</span></div>
            </div>
            <button onClick={() => syncNow()} disabled={!isOnline || isSyncing || pendingCount === 0} className="w-full py-3 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 disabled:opacity-50">
              {isSyncing ? 'Syncing...' : 'Sync Now'}
            </button>
          </div>

          {/* Account */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">Account</h3>
            <div className="space-y-3">
              <div className="flex justify-between"><span className="text-gray-500">Plan</span><span className="font-medium capitalize">{user?.plan || 'free'}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Email</span><span className="font-medium">{user?.email}</span></div>
              <div className="flex justify-between"><span className="text-gray-500">Phone</span><span className="font-medium">{user?.phone}</span></div>
            </div>
          </div>

          {/* Two-Factor Authentication */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">Two-Factor Authentication</h3>
            
            {!twoFactorSetup && !twoFactorEnabled && (
              <div className="text-center">
                <div className="mb-4 p-4 bg-gray-50 rounded-lg">
                  <p className="text-gray-600 text-sm">Add an extra layer of security to your account by enabling 2FA</p>
                </div>
                <button 
                  onClick={setupTwoFactor} 
                  disabled={twoFactorLoading}
                  className="w-full py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50"
                >
                  {twoFactorLoading ? 'Setting up...' : 'Enable 2FA'}
                </button>
              </div>
            )}

            {twoFactorSetup && (
              <div className="space-y-4">
                <div className="text-center">
                  <p className="text-sm text-gray-600 mb-3">Scan this QR code with your authenticator app</p>
                  {twoFactorQRCode && (
                    <div className="inline-block p-2 bg-white border rounded-lg mb-3">
                      <img src={twoFactorQRCode} alt="2FA QR Code" className="w-40 h-40" />
                    </div>
                  )}
                  <p className="text-xs text-gray-500 mb-2">Or enter this secret manually:</p>
                  <code className="text-xs bg-gray-100 px-2 py-1 rounded break-all">{twoFactorSecret}</code>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Verification Code</label>
                  <input 
                    type="text" 
                    value={twoFactorToken}
                    onChange={(e) => setTwoFactorToken(e.target.value)}
                    placeholder="Enter 6-digit code"
                    className="w-full px-4 py-3 border rounded-xl"
                    maxLength={6}
                  />
                </div>
                
                <div className="flex gap-2">
                  <button 
                    onClick={verifyAndEnableTwoFactor}
                    disabled={twoFactorLoading || twoFactorToken.length < 6}
                    className="flex-1 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark disabled:opacity-50"
                  >
                    {twoFactorLoading ? 'Verifying...' : 'Verify & Enable'}
                  </button>
                  <button 
                    onClick={() => { setTwoFactorSetup(false); setTwoFactorSecret(''); setTwoFactorQRCode(''); }}
                    className="flex-1 py-3 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}

            {twoFactorEnabled && backupCodes.length > 0 && (
              <div className="space-y-4">
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                  <p className="text-green-700 font-medium text-sm">2FA is enabled!</p>
                  <p className="text-green-600 text-xs mt-1">Save these backup codes in a safe place. You can use them to access your account if you lose your phone.</p>
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  {backupCodes.map((code, index) => (
                    <code key={index} className="text-xs bg-gray-100 px-2 py-1 rounded text-center font-mono">
                      {code}
                    </code>
                  ))}
                </div>
                
                <button 
                  onClick={disableTwoFactor}
                  disabled={twoFactorLoading}
                  className="w-full py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 disabled:opacity-50"
                >
                  {twoFactorLoading ? 'Disabling...' : 'Disable 2FA'}
                </button>
              </div>
            )}

            {twoFactorEnabled && backupCodes.length === 0 && !twoFactorSetup && (
              <div className="space-y-4">
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                  <p className="text-green-700 font-medium text-sm">Two-Factor Authentication is enabled</p>
                </div>
                
                <button 
                  onClick={() => {
                    authApiService.generateBackupCodes().then(r => setBackupCodes(r.backup_codes))
                  }}
                  className="w-full py-3 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200"
                >
                  View Backup Codes
                </button>
                
                <button 
                  onClick={disableTwoFactor}
                  disabled={twoFactorLoading}
                  className="w-full py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 disabled:opacity-50"
                >
                  {twoFactorLoading ? 'Disabling...' : 'Disable 2FA'}
                </button>
              </div>
            )}
          </div>

          {/* Export */}
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h3 className="font-semibold mb-4">Export Data</h3>
            <div className="grid grid-cols-2 gap-3">
              {['products', 'sales', 'report'].map((type) => (
                <div key={type} className="space-y-2">
                  <p className="text-sm font-medium capitalize">{type}</p>
                  <div className="flex gap-2">
                    {['csv', 'pdf'].map((format) => (
                      <button
                        key={format}
                        onClick={() => handleExport(type, format)}
                        disabled={exportLoading === type}
                        className="flex-1 px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition disabled:opacity-50"
                      >
                        {format.toUpperCase()}
                      </button>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
