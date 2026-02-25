// API Types for DukaPOS

export interface Shop {
  id: number
  account_id: number
  name: string
  phone: string
  owner_name?: string
  address?: string
  plan: 'free' | 'pro' | 'business'
  mpesa_shortcode?: string
  is_active: boolean
  email?: string
  created_at: string
  updated_at: string
}

export interface Account {
  id: number
  email: string
  name: string
  phone: string
  is_active: boolean
  is_verified: boolean
  is_admin: boolean
  plan: 'free' | 'pro' | 'business'
  created_at: string
}

export interface Product {
  id: number
  shop_id: number
  name: string
  category?: string
  unit: string
  cost_price: number
  selling_price: number
  currency: string
  current_stock: number
  low_stock_threshold: number
  barcode?: string
  image_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Sale {
  id: number
  shop_id: number
  product_id: number
  product?: Product
  customer_id?: number
  quantity: number
  unit_price: number
  total_amount: number
  cost_amount: number
  profit: number
  payment_method: 'cash' | 'mpesa' | 'card' | 'bank'
  mpesa_receipt?: string
  mpesa_phone?: string
  staff_id?: number
  notes?: string
  created_at: string
}

export interface Customer {
  id: number
  shop_id: number
  name: string
  phone: string
  email?: string
  loyalty_points: number
  total_purchases: number
  created_at: string
}

export interface Supplier {
  id: number
  shop_id: number
  name: string
  phone?: string
  email?: string
  address?: string
  created_at: string
}

export interface Order {
  id: number
  shop_id: number
  supplier_id: number
  supplier?: Supplier
  status: string
  total_amount: number
  notes?: string
  items?: OrderItem[]
  created_at: string
  updated_at: string
}

export interface OrderItem {
  id?: number
  product_id: number
  product?: Product
  quantity: number
  price: number
  total: number
}

export interface Staff {
  id: number
  shop_id: number
  name: string
  phone: string
  role: string
  is_active: boolean
  created_at: string
}

export interface DailySummary {
  id: number
  shop_id: number
  date: string
  total_sales: number
  total_transactions: number
  total_profit: number
  total_cost: number
}

export interface DashboardData {
  shop?: Shop
  total_sales: number
  total_profit: number
  transaction_count: number
  product_count: number
  low_stock_count: number
  recent_sales: Sale[]
  top_products: Product[]
  today_sales: number
  total_products: number
  low_stock: Product[]
}

export interface AuthResponse {
  token: string
  user?: Account
  account?: Account
  shop?: Shop
}

export interface ApiError {
  error: string
  message?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface MpesaPayment {
  id: number
  shop_id: number
  phone: string
  amount: number
  checkout_request_id: string
  status: 'pending' | 'completed' | 'failed'
  mpesa_receipt?: string
  created_at: string
}

export interface AIPredictions {
  product_id: number
  product_name: string
  current_stock: number
  avg_daily_sales: number
  days_until_stockout: number
  recommended_order: number
  confidence: number
  trend: 'up' | 'down' | 'stable'
}

export interface SalesReport {
  period: 'daily' | 'weekly' | 'monthly'
  total_sales: number
  total_profit: number
  transaction_count: number
  top_products: Array<{
    name: string
    quantity: number
    revenue: number
    percentage: number
  }>
}

export interface APIKey {
  id: number
  key: string
  name: string
  last_used?: string
  created_at: string
}

export interface Webhook {
  id: number
  url: string
  events: string[]
  secret: string
  is_active: boolean
  created_at: string
}

export interface SMSMessage {
  id: number
  to: string
  message: string
  status: 'pending' | 'sent' | 'delivered' | 'failed'
  created_at: string
}

export interface EmailMessage {
  id: number
  to: string
  subject: string
  status: 'pending' | 'sent' | 'failed'
  created_at: string
}

export interface BillingPlan {
  id: string
  name: string
  price: number
  features: string[]
}

export interface AdminDashboard {
  total_accounts: number
  total_shops: number
  total_revenue: number
  active_accounts: number
  new_accounts_today: number
  new_shops_today: number
}

export interface AdminAccount {
  id: number
  email: string
  name: string
  phone: string
  plan: string
  is_active: boolean
  is_verified: boolean
  shops_count: number
  created_at: string
}

export interface AdminUser {
  id: number
  name: string
  email: string
  phone: string
  role: string
  shop_id: number
  shop_name: string
  is_active: boolean
  created_at: string
}

export interface AdminShop {
  id: number
  account_id: number
  name: string
  phone: string
  owner_name: string
  email: string
  plan: string
  is_active: boolean
  created_at: string
}

export interface RevenueStats {
  daily: Array<{ date: string; revenue: number }>
  monthly: Array<{ month: string; revenue: number }>
  total: number
}

export interface NotificationPreferences {
  low_stock_alerts: boolean
  daily_reports: boolean
  order_updates: boolean
  marketing: boolean
}

export interface NotificationDevice {
  id: number
  device_token: string
  platform: 'ios' | 'android' | 'web'
  is_active: boolean
  last_used_at: string
  created_at: string
}

export interface PrinterConfig {
  id: number
  shop_id: number
  name: string
  type: 'thermal' | 'network' | 'usb'
  connection_type: string
  ip_address?: string
  port?: number
  is_default: boolean
  created_at: string
}

export interface QRPayment {
  id: string
  amount: number
  phone: string
  status: 'pending' | 'completed' | 'failed'
  qr_code?: string
  created_at: string
}

export interface LoyaltyCustomer {
  id: number
  shop_id: number
  customer_id: number
  points: number
  tier: 'bronze' | 'silver' | 'gold' | 'platinum'
  total_spent: number
  visits: number
  last_visit?: string
  created_at: string
  updated_at: string
}
