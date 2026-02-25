import { useState, useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'
import { api } from '@/api/client'
import type { Shop } from '@/api/types'

export default function Header() {
  const { shop, logout, setShop } = useAuthStore()
  const { isOnline, isSyncing, pendingCount } = useSyncStore()
  const [showShopSwitcher, setShowShopSwitcher] = useState(false)
  const [shops, setShops] = useState<Shop[]>([])
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetchShops()
  }, [])

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowShopSwitcher(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const fetchShops = async () => {
    try {
      const response = await api.get<{ data: Shop[] }>('/v1/shops')
      if (response.data?.data) {
        setShops(response.data.data)
      }
    } catch (e) {
      console.error('Failed to fetch shops:', e)
    }
  }

  const handleShopSwitch = async (newShop: Shop) => {
    setShop(newShop)
    setShowShopSwitcher(false)
  }

  return (
    <header className="fixed top-0 left-0 right-0 z-50 h-16 bg-white/80 backdrop-blur-xl border-b border-surface-200/50 shadow-sm">
      <div className="max-w-full mx-auto px-4 h-full flex items-center justify-between">
        <div className="flex items-center">
          <button 
            onClick={() => document.getElementById('sidebar')?.classList.toggle('-translate-x-full')}
            className="md:hidden p-2 -ml-2 mr-2 text-surface-600 hover:bg-surface-100 rounded-xl transition"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          
          <Link to="/dashboard" className="flex items-center gap-2.5">
            <div className="w-9 h-9 bg-gradient-to-br from-primary to-primary-700 rounded-xl flex items-center justify-center shadow-lg shadow-primary/30">
              <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <span className="text-lg font-bold bg-gradient-to-r from-surface-900 to-surface-700 bg-clip-text text-transparent hidden sm:block">DukaPOS</span>
          </Link>
          
          {/* Shop Switcher Dropdown */}
          <div className="relative ml-2" ref={dropdownRef}>
            <button
              onClick={() => setShowShopSwitcher(!showShopSwitcher)}
              className="flex items-center gap-2 px-3 py-1.5 bg-primary-50 hover:bg-primary-100 rounded-lg text-sm font-medium text-primary-700 transition"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
              <span className="max-w-[120px] truncate">{shop?.name || 'Select Shop'}</span>
              <svg className={`w-4 h-4 transition-transform ${showShopSwitcher ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {showShopSwitcher && shops.length > 0 && (
              <div className="absolute top-full left-0 mt-1 w-56 bg-white rounded-xl shadow-lg border border-gray-100 py-1 max-h-64 overflow-y-auto z-50">
                {shops.map((s) => (
                  <button
                    key={s.id}
                    onClick={() => handleShopSwitch(s)}
                    className={`w-full px-4 py-2.5 text-left hover:bg-gray-50 flex items-center gap-2 transition ${
                      s.id === shop?.id ? 'bg-primary-50 text-primary-700' : 'text-gray-700'
                    }`}
                  >
                    <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                    <span className="truncate">{s.name}</span>
                    {s.id === shop?.id && (
                      <svg className="w-4 h-4 ml-auto text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    )}
                  </button>
                ))}
                <div className="border-t border-gray-100 mt-1 pt-1">
                  <Link
                    to="/settings"
                    onClick={() => setShowShopSwitcher(false)}
                    className="block px-4 py-2 text-sm text-gray-500 hover:bg-gray-50"
                  >
                    + Add New Shop
                  </Link>
                </div>
              </div>
            )}
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          {!isOnline && (
            <span className="px-3 py-1.5 bg-amber-100 text-amber-700 rounded-lg text-xs font-medium">
              Offline
            </span>
          )}
          
          {isSyncing && (
            <span className="px-3 py-1.5 bg-blue-100 text-blue-700 rounded-lg text-xs font-medium flex items-center gap-1.5">
              <span className="w-3 h-3 border-2 border-blue-700 border-t-transparent rounded-full animate-spin"></span>
              Syncing
            </span>
          )}
          
          {pendingCount > 0 && isOnline && !isSyncing && (
            <span className="px-3 py-1.5 bg-orange-100 text-orange-700 rounded-lg text-xs font-medium">
              {pendingCount} pending
            </span>
          )}
          
          <button
            onClick={logout}
            className="p-2.5 text-surface-500 hover:text-red-600 hover:bg-red-50 rounded-xl transition"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      </div>
    </header>
  )
}
