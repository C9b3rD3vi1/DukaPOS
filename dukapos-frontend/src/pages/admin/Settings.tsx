import { useState } from 'react'
import { adminApi } from '@/api/client'

interface Settings {
  site_name: string
  site_url: string
  support_email: string
  mpesa_shortcode: string
  mpesa_passkey: string
  mpesa_secret: string
  sms_enabled: boolean
  email_enabled: boolean
}

export default function AdminSettings() {
  const [settings, setSettings] = useState<Settings>({
    site_name: 'DukaPOS',
    site_url: 'https://dukapos.com',
    support_email: 'support@dukapos.com',
    mpesa_shortcode: '',
    mpesa_passkey: '',
    mpesa_secret: '',
    sms_enabled: true,
    email_enabled: true
  })
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await adminApi.put('/admin/settings', settings)
      setMessage('Settings saved successfully!')
      setTimeout(() => setMessage(''), 3000)
    } catch (err) { console.error(err) }
    finally { setSaving(false) }
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500 mt-1">System configuration</p>
      </div>

      {message && (
        <div className="p-4 bg-green-50 text-green-600 rounded-xl mb-6">
          {message}
        </div>
      )}

      <form onSubmit={handleSave} className="space-y-6">
        {/* General Settings */}
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
          <h2 className="font-semibold text-gray-900 mb-4">General</h2>
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Site Name</label>
              <input
                type="text"
                value={settings.site_name}
                onChange={(e) => setSettings({ ...settings, site_name: e.target.value })}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Site URL</label>
              <input
                type="url"
                value={settings.site_url}
                onChange={(e) => setSettings({ ...settings, site_url: e.target.value })}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Support Email</label>
              <input
                type="email"
                value={settings.support_email}
                onChange={(e) => setSettings({ ...settings, support_email: e.target.value })}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* M-Pesa Settings */}
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
          <h2 className="font-semibold text-gray-900 mb-4">M-Pesa Configuration</h2>
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Shortcode</label>
              <input
                type="text"
                value={settings.mpesa_shortcode}
                onChange={(e) => setSettings({ ...settings, mpesa_shortcode: e.target.value })}
                placeholder="174379"
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Passkey</label>
              <input
                type="password"
                value={settings.mpesa_passkey}
                onChange={(e) => setSettings({ ...settings, mpesa_passkey: e.target.value })}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Consumer Secret</label>
              <input
                type="password"
                value={settings.mpesa_secret}
                onChange={(e) => setSettings({ ...settings, mpesa_secret: e.target.value })}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* Integrations */}
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
          <h2 className="font-semibold text-gray-900 mb-4">Integrations</h2>
          <div className="space-y-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={settings.sms_enabled}
                onChange={(e) => setSettings({ ...settings, sms_enabled: e.target.checked })}
                className="w-5 h-5 text-primary rounded"
              />
              <div>
                <p className="font-medium text-gray-900">SMS (Africa Talking)</p>
                <p className="text-sm text-gray-500">Enable SMS notifications</p>
              </div>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={settings.email_enabled}
                onChange={(e) => setSettings({ ...settings, email_enabled: e.target.checked })}
                className="w-5 h-5 text-primary rounded"
              />
              <div>
                <p className="font-medium text-gray-900">Email (SendGrid)</p>
                <p className="text-sm text-gray-500">Enable email notifications</p>
              </div>
            </label>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="px-6 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </form>
    </div>
  )
}
