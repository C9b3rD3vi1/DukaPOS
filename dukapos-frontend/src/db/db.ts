import Dexie, { type Table } from 'dexie'

export interface DBDProduct {
  id?: number
  serverId: number
  name: string
  category: string
  unit: string
  costPrice: number
  sellingPrice: number
  currency: string
  currentStock: number
  lowStockThreshold: number
  barcode?: string
  imageUrl?: string
  isActive: boolean
  updatedAt: Date
  synced: boolean
}

export interface DBSale {
  id?: number
  serverId?: number
  shopId?: number
  productId: number
  productName: string
  quantity: number
  unitPrice: number
  totalAmount: number
  paymentMethod: 'cash' | 'mpesa' | 'card' | 'bank'
  mpesaReceipt?: string
  staffId?: number
  notes?: string
  createdAt: Date
  synced: boolean
}

export interface DBSyncQueue {
  id?: number
  type: 'sale' | 'product' | 'customer' | 'supplier' | 'order'
  action: 'create' | 'update' | 'delete'
  data: unknown
  attempts: number
  lastAttempt?: Date
  createdAt: Date
  priority?: 'high' | 'normal' | 'low'
}

export interface DBCustomer {
  id?: number
  serverId: number
  name: string
  phone: string
  email?: string
  loyaltyPoints: number
  totalPurchases: number
  updatedAt: Date
  synced: boolean
}

export interface DBSupplier {
  id?: number
  serverId: number
  name: string
  phone?: string
  email?: string
  address?: string
  updatedAt: Date
  synced: boolean
}

export interface DBOrder {
  id?: number
  serverId: number
  supplierId: number
  supplierName: string
  status: 'pending' | 'approved' | 'received' | 'cancelled'
  totalAmount: number
  notes?: string
  items: DBOrderItem[]
  createdAt: Date
  updatedAt: Date
  synced: boolean
}

export interface DBOrderItem {
  productId: number
  productName: string
  quantity: number
  price: number
  total: number
}

class DukaPOSDatabase extends Dexie {
  products!: Table<DBDProduct>
  sales!: Table<DBSale>
  syncQueue!: Table<DBSyncQueue>
  customers!: Table<DBCustomer>
  suppliers!: Table<DBSupplier>
  orders!: Table<DBOrder>

  constructor() {
    super('dukapos')
    this.version(1).stores({
      products: '++id, serverId, name, category, barcode, synced, isActive',
      sales: '++id, serverId, createdAt, synced, paymentMethod',
      syncQueue: '++id, type, action, createdAt, attempts, priority',
      customers: '++id, serverId, name, phone, synced'
    })
    this.version(2).stores({
      products: '++id, serverId, name, category, barcode, synced, isActive',
      sales: '++id, serverId, createdAt, synced, paymentMethod',
      syncQueue: '++id, type, action, createdAt, attempts, priority',
      customers: '++id, serverId, name, phone, synced',
      suppliers: '++id, serverId, name, phone, synced',
      orders: '++id, serverId, supplierId, status, createdAt, synced'
    })
  }
}

export const db = new DukaPOSDatabase()

export const dbProducts = {
  async add(product: DBDProduct): Promise<number> {
    return await db.products.add(product)
  },

  async getByServerId(serverId: number): Promise<DBDProduct | undefined> {
    return await db.products.where('serverId').equals(serverId).first()
  },

  async getAll(): Promise<DBDProduct[]> {
    return await db.products.where('isActive').equals(1).toArray()
  },

  async getLowStock(): Promise<DBDProduct[]> {
    const products = await db.products.toArray()
    return products.filter(p => p.currentStock <= p.lowStockThreshold)
  },

  async update(product: DBDProduct): Promise<void> {
    await db.products.put(product)
  },

  async delete(id: number): Promise<void> {
    await db.products.delete(id)
  },

  async clear(): Promise<void> {
    await db.products.clear()
  }
}

export const dbSales = {
  async add(sale: DBSale): Promise<number> {
    return await db.sales.add(sale)
  },

  async getUnsynced(): Promise<DBSale[]> {
    return await db.sales.where('synced').equals(0).toArray()
  },

  async markSynced(id: number, serverId: number): Promise<void> {
    await db.sales.update(id, { synced: true, serverId })
  },

  async getByDateRange(startDate: Date, endDate: Date): Promise<DBSale[]> {
    return await db.sales
      .where('createdAt')
      .between(startDate, endDate)
      .toArray()
  },

  async clear(): Promise<void> {
    await db.sales.clear()
  }
}

export const dbCustomers = {
  async add(customer: DBCustomer): Promise<number> {
    return await db.customers.add(customer)
  },

  async get(id: number): Promise<DBCustomer | undefined> {
    return await db.customers.get(id)
  },

  async getByServerId(serverId: number): Promise<DBCustomer | undefined> {
    return await db.customers.where('serverId').equals(serverId).first()
  },

  async getAll(): Promise<DBCustomer[]> {
    return await db.customers.toArray()
  },

  async update(customer: DBCustomer): Promise<void> {
    await db.customers.put(customer)
  },

  async delete(id: number): Promise<void> {
    await db.customers.delete(id)
  },

  async clear(): Promise<void> {
    await db.customers.clear()
  }
}

export const dbSyncQueue = {
  async add(item: Omit<DBSyncQueue, 'id'>): Promise<number> {
    return await db.syncQueue.add(item as DBSyncQueue)
  },

  async getAll(): Promise<DBSyncQueue[]> {
    return await db.syncQueue.orderBy('createdAt').toArray()
  },

  async getPending(): Promise<DBSyncQueue[]> {
    return await db.syncQueue.where('attempts').below(5).toArray()
  },

  async incrementAttempts(id: number): Promise<void> {
    const item = await db.syncQueue.get(id)
    if (item) {
      await db.syncQueue.update(id, {
        attempts: item.attempts + 1,
        lastAttempt: new Date()
      })
    }
  },

  async remove(id: number): Promise<void> {
    await db.syncQueue.delete(id)
  },

  async count(): Promise<number> {
    return await db.syncQueue.count()
  }
}

export const dbSuppliers = {
  async add(supplier: DBSupplier): Promise<number> {
    return await db.suppliers.add(supplier)
  },

  async getByServerId(serverId: number): Promise<DBSupplier | undefined> {
    return await db.suppliers.where('serverId').equals(serverId).first()
  },

  async getAll(): Promise<DBSupplier[]> {
    return await db.suppliers.toArray()
  },

  async update(supplier: DBSupplier): Promise<void> {
    await db.suppliers.put(supplier)
  },

  async delete(id: number): Promise<void> {
    await db.suppliers.delete(id)
  },

  async clear(): Promise<void> {
    await db.suppliers.clear()
  }
}

export const dbOrders = {
  async add(order: DBOrder): Promise<number> {
    return await db.orders.add(order)
  },

  async getByServerId(serverId: number): Promise<DBOrder | undefined> {
    return await db.orders.where('serverId').equals(serverId).first()
  },

  async getAll(): Promise<DBOrder[]> {
    return await db.orders.orderBy('createdAt').reverse().toArray()
  },

  async getByStatus(status: string): Promise<DBOrder[]> {
    return await db.orders.where('status').equals(status).toArray()
  },

  async update(order: DBOrder): Promise<void> {
    await db.orders.put(order)
  },

  async delete(id: number): Promise<void> {
    await db.orders.delete(id)
  },

  async clear(): Promise<void> {
    await db.orders.clear()
  }
}
