import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { I18nProvider } from '@/utils/i18n'
import Dashboard from '@/pages/Dashboard'

vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn(() => ({
    shop: { id: 1, name: 'Test Shop' },
    user: { name: 'Test User' }
  }))
}))

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn().mockResolvedValue({
      data: {
        total_sales: 10000,
        total_profit: 3000,
        transaction_count: 50,
        product_count: 100,
        low_stock_count: 5,
        recent_sales: [],
        top_products: [],
        low_stock: []
      }
    })
  }
}))

const renderWithProviders = (component: React.ReactNode) => {
  return render(
    <BrowserRouter>
      <I18nProvider>
        {component}
      </I18nProvider>
    </BrowserRouter>
  )
}

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders dashboard page', async () => {
    renderWithProviders(<Dashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
  })

  it('displays stats cards', async () => {
    renderWithProviders(<Dashboard />)
    
    await waitFor(() => {
      expect(screen.getByText("Today's Sales")).toBeInTheDocument()
    })
  })

  it('displays weekly sales chart', async () => {
    renderWithProviders(<Dashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Weekly Sales')).toBeInTheDocument()
    })
  })
})
