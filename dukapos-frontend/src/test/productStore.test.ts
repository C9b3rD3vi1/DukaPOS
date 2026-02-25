import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useProductStore } from '@/stores/productStore'

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

vi.mock('@/db/db', () => ({
  dbProducts: {
    add: vi.fn().mockResolvedValue(1),
    getByServerId: vi.fn().mockResolvedValue(null),
    getAll: vi.fn().mockResolvedValue([]),
    update: vi.fn().mockResolvedValue(undefined),
    delete: vi.fn().mockResolvedValue(undefined),
    clear: vi.fn().mockResolvedValue(undefined)
  },
  dbSyncQueue: {
    add: vi.fn().mockResolvedValue(1)
  },
  db: {
    products: {
      add: vi.fn().mockResolvedValue(1)
    }
  }
}))

vi.mock('@/db/sync', () => ({
  syncEngine: {
    syncAll: vi.fn().mockResolvedValue({ success: true, synced: 0, failed: 0, errors: [], conflicts: [] })
  }
}))

describe('ProductStore', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useProductStore.setState({
      products: [],
      categories: [],
      isLoading: false,
      error: null,
      searchQuery: '',
      selectedCategory: ''
    })
  })

  it('should have initial state', () => {
    const store = useProductStore.getState()
    expect(store.products).toEqual([])
    expect(store.categories).toEqual([])
    expect(store.isLoading).toBe(false)
    expect(store.error).toBe(null)
  })

  it('should update search query', () => {
    const { setSearchQuery } = useProductStore.getState()
    setSearchQuery('test')
    expect(useProductStore.getState().searchQuery).toBe('test')
  })

  it('should update selected category', () => {
    const { setSelectedCategory } = useProductStore.getState()
    setSelectedCategory('Electronics')
    expect(useProductStore.getState().selectedCategory).toBe('Electronics')
  })

  it('should get product by barcode', () => {
    useProductStore.setState({
      products: [
        {
          id: 1,
          shop_id: 1,
          name: 'Test Product',
          category: 'Test',
          unit: 'pcs',
          cost_price: 100,
          selling_price: 150,
          currency: 'KES',
          current_stock: 10,
          low_stock_threshold: 5,
          barcode: '123456789',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ]
    })

    const { getProductByBarcode } = useProductStore.getState()
    const product = getProductByBarcode('123456789')
    expect(product?.name).toBe('Test Product')
  })

  it('should get low stock products', () => {
    useProductStore.setState({
      products: [
        {
          id: 1,
          shop_id: 1,
          name: 'Low Stock Product',
          category: 'Test',
          unit: 'pcs',
          cost_price: 100,
          selling_price: 150,
          currency: 'KES',
          current_stock: 3,
          low_stock_threshold: 10,
          barcode: '',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: 2,
          shop_id: 1,
          name: 'Normal Stock Product',
          category: 'Test',
          unit: 'pcs',
          cost_price: 100,
          selling_price: 150,
          currency: 'KES',
          current_stock: 50,
          low_stock_threshold: 10,
          barcode: '',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ]
    })

    const { getLowStockProducts } = useProductStore.getState()
    const lowStock = getLowStockProducts()
    expect(lowStock.length).toBe(1)
    expect(lowStock[0].name).toBe('Low Stock Product')
  })
})
