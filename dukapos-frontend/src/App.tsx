import { useEffect, Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useSyncStore } from '@/stores/syncStore'
import { I18nProvider } from '@/utils/i18n'
import { AddToHomeScreenPrompt } from '@/components/common/AddToHomeScreen'
import { ToastProvider, useToast } from '@/components/common/Toast'
import { OnboardingModal, useOnboarding } from '@/components/common/Onboarding'
import { setErrorHandler } from '@/api/client'
import Layout from '@/components/layout/Layout'
import AdminLayout from '@/components/layout/AdminLayout'
import AdminProtectedRoute from '@/components/AdminProtectedRoute'
import { usePushNotifications } from '@/hooks/usePWA'

// Lazy load all pages for code splitting
const Landing = lazy(() => import('@/pages/Landing'))
const Login = lazy(() => import('@/pages/Login'))
const Register = lazy(() => import('@/pages/Register'))
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Loyalty = lazy(() => import('@/pages/Loyalty'))
const Products = lazy(() => import('@/pages/Products'))
const ProductDetail = lazy(() => import('@/pages/ProductDetail'))
const Sales = lazy(() => import('@/pages/Sales'))
const NewSale = lazy(() => import('@/pages/NewSale'))
const Orders = lazy(() => import('@/pages/Orders'))
const Customers = lazy(() => import('@/pages/Customers'))
const Suppliers = lazy(() => import('@/pages/Suppliers'))
const Mpesa = lazy(() => import('@/pages/Mpesa'))
const Reports = lazy(() => import('@/pages/Reports'))
const Settings = lazy(() => import('@/pages/Settings'))
const Staff = lazy(() => import('@/pages/Staff'))
const AIInsights = lazy(() => import('@/pages/AIInsights'))
const APIKeys = lazy(() => import('@/pages/APIKeys'))
const SMS = lazy(() => import('@/pages/SMS'))
const Email = lazy(() => import('@/pages/Email'))
const Webhooks = lazy(() => import('@/pages/Webhooks'))
const Billing = lazy(() => import('@/pages/Billing'))
const Printer = lazy(() => import('@/pages/Printer'))
const Export = lazy(() => import('@/pages/Export'))
const WhiteLabel = lazy(() => import('@/pages/WhiteLabel'))
const ScheduledReports = lazy(() => import('@/pages/ScheduledReports'))
const StaffRoles = lazy(() => import('@/pages/StaffRoles'))


// Admin pages
const AdminLogin = lazy(() => import('@/pages/admin/Login'))
const AdminDashboard = lazy(() => import('@/pages/admin/Dashboard'))
const AdminShops = lazy(() => import('@/pages/admin/Shops'))
const AdminAccounts = lazy(() => import('@/pages/admin/Accounts'))
const AdminSubscriptions = lazy(() => import('@/pages/admin/Subscriptions'))
const AdminSettings = lazy(() => import('@/pages/admin/Settings'))
const AdminUsers = lazy(() => import('@/pages/admin/Users'))

// Loading component
function PageLoader() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-gray-600">Loading...</p>
      </div>
    </div>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const isLoading = useAuthStore((state) => state.isLoading)
  
  if (isLoading) {
    return <PageLoader />
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

function AppRoutes() {
  const { showToast } = useToast()
  const isLoading = useAuthStore((state) => state.isLoading)
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const initialize = useAuthStore((state) => state.initialize)
  const syncInitialize = useSyncStore((state) => state.initialize)
  const syncGetPendingCount = useSyncStore((state) => state.getPendingCount)
  const { showOnboarding, skipOnboarding } = useOnboarding()
  
  useEffect(() => {
    // Set up global error handler
    setErrorHandler((message, type) => {
      showToast(message, type)
    })
  }, [showToast])
  
  useEffect(() => {
    initialize()
    syncInitialize()
    syncGetPendingCount()
  }, [])
  
  if (isLoading) {
    return <PageLoader />
  }

  return (
    <>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {/* Public Routes */}
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <Login />} />
          <Route path="/register" element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <Register />} />
          <Route path="/admin/login" element={<AdminLogin />} />
          
          {/* Admin Routes - Protected */}
          <Route path="/admin" element={
            <AdminProtectedRoute>
              <AdminLayout />
            </AdminProtectedRoute>
          }>
            <Route index element={<AdminDashboard />} />
            <Route path="shops" element={<AdminShops />} />
            <Route path="accounts" element={<AdminAccounts />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="subscriptions" element={<AdminSubscriptions />} />
            <Route path="settings" element={<AdminSettings />} />
          </Route>
          
          {/* User Routes - Protected */}
          <Route path="/" element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }>
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="loyalty" element={<Loyalty />} />
            <Route path="products/new" element={<ProductDetail />} />
            <Route path="products" element={<Products />} />
            <Route path="products/:id" element={<ProductDetail />} />
            <Route path="sales" element={<Sales />} />
            <Route path="sales/new" element={<NewSale />} />
            <Route path="orders" element={<Orders />} />
            <Route path="customers" element={<Customers />} />
            <Route path="suppliers" element={<Suppliers />} />
            <Route path="mpesa" element={<Mpesa />} />
            <Route path="reports" element={<Reports />} />
            <Route path="staff" element={<Staff />} />
            <Route path="ai" element={<AIInsights />} />
            <Route path="apikeys" element={<APIKeys />} />
            <Route path="sms" element={<SMS />} />
            <Route path="email" element={<Email />} />
            <Route path="webhooks" element={<Webhooks />} />
            <Route path="billing" element={<Billing />} />
            <Route path="printer" element={<Printer />} />
            <Route path="white-label" element={<WhiteLabel />} />
            <Route path="staff-roles" element={<StaffRoles />} />
            <Route path="scheduled-reports" element={<ScheduledReports />} />
            <Route path="export" element={<Export />} />
            <Route path="settings" element={<Settings />} />
          </Route>
          
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
      
      {/* Onboarding Modal */}
      <OnboardingModal 
        isOpen={showOnboarding && isAuthenticated} 
        onClose={skipOnboarding} 
      />
    </>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
        <I18nProvider>
          <OfflineIndicator />
          <WebSocketInitializer />
          <PushNotificationInitializer />
          <AddToHomeScreenPrompt>
            <AppRoutes />
          </AddToHomeScreenPrompt>
        </I18nProvider>
      </ToastProvider>
    </BrowserRouter>
  )
}

function OfflineIndicator() {
  const { showToast } = useToast()
  const syncNow = useSyncStore((state) => state.syncNow)
  
  useEffect(() => {
    const handleOnline = () => {
      showToast('Back online! Syncing data...', 'success')
      syncNow()
    }
    
    const handleOffline = () => {
      showToast('You are offline. Changes will sync when connected.', 'warning')
    }
    
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    
    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [showToast, syncNow])
  
  return null
}

function WebSocketInitializer() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const token = useAuthStore((state) => state.token)
  const { showToast } = useToast()
  
  // Only connect WebSocket when authenticated
  useEffect(() => {
    if (!isAuthenticated || !token) return
    
    const ws = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?token=${token}`)
    
    ws.onopen = () => {
      console.log('Real-time updates connected')
    }
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        switch (message.type) {
          case 'new_sale':
            showToast(`New sale: ${message.payload?.product} - KES ${message.payload?.amount}`, 'success')
            break
          case 'low_stock':
            showToast(`Low stock alert: ${message.payload?.product} (${message.payload?.current_stock} remaining)`, 'warning')
            break
          case 'payment_received':
            showToast(`Payment received: KES ${message.payload?.amount}`, 'success')
            break
          case 'order_update':
            showToast(`Order #${message.payload?.order_id} updated: ${message.payload?.status}`, 'info')
            break
        }
      } catch (e) {
        console.error('WebSocket message error:', e)
      }
    }
    
    ws.onclose = () => {
      console.log('Real-time updates disconnected')
    }
    
    return () => {
      ws.close()
    }
  }, [isAuthenticated, token, showToast])
  
  if (!isAuthenticated) return null
  
  return null
}

function PushNotificationInitializer() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  
  const { requestPermission, isSupported } = usePushNotifications({
    onRegister: async (token) => {
      if (token && isAuthenticated) {
        try {
          const authToken = useAuthStore.getState().token
          if (authToken) {
            await fetch('/api/v1/notifications/register-device', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
              },
              body: JSON.stringify({
                device_token: token,
                platform: /iPhone|iPad/i.test(navigator.userAgent) ? 'ios' : 'android'
              })
            })
          }
        } catch (e) {
          console.error('Failed to register device token:', e)
        }
      }
    },
    onNotification: (notification) => {
      console.log('Push notification received:', notification)
      const n = notification as { title?: string; body?: string; data?: unknown }
      if (n.title) {
        new Notification(n.title, { body: n.body })
      }
    },
    onError: (error) => {
      console.error('Push notification error:', error)
    }
  })
  
  useEffect(() => {
    if (isSupported && isAuthenticated) {
      requestPermission()
    }
  }, [isSupported, isAuthenticated])
  
  return null
}
