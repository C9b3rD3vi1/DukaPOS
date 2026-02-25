import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/Input'

export interface WhiteLabelConfig {
  shop_name: string
  shop_logo?: string
  shop_favicon?: string
  primary_color: string
  secondary_color: string
  accent_color: string
  background_color: string
  text_color: string
  custom_css?: string
  custom_js?: string
  custom_domain?: string
  footer_text?: string
  show_branding: boolean
}

const defaultConfig: WhiteLabelConfig = {
  shop_name: '',
  primary_color: '#0D9488',
  secondary_color: '#1E293B',
  accent_color: '#F97316',
  background_color: '#FFFFFF',
  text_color: '#1E293B',
  show_branding: true
}

const colorPresets = [
  { name: 'Teal', primary: '#0D9488', secondary: '#1E293B', accent: '#F97316' },
  { name: 'Blue', primary: '#2563EB', secondary: '#1E293B', accent: '#F97316' },
  { name: 'Green', primary: '#16A34A', secondary: '#1E293B', accent: '#F97316' },
  { name: 'Purple', primary: '#7C3AED', secondary: '#1E293B', accent: '#F97316' },
  { name: 'Red', primary: '#DC2626', secondary: '#1E293B', accent: '#F97316' },
  { name: 'Rose', primary: '#E11D48', secondary: '#1E293B', accent: '#F97316' },
]

export default function WhiteLabel() {
  const shop = useAuthStore((state) => state.shop)
  const [config, setConfig] = useState<WhiteLabelConfig>(defaultConfig)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [message, setMessage] = useState('')
  const [activeTab, setActiveTab] = useState<'colors' | 'branding' | 'advanced'>('colors')
  const [previewMode, setPreviewMode] = useState(false)

  useEffect(() => {
    fetchConfig()
  }, [shop?.id])

  const fetchConfig = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get(`/v1/shop/whitelabel/${shop.id}`)
      if (response.data?.data) {
        setConfig({ ...defaultConfig, ...response.data.data })
      }
    } catch (e) {
      console.error('Failed to load white label config:', e)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSave = async () => {
    if (!shop?.id) return
    setIsSaving(true)
    try {
      await api.put(`/v1/shop/whitelabel/${shop.id}`, config)
      setMessage('Settings saved successfully!')
      setTimeout(() => setMessage(''), 3000)
    } catch (e) {
      setMessage('Failed to save settings')
    } finally {
      setIsSaving(false)
    }
  }

  const handleColorChange = (key: keyof WhiteLabelConfig, value: string | boolean) => {
    setConfig(prev => ({ ...prev, [key]: value }))
  }

  const applyPreset = (preset: typeof colorPresets[0]) => {
    setConfig(prev => ({
      ...prev,
      primary_color: preset.primary,
      secondary_color: preset.secondary,
      accent_color: preset.accent
    }))
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 bg-gray-200 animate-pulse rounded"></div>
        <div className="grid grid-cols-2 gap-4">
          <div className="h-32 bg-gray-200 animate-pulse rounded-xl"></div>
          <div className="h-32 bg-gray-200 animate-pulse rounded-xl"></div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">White Label</h1>
        <p className="text-gray-500 mt-1">Customize your shop's appearance and branding</p>
      </div>

      {message && (
        <div className={`p-4 rounded-xl mb-6 ${message.includes('Failed') ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'}`}>
          {message}
        </div>
      )}

      <div className="grid lg:grid-cols-3 gap-6">
        {/* Settings Panel */}
        <div className="lg:col-span-2 space-y-6">
          {/* Tabs */}
          <div className="flex gap-2 p-1 bg-gray-100 rounded-xl">
            {(['colors', 'branding', 'advanced'] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`flex-1 py-2 px-4 rounded-lg font-medium text-sm capitalize transition ${
                  activeTab === tab
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                {tab}
              </button>
            ))}
          </div>

          {/* Colors Tab */}
          {activeTab === 'colors' && (
            <Card>
              <h3 className="font-semibold mb-4">Color Scheme</h3>
              
              {/* Color Presets */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">Quick Presets</label>
                <div className="flex gap-2 flex-wrap">
                  {colorPresets.map((preset) => (
                    <button
                      key={preset.name}
                      onClick={() => applyPreset(preset)}
                      className="flex items-center gap-2 px-3 py-2 rounded-lg border border-gray-200 hover:border-gray-300 transition"
                    >
                      <div className="flex gap-1">
                        <div className="w-4 h-4 rounded-full" style={{ backgroundColor: preset.primary }}></div>
                        <div className="w-4 h-4 rounded-full" style={{ backgroundColor: preset.accent }}></div>
                      </div>
                      <span className="text-sm">{preset.name}</span>
                    </button>
                  ))}
                </div>
              </div>

              {/* Primary Color */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">Primary Color</label>
                <div className="flex gap-3">
                  <input
                    type="color"
                    value={config.primary_color}
                    onChange={(e) => handleColorChange('primary_color', e.target.value)}
                    className="w-12 h-12 rounded-lg cursor-pointer"
                  />
                  <input
                    type="text"
                    value={config.primary_color}
                    onChange={(e) => handleColorChange('primary_color', e.target.value)}
                    className="flex-1 px-4 py-3 border rounded-xl font-mono"
                    placeholder="#0D9488"
                  />
                </div>
                <p className="text-xs text-gray-500 mt-1">Used for buttons, links, and main accents</p>
              </div>

              {/* Secondary Color */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">Secondary Color</label>
                <div className="flex gap-3">
                  <input
                    type="color"
                    value={config.secondary_color}
                    onChange={(e) => handleColorChange('secondary_color', e.target.value)}
                    className="w-12 h-12 rounded-lg cursor-pointer"
                  />
                  <input
                    type="text"
                    value={config.secondary_color}
                    onChange={(e) => handleColorChange('secondary_color', e.target.value)}
                    className="flex-1 px-4 py-3 border rounded-xl font-mono"
                  />
                </div>
              </div>

              {/* Accent Color */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">Accent Color</label>
                <div className="flex gap-3">
                  <input
                    type="color"
                    value={config.accent_color}
                    onChange={(e) => handleColorChange('accent_color', e.target.value)}
                    className="w-12 h-12 rounded-lg cursor-pointer"
                  />
                  <input
                    type="text"
                    value={config.accent_color}
                    onChange={(e) => handleColorChange('accent_color', e.target.value)}
                    className="flex-1 px-4 py-3 border rounded-xl font-mono"
                  />
                </div>
                <p className="text-xs text-gray-500 mt-1">Used for CTAs, notifications, and highlights</p>
              </div>

              <Button onClick={handleSave} isLoading={isSaving} className="w-full">
                Save Colors
              </Button>
            </Card>
          )}

          {/* Branding Tab */}
          {activeTab === 'branding' && (
            <Card>
              <h3 className="font-semibold mb-4">Branding</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Shop Name</label>
                  <Input
                    type="text"
                    value={config.shop_name}
                    onChange={(e) => handleColorChange('shop_name', e.target.value)}
                    placeholder="My Shop"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Shop Logo URL</label>
                  <Input
                    type="url"
                    value={config.shop_logo || ''}
                    onChange={(e) => handleColorChange('shop_logo', e.target.value)}
                    placeholder="https://example.com/logo.png"
                  />
                  <p className="text-xs text-gray-500 mt-1">Recommended: 200x50px PNG with transparent background</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Favicon URL</label>
                  <Input
                    type="url"
                    value={config.shop_favicon || ''}
                    onChange={(e) => handleColorChange('shop_favicon', e.target.value)}
                    placeholder="https://example.com/favicon.ico"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Custom Domain</label>
                  <Input
                    type="text"
                    value={config.custom_domain || ''}
                    onChange={(e) => handleColorChange('custom_domain', e.target.value)}
                    placeholder="shop.mydomain.com"
                  />
                  <p className="text-xs text-gray-500 mt-1">Point your domain's CNAME record to this service</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Footer Text</label>
                  <textarea
                    value={config.footer_text || ''}
                    onChange={(e) => handleColorChange('footer_text', e.target.value)}
                    className="w-full px-4 py-3 border rounded-xl"
                    rows={2}
                    placeholder="Powered by DukaPOS"
                  />
                </div>

                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={config.show_branding}
                    onChange={(e) => handleColorChange('show_branding', e.target.checked)}
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm text-gray-700">Show "Powered by DukaPOS" branding</span>
                </label>

                <Button onClick={handleSave} isLoading={isSaving} className="w-full">
                  Save Branding
                </Button>
              </div>
            </Card>
          )}

          {/* Advanced Tab */}
          {activeTab === 'advanced' && (
            <Card>
              <h3 className="font-semibold mb-4">Advanced</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Background Color</label>
                  <div className="flex gap-3">
                    <input
                      type="color"
                      value={config.background_color}
                      onChange={(e) => handleColorChange('background_color', e.target.value)}
                      className="w-12 h-12 rounded-lg cursor-pointer"
                    />
                    <input
                      type="text"
                      value={config.background_color}
                      onChange={(e) => handleColorChange('background_color', e.target.value)}
                      className="flex-1 px-4 py-3 border rounded-xl font-mono"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Text Color</label>
                  <div className="flex gap-3">
                    <input
                      type="color"
                      value={config.text_color}
                      onChange={(e) => handleColorChange('text_color', e.target.value)}
                      className="w-12 h-12 rounded-lg cursor-pointer"
                    />
                    <input
                      type="text"
                      value={config.text_color}
                      onChange={(e) => handleColorChange('text_color', e.target.value)}
                      className="flex-1 px-4 py-3 border rounded-xl font-mono"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Custom CSS</label>
                  <textarea
                    value={config.custom_css || ''}
                    onChange={(e) => handleColorChange('custom_css', e.target.value)}
                    className="w-full px-4 py-3 border rounded-xl font-mono text-sm"
                    rows={6}
                    placeholder=".custom-class { color: red; }"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Custom JavaScript</label>
                  <textarea
                    value={config.custom_js || ''}
                    onChange={(e) => handleColorChange('custom_js', e.target.value)}
                    className="w-full px-4 py-3 border rounded-xl font-mono text-sm"
                    rows={4}
                    placeholder="console.log('Hello');"
                  />
                  <p className="text-xs text-gray-500 mt-1">Advanced: Add custom JavaScript to all pages</p>
                </div>

                <Button onClick={handleSave} isLoading={isSaving} className="w-full">
                  Save Advanced Settings
                </Button>
              </div>
            </Card>
          )}
        </div>

        {/* Preview Panel */}
        <div className="lg:col-span-1">
          <Card className="sticky top-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Live Preview</h3>
              <button
                onClick={() => setPreviewMode(!previewMode)}
                className="text-sm text-primary hover:underline"
              >
                {previewMode ? 'Edit Mode' : 'Preview Mode'}
              </button>
            </div>
            
            <div 
              className="rounded-lg border overflow-hidden"
              style={{ backgroundColor: config.background_color }}
            >
              {/* Mock Mobile Preview */}
              <div className="bg-gray-800 p-2">
                <div className="flex justify-center gap-1">
                  <div className="w-3 h-3 rounded-full bg-red-500"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                  <div className="w-3 h-3 rounded-full bg-green-500"></div>
                </div>
              </div>
              
              <div className="p-4" style={{ color: config.text_color }}>
                {/* Header */}
                <div className="flex items-center gap-2 mb-4 pb-4 border-b" style={{ borderColor: config.primary_color + '20' }}>
                  {config.shop_logo ? (
                    <img src={config.shop_logo} alt="Logo" className="h-8" />
                  ) : (
                    <div 
                      className="w-8 h-8 rounded-lg flex items-center justify-center text-white font-bold"
                      style={{ backgroundColor: config.primary_color }}
                    >
                      {config.shop_name?.charAt(0) || 'S'}
                    </div>
                  )}
                  <span className="font-semibold">{config.shop_name || 'My Shop'}</span>
                </div>

                {/* Buttons */}
                <div className="space-y-2">
                  <button
                    className="w-full py-2 px-4 rounded-lg text-white font-medium"
                    style={{ backgroundColor: config.primary_color }}
                  >
                    Primary Button
                  </button>
                  <button
                    className="w-full py-2 px-4 rounded-lg font-medium"
                    style={{ 
                      backgroundColor: config.accent_color,
                      color: '#fff'
                    }}
                  >
                    Accent Button
                  </button>
                  <button
                    className="w-full py-2 px-4 rounded-lg border-2 font-medium"
                    style={{ 
                      borderColor: config.primary_color,
                      color: config.primary_color
                    }}
                  >
                    Outline Button
                  </button>
                </div>

                {/* Badges */}
                <div className="flex gap-2 mt-4">
                  <span 
                    className="px-2 py-1 rounded-full text-xs font-medium"
                    style={{ backgroundColor: config.primary_color + '20', color: config.primary_color }}
                  >
                    New
                  </span>
                  <span 
                    className="px-2 py-1 rounded-full text-xs font-medium"
                    style={{ backgroundColor: config.accent_color + '20', color: config.accent_color }}
                  >
                    Sale
                  </span>
                </div>
              </div>

              {/* Footer */}
              {config.show_branding && (
                <div className="p-2 text-center text-xs text-gray-400 border-t">
                  Powered by DukaPOS
                </div>
              )}
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
