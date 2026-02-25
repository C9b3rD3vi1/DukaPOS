import { dbSales, dbProducts, dbSyncQueue } from './db'

export interface SyncResult {
  success: boolean
  synced: number
  failed: number
  errors: string[]
  conflicts: ConflictInfo[]
}

export interface ConflictInfo {
  id: number
  type: string
  localVersion: unknown
  serverVersion: unknown
  resolved: 'local' | 'server' | 'merged' | 'pending'
}

export interface SyncConfig {
  maxRetries: number
  baseDelay: number
  maxDelay: number
  retryableErrors: string[]
}

const DEFAULT_CONFIG: SyncConfig = {
  maxRetries: 5,
  baseDelay: 1000,
  maxDelay: 60000,
  retryableErrors: [
    'network',
    'timeout',
    'ECONNREFUSED',
    'ETIMEDOUT',
    'ENOTFOUND',
    '500',
    '502',
    '503',
    '504'
  ]
}

export class SyncEngine {
  private static instance: SyncEngine
  private isSyncing = false
  private syncInterval: ReturnType<typeof setInterval> | null = null
  private config: SyncConfig
  private conflictResolution: Map<string, ConflictInfo> = new Map()
  private syncListeners: Set<(result: SyncResult) => void> = new Set()

  static getInstance(config?: Partial<SyncConfig>): SyncEngine {
    if (!SyncEngine.instance) {
      SyncEngine.instance = new SyncEngine(config)
    }
    return SyncEngine.instance
  }

  private constructor(config?: Partial<SyncConfig>) {
    this.config = { ...DEFAULT_CONFIG, ...config }
  }

  subscribe(listener: (result: SyncResult) => void): () => void {
    this.syncListeners.add(listener)
    return () => this.syncListeners.delete(listener)
  }

  private notifyListeners(result: SyncResult): void {
    this.syncListeners.forEach(listener => {
      try {
        listener(result)
      } catch (e) {
        console.error('Sync listener error:', e)
      }
    })
  }

  private isRetryableError(error: unknown): boolean {
    if (!error) return false
    const errorStr = String(error).toLowerCase()
    return this.config.retryableErrors.some(e => errorStr.includes(e.toLowerCase()))
  }

  private calculateBackoff(attempts: number): number {
    const delay = this.config.baseDelay * Math.pow(2, attempts)
    const jitter = Math.random() * 1000
    return Math.min(delay + jitter, this.config.maxDelay)
  }

  async syncAll(): Promise<SyncResult> {
    if (this.isSyncing) {
      return { success: false, synced: 0, failed: 0, errors: ['Sync already in progress'], conflicts: [] }
    }

    if (!navigator.onLine) {
      return { success: false, synced: 0, failed: 0, errors: ['No internet connection'], conflicts: [] }
    }

    this.isSyncing = true
    const result: SyncResult = {
      success: true,
      synced: 0,
      failed: 0,
      errors: [],
      conflicts: []
    }

    try {
      const pendingItems = await dbSyncQueue.getPending()
      
      for (const item of pendingItems) {
        try {
          await this.processQueueItemWithRetry(item)
          await dbSyncQueue.remove(item.id!)
          result.synced++
        } catch (error) {
          result.failed++
          
          if (this.isRetryableError(error) && item.attempts < this.config.maxRetries) {
            await dbSyncQueue.incrementAttempts(item.id!)
            result.errors.push(`Retry ${item.attempts + 1}/${this.config.maxRetries} for ${item.type}: ${error}`)
          } else {
            await dbSyncQueue.remove(item.id!)
            result.errors.push(`Failed permanently ${item.type}: ${error}`)
          }
        }
      }

      const unsyncedSales = await dbSales.getUnsynced()
      for (const sale of unsyncedSales) {
        if (sale.serverId) {
          await dbSales.markSynced(sale.id!, sale.serverId)
        }
      }

      result.conflicts = Array.from(this.conflictResolution.values())
      result.success = result.failed === 0 && result.conflicts.filter(c => c.resolved === 'pending').length === 0
    } catch (error) {
      result.success = false
      result.errors.push(`Sync failed: ${error}`)
    } finally {
      this.isSyncing = false
    }

    this.notifyListeners(result)
    return result
  }

  private async processQueueItemWithRetry(item: import('./db').DBSyncQueue): Promise<void> {
    const maxAttempts = item.attempts + 1
    
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        await this.processQueueItem(item)
        return
      } catch (error) {
        if (attempt < maxAttempts - 1 && this.isRetryableError(error)) {
          const backoff = this.calculateBackoff(attempt)
          await this.delay(backoff)
        } else if (attempt >= maxAttempts - 1) {
          throw error
        }
      }
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  private async processQueueItem(item: import('./db').DBSyncQueue): Promise<void> {
    const { type, action, data } = item

    switch (type) {
      case 'sale':
        await this.syncSale(action, data)
        break
      case 'product':
        await this.syncProduct(action, data)
        break
      case 'customer':
        await this.syncCustomer(action, data)
        break
      default:
        console.warn(`Unknown sync type: ${type}`)
    }
  }

  private async syncSale(action: string, data: unknown): Promise<void> {
    const saleData = data as {
      id?: number
      serverId?: number
      shop_id: number
      product_id: number
      quantity: number
      unit_price: number
      payment_method: string
      mpesa_phone?: string
      updatedAt?: string
    }

    const { api } = await import('@/api/client')
    
    if (action === 'create') {
      const response = await api.post('/v1/sales', saleData)
      const serverData = response.data as { data: { id: number } }
      
      const unsyncedSales = await dbSales.getUnsynced()
      const localSale = unsyncedSales.find(
        s => s.productId === saleData.product_id && 
             s.quantity === saleData.quantity &&
             !s.serverId
      )
      if (localSale && localSale.id) {
        await dbSales.markSynced(localSale.id, serverData.data.id)
      }
    } else if (action === 'update') {
      await api.put(`/v1/sales/${saleData.id}`, saleData)
    } else if (action === 'delete') {
      await api.delete(`/v1/sales/${saleData.id}`)
    }
  }

  private async syncProduct(action: string, data: unknown): Promise<void> {
    const productData = data as {
      id?: number
      serverId?: number
      shop_id: number
      name: string
      category?: string
      unit: string
      cost_price: number
      selling_price: number
      current_stock: number
      low_stock_threshold: number
      barcode?: string
      updatedAt?: string
    }

    const { api } = await import('@/api/client')

    if (action === 'create') {
      const response = await api.post('/v1/products', productData)
      const serverData = response.data as { data: { id: number } }
      
      const localProduct = await dbProducts.getByServerId(serverData.data.id)
      if (localProduct && localProduct.id) {
        await dbProducts.update({ ...localProduct, serverId: serverData.data.id, synced: true })
      }
    } else if (action === 'update') {
      const serverId = productData.serverId || productData.id
      
      try {
        const serverResponse = await api.get(`/v1/products/${serverId}`)
        const serverProduct = serverResponse.data as { data: unknown }
        
        const conflictResult = await this.resolveConflict(
          productData,
          serverProduct.data,
          'product',
          productData.updatedAt
        )
        
        if (conflictResult.resolution === 'server') {
          await dbProducts.update({ ...conflictResult.localData as import('./db').DBDProduct, synced: true })
        } else {
          const { id: _, serverId: __, ...updateData } = conflictResult.localData as Record<string, unknown>
          await api.put(`/v1/products/${serverId}`, updateData)
        }
      } catch {
        await api.put(`/v1/products/${serverId}`, productData)
      }
    } else if (action === 'delete') {
      await api.delete(`/v1/products/${productData.id}`)
    }
  }

  private async syncCustomer(action: string, data: unknown): Promise<void> {
    const customerData = data as {
      id?: number
      shop_id: number
      name: string
      phone: string
      email?: string
      updatedAt?: string
    }

    const { api } = await import('@/api/client')

    if (action === 'create') {
      await api.post('/v1/customers', customerData)
    } else if (action === 'update') {
      await api.put(`/v1/customers/${customerData.id}`, customerData)
    }
  }

  private async resolveConflict(
    localData: unknown,
    serverData: unknown,
    _type: string,
    localTimestamp?: string
  ): Promise<{ resolution: 'local' | 'server' | 'merged'; localData: unknown }> {
    const localTime = localTimestamp ? new Date(localTimestamp).getTime() : 0
    const serverTime = (serverData as { updated_at?: string })?.updated_at 
      ? new Date((serverData as { updated_at: string }).updated_at).getTime() 
      : 0

    if (localTime > serverTime) {
      return { resolution: 'local', localData }
    } else if (serverTime > localTime) {
      return { resolution: 'server', localData: serverData }
    } else {
      return { resolution: 'merged', localData: { ...(serverData as object), ...(localData as object) } }
    }
  }

  async resolveConflictManual(
    _type: string,
    id: number,
    resolution: 'local' | 'server'
  ): Promise<void> {
    const key = `${_type}-${id}`
    const conflict = this.conflictResolution.get(key)
    
    if (!conflict) {
      throw new Error(`No conflict found for ${_type} #${id}`)
    }

    conflict.resolved = resolution
    this.conflictResolution.set(key, conflict)

    if (resolution === 'server') {
      await this.pullFromServer(_type, id)
    } else {
      await this.pushToServer(_type, id)
    }
  }

  private async pullFromServer(type: string, id: number): Promise<void> {
    const { api } = await import('@/api/client')
    
    try {
      if (type === 'product') {
        const response = await api.get(`/v1/products/${id}`)
        const serverProduct = response.data as import('./db').DBDProduct
        await dbProducts.update({ ...serverProduct, synced: true })
      }
    } catch (error) {
      console.error(`Failed to pull ${type} #${id} from server:`, error)
      throw error
    }
  }

  private async pushToServer(type: string, id: number): Promise<void> {
    const { api } = await import('@/api/client')
    
    try {
      if (type === 'product') {
        const product = await dbProducts.getByServerId(id)
        if (product) {
          await api.put(`/v1/products/${id}`, product)
        }
      }
    } catch (error) {
      console.error(`Failed to push ${type} #${id} to server:`, error)
      throw error
    }
  }

  async queueForSync(
    type: 'sale' | 'product' | 'customer',
    action: 'create' | 'update' | 'delete',
    data: unknown,
    _priority: 'high' | 'normal' | 'low' = 'normal'
  ): Promise<void> {
    await dbSyncQueue.add({
      type,
      action,
      data: {
        ...(data as object),
        updatedAt: new Date().toISOString()
      },
      attempts: 0,
      createdAt: new Date(),
      lastAttempt: undefined
    })
  }

  async queueSaleForSync(
    sale: Omit<import('./db').DBSale, 'id' | 'synced' | 'createdAt'>
  ): Promise<void> {
    await dbSales.add({
      ...sale,
      synced: false,
      createdAt: new Date()
    })
    
    await this.queueForSync('sale', 'create', {
      shop_id: sale.shopId,
      product_id: sale.productId,
      quantity: sale.quantity,
      unit_price: sale.unitPrice,
      total_amount: sale.totalAmount,
      payment_method: sale.paymentMethod,
      mpesa_phone: sale.mpesaReceipt,
      notes: sale.notes
    })
  }

  async queueProductForSync(
    product: Omit<import('./db').DBDProduct, 'id' | 'synced' | 'updatedAt'>,
    action: 'create' | 'update' | 'delete'
  ): Promise<void> {
    await this.queueForSync('product', action, product)
  }

  startAutoSync(intervalMs = 30000): void {
    if (this.syncInterval) {
      this.stopAutoSync()
    }

    const doSync = async () => {
      if (navigator.onLine) {
        try {
          await this.syncAll()
        } catch (error) {
          console.error('Auto sync error:', error)
        }
      }
    }

    doSync()
    this.syncInterval = setInterval(doSync, intervalMs)
  }

  stopAutoSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval)
      this.syncInterval = null
    }
  }

  async getPendingCount(): Promise<number> {
    return await dbSyncQueue.count()
  }

  async getConflicts(): Promise<ConflictInfo[]> {
    return Array.from(this.conflictResolution.values()).filter(c => c.resolved === 'pending')
  }

  async clearSyncQueue(): Promise<void> {
    const items = await dbSyncQueue.getAll()
    for (const item of items) {
      await dbSyncQueue.remove(item.id!)
    }
    this.conflictResolution.clear()
  }

  async clearFailedItems(): Promise<number> {
    const items = await dbSyncQueue.getAll()
    let cleared = 0
    for (const item of items) {
      if (item.attempts >= this.config.maxRetries) {
        await dbSyncQueue.remove(item.id!)
        cleared++
      }
    }
    return cleared
  }

  async pullProductsFromServer(shopId: number): Promise<void> {
    if (!navigator.onLine) return
    
    const { api } = await import('@/api/client')
    try {
      const response = await api.get(`/v1/products?shop_id=${shopId}&limit=1000`)
      const products = response.data?.data || response.data || []
      
      for (const product of products) {
        await dbProducts.add({
          serverId: product.id,
          name: product.name,
          category: product.category || '',
          unit: product.unit || 'pcs',
          costPrice: product.cost_price || 0,
          sellingPrice: product.selling_price,
          currency: product.currency || 'KES',
          currentStock: product.current_stock || 0,
          lowStockThreshold: product.low_stock_threshold || 10,
          barcode: product.barcode,
          imageUrl: product.image_url,
          isActive: product.is_active,
          updatedAt: new Date(product.updated_at),
          synced: true
        })
      }
    } catch (error) {
      console.error('Failed to pull products from server:', error)
    }
  }

  async pullSalesFromServer(shopId: number, limit = 100): Promise<void> {
    if (!navigator.onLine) return
    
    const { api } = await import('@/api/client')
    try {
      const response = await api.get(`/v1/sales?shop_id=${shopId}&limit=${limit}`)
      const sales = response.data?.data || response.data || []
      
      for (const sale of sales) {
        await dbSales.add({
          serverId: sale.id,
          productId: sale.product_id,
          productName: sale.product_name || '',
          quantity: sale.quantity,
          unitPrice: sale.unit_price,
          totalAmount: sale.total_amount,
          paymentMethod: sale.payment_method,
          mpesaReceipt: sale.mpesa_receipt,
          staffId: sale.staff_id,
          notes: sale.notes,
          createdAt: new Date(sale.created_at),
          synced: true
        })
      }
    } catch (error) {
      console.error('Failed to pull sales from server:', error)
    }
  }
}

export const syncEngine = SyncEngine.getInstance()
