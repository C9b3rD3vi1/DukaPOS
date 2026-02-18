# DukaPOS Features

## 📋 Table of Contents

1. [MVP Features](#mvp-features)
2. [Pro Features](#pro-features)
3. [Enterprise Features](#enterprise-features)
4. [Technical Architecture](#technical-architecture)
5. [Design Patterns](#design-patterns)
6. [Database Schema](#database-schema)
7. [Command Reference](#command-reference)

---

## 🎯 MVP Features

### Minimum Viable Product - Launch Ready

| Feature | Description | Status |
|---------|-------------|--------|
| **WhatsApp Integration** | Receive and send messages via Twilio | ✅ Implemented |
| **Product Management** | Add, update, delete products | ✅ Implemented |
| **Sales Recording** | Record sales via WhatsApp commands | ✅ Implemented |
| **Inventory Tracking** | Real-time stock levels | ✅ Implemented |
| **Daily Reports** | Automatic daily sales summary | ✅ Implemented |
| **Low Stock Alerts** | Notify when items run low | ✅ Implemented |
| **Multi-Product Support** | Unlimited products | ✅ Implemented |
| **Basic Analytics** | Sales by product, quantity | ✅ Implemented |
| **Product Categories** | Organize products by type | ✅ Implemented |
| **Barcode Support** | Barcode lookup | ✅ Implemented |
| **Threshold Alerts** | Customizable low stock alerts | ✅ Implemented |
| **Weekly/Monthly Reports** | Historical sales analysis | ✅ Implemented |
| **Staff Management** | Multiple users per shop | ✅ Implemented |
| **Supplier Management** | Track suppliers | ✅ Implemented |
| **Order Management** | Supplier orders | ✅ Implemented |
| **USSD Support** | USSD menu interface | ✅ Implemented |

### MVP Command Set

| Command | Example | Description |
|---------|----------|-------------|
| `add [product] [price] [qty]` | `add bread 50 30` | Add stock |
| `sell [product] [qty]` | `sell bread 2` | Record sale |
| `stock` | `stock` | View inventory |
| `stock [product]` | `stock bread` | View specific item |
| `report` | `report` | Daily summary |
| `price [product] [newprice]` | `price bread 55` | Update price |
| `remove [product] [qty]` | `remove bread 5` | Remove stock |
| `help` | `help` | Show commands |

### MVP User Flow

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Shop Owner  │───▶│   WhatsApp   │───▶│  DukaPOS Bot │
│   sends:    │    │   message   │    │   processes  │
│ "add milk   │    │             │    │   command    │
│  60 20"     │    │             │    │              │
└──────────────┘    └──────────────┘    └──────────────┘
                                                 │
                                                 ▼
                                        ┌──────────────┐
                                        │   SQLite DB  │
                                        │  Updates    │
                                        │  inventory  │
                                        └──────────────┘
```

---

## 🚀 Pro Features

### Phase 2 - Monetization Ready

| Feature | Description | Priority | Status |
|---------|-------------|----------|--------|
| **M-Pesa Integration** | Accept payments via STK Push | P0 | ✅ Implemented |
| **Multiple Shops** | Manage 2+ shops from one account | P0 | ✅ Implemented |
| **Weekly Reports** | 7-day sales analysis | P1 | ✅ Implemented |
| **Monthly Reports** | 30-day comprehensive report | P1 | ✅ Implemented |
| **Staff Accounts** | Multiple users per shop | P1 | ✅ Implemented |
| **Supplier Orders** | Auto-order from suppliers | P2 | ✅ Implemented |
| **Product Categories** | Organize products by type | P2 | ✅ Implemented |
| **Barcode Support** | Scan product barcodes | P2 | ✅ Implemented |
| **Threshold Alerts** | Customizable low stock | P2 | ✅ Implemented |

### Pro Pricing

| Tier | Price | Features |
|------|-------|----------|
| **Free** | KSh 0 | MVP features, 1 shop, 50 products |
| **Pro** | KSh 500/mo | Everything in Free + M-Pesa, unlimited products |
| **Business** | KSh 1,500/mo | Everything in Pro + 5 staff, API access |

### Pro Command Set

| Command | Example | Description |
|---------|----------|-------------|
| `weekly` | `weekly` | This week's sales |
| `monthly` | `monthly` | This month's sales |
| `profit` | `profit` | Calculate profit |
| `category [name]` | `category drinks` | View category |
| `staff [name]` | `staff John` | Add staff |
| `order [product] [qty]` | `order milk 50` | Create order |

---

## 🏢 Enterprise Features

### Phase 3 - Scale Ready

| Feature | Description | Priority | Status |
|---------|-------------|----------|--------|
| **AI Predictions** | ML-based restock predictions | P2 | ✅ Implemented |
| **QR Payments** | Scan QR to pay | P3 | ✅ Implemented |
| **Customer Loyalty** | Points system | P3 | ✅ Implemented |
| **Multi-Currency** | USD, TZS, UGX support | P3 | ✅ Implemented |
| **POS Hardware** | Receipt printer support | P3 | ✅ Implemented |
| **API Access** | Third-party integrations | P3 | ✅ Implemented |
| **White Label** | Custom branding | P3 | ❌ Not implemented |
| **Web Dashboard** | Browser-based management | P2 | ✅ Implemented |
| **Mobile App** | iOS/Android apps | P3 | ❌ Not started |
| **Scheduled Reports** | Auto daily/weekly reports | P2 | ✅ Implemented |
| **Phone Verification** | OTP verification | P2 | ✅ Implemented |
| **Data Encryption** | AES-256-GCM encryption | P2 | ✅ Implemented |
| **Webhook Events** | Async event delivery | P2 | ✅ Implemented |

---

## 🏗️ Technical Architecture

### System Design

```
                           ┌─────────────────┐
                           │   WhatsApp      │
                           │   (Twilio)      │
                           └────────┬────────┘
                                    │
                           ┌────────▼────────┐
                           │   Load Balancer │
                           │   (Nginx)       │
                           └────────┬────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
              ┌─────▼─────┐  ┌────▼────┐  ┌─────▼─────┐
              │  API GW   │  │  Auth   │  │  Web App  │
              │  Service  │  │ Service │  │  (React)  │
              └─────┬─────┘  └─────────┘  └───────────┘
                    │
         ┌────────┼────────┐
         │        │        │
   ┌─────▼──┐ ┌──▼────┐ ┌──▼─────┐
   │Product │ │ Sales │ │ Report │
   │Service │ │Service│ │Service │
   └────┬───┘ └───┬───┘ └───┬───┘
        │         │         │
        └─────────┼─────────┘
                  │
         ┌────────▼────────┐
         │   Database      │
         │   (PostgreSQL)  │
         └─────────────────┘
```

### Microservices Architecture

| Service | Responsibility | Tech |
|---------|---------------|------|
| **API Gateway** | Routing, rate limiting | Go/Gin |
| **Auth Service** | User authentication | Go |
| **Product Service** | Inventory CRUD | Go |
| **Sales Service** | Transaction processing | Go |
| **Report Service** | Analytics, reports | Go |
| **Notification Service** | WhatsApp, SMS alerts | Go |
| **Payment Service** | M-Pesa integration | Go |
| **Web App** | Admin dashboard | React |

### Database Schema (PostgreSQL)

```sql
-- Shops table
CREATE TABLE shops (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    owner_name VARCHAR(100),
    plan VARCHAR(20) DEFAULT 'free',
    mpesa_shortcode VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Products table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    shop_id INTEGER REFERENCES shops(id),
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    unit VARCHAR(20) DEFAULT 'pcs',
    cost_price DECIMAL(12, 2) DEFAULT 0,
    selling_price DECIMAL(12, 2) NOT NULL,
    current_stock INTEGER DEFAULT 0,
    low_stock_threshold INTEGER DEFAULT 10,
    barcode VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Sales table
CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    shop_id INTEGER REFERENCES shops(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    payment_method VARCHAR(20) DEFAULT 'cash',
    mpesa_receipt VARCHAR(50),
    staff_id INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Daily summary (cached)
CREATE TABLE daily_summaries (
    id SERIAL PRIMARY KEY,
    shop_id INTEGER REFERENCES shops(id),
    date DATE NOT NULL,
    total_sales DECIMAL(12, 2) DEFAULT 0,
    total_transactions INTEGER DEFAULT 0,
    total_profit DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🎨 Design Patterns

### 1. Clean Architecture

```
┌─────────────────────────────────────────────┐
│              Presentation Layer             │
│         (Handlers, Web Controllers)        │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│              Application Layer              │
│            (Use Cases, Services)           │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│                Domain Layer                 │
│           (Entities, Business Rules)        │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│            Infrastructure Layer             │
│      (Database, External APIs)             │
└─────────────────────────────────────────────┘
```

### 2. Repository Pattern

```go
// Product repository interface
type ProductRepository interface {
    GetByID(id int64) (*Product, error)
    GetByShopID(shopID int64) ([]Product, error)
    Create(product *Product) error
    Update(product *Product) error
    Delete(id int64) error
}

// Concrete implementation
type SQLiteProductRepository struct {
    db *sql.DB
}
```

### 3. Service Layer Pattern

```go
type InventoryService struct {
    productRepo ProductRepository
    saleRepo SaleRepository
    notifier NotificationService
}

func (s *InventoryService) ProcessSale(shopID int64, productName string, qty int) (*Sale, error) {
    // 1. Get product
    // 2. Check stock
    // 3. Create sale
    // 4. Update inventory
    // 5. Send notification if low stock
    // 6. Return result
}
```

### 4. Command Pattern (WhatsApp Parser)

```go
type Command interface {
    Execute(ctx *CommandContext) (*Response, error)
    Validate(ctx *CommandContext) error
}

type AddCommand struct{}
type SellCommand struct{}
type StockCommand struct{}

// Command parser
type Parser struct {
    commands map[string]Command
}

func (p *Parser) Parse(input string) (Command, error) {
    // Parse "add bread 50 30" → AddCommand{product: "bread", price: 50, qty: 30}
}
```

---

## 📱 WhatsApp Command Reference

### Free Tier Commands

| Command | Example | Response |
|---------|---------|----------|
| `add [product] [price] [qty]` | `add bread 50 30` | ✅ Added 30 bread @ KSh 50 |
| `sell [product] [qty]` | `sell bread 2` | ✅ Sold 2 bread. Total: KSh 100 |
| `stock` | `stock` | 📋 Current inventory... |
| `stock [product]` | `stock bread` | 🍞 Bread: 28 left |
| `price [product]` | `price bread` | Current price: KSh 50 |
| `price [product] [new]` | `price bread 55` | ✅ Price updated to KSh 55 |
| `report` | `report` | 📊 Daily Report: Sales: KSh 5,000... |
| `remove [product] [qty]` | `remove bread 5` | ✅ Removed 5 bread |
| `low` | `low` | ⚠️ Low stock: Milk (3), Eggs (5) |
| `delete [product]` | `delete bread` | ✅ Product deleted |
| `help` | `help` | 📖 Available commands... |

### Pro Tier Commands

| Command | Example | Response |
|---------|---------|----------|
| `weekly` | `weekly` | 📊 Weekly Report: KSh 35,000 |
| `monthly` | `monthly` | 📊 Monthly Report: KSh 150,000 |
| `profit` | `profit` | 💰 Today's profit: KSh 2,500 |
| `mpesa [amount]` | `mpesa 100` | 💳 STK push sent... |
| `category [name]` | `category drinks` | 🥤 Drinks: Milk, Water, Soda |
| `supplier [product]` | `supplier milk` | 📦 Supplier: Brookside @ KSh 45 |

---

## 🔄 State Machine: Sale Flow

```
    ┌──────────┐
    │  Start   │
    └────┬─────┘
         │
         ▼
┌────────────────┐
│ Parse Command   │
│ (sell milk 2)  │
└────┬───────────┘
         │
         ▼
┌────────────────┐
│ Validate Input  │──── No ────┐
│ (product, qty) │            │
└────┬───────────┘            │
     │ Yes                    ▼
     ▼                  ┌──────────┐
┌────────────────┐     │  Error  │
│ Check Stock    │──── No ────┤
│ (milk >= 2)   │            │
└────┬───────────┘            │
     │ Yes                    ▼
     ▼                  ┌──────────┐
┌────────────────┐     │  Return  │
│ Create Sale    │     │  Error   │
│ (id, amount)  │     └──────────┘
└────┬───────────┘
     │
     ▼
┌────────────────┐
│ Update Stock   │
│ (milk - 2)    │
└────┬───────────┘
     │
     ▼
┌────────────────┐     ┌──────────┐
│ Check Low Stock│─Yes─▶│ Send    │
│ (milk <= 5)   │     │ Alert   │
└────┬───────────┘     └──────────┘
     │ No
     ▼
┌────────────────┐
│ Return Success │
│ Message        │
└───────┬────────┘
        │
        ▼
    ┌────────┐
    │  Done  │
    └────────┘
```

---

## 📈 Analytics Features

### Dashboard Metrics

| Metric | Calculation | Display |
|--------|-------------|---------|
| **Total Sales** | SUM(all sales today) | KSh |
| **Transactions** | COUNT(sales today) | Number |
| **Top Product** | MAX(sales by product) | Product Name |
| **Profit** | SUM(sales - costs) | KSh |
| **Low Stock** | WHERE stock < threshold | List |

### Report Types

1. **Daily Report** - Every morning at 8 AM
2. **Weekly Report** - Every Monday at 8 AM
3. **Monthly Report** - 1st of month
4. **On-Demand** - User requests via WhatsApp

---

## 🔐 Security

| Feature | Implementation |
|---------|---------------|
| **Authentication** | ✅ JWT tokens implemented |
| **Phone Verification** | ✅ OTP implemented |
| **Data Encryption** | ✅ AES-256-GCM encryption at rest |
| **SSL/TLS** | ✅ Configurable (external) |
| **Rate Limiting** | ✅ Implemented (per API key) |
| **Input Validation** | ✅ Sanitize all inputs |
| **SQL Injection** | ✅ Parameterized queries (GORM) |
| **Account Lockout** | ✅ Brute force protection |

---

## 📱 Supported Platforms

| Platform | Status |
|----------|--------|
| WhatsApp (Twilio) | ✅ Implemented |
| USSD | ✅ Implemented |
| Web Dashboard | ✅ Implemented |
| REST API | ✅ Implemented |
| Android App | ❌ Future |
| iOS App | ❌ Future |

---

## 🔗 Integration APIs

| Service | Purpose | Status |
|---------|---------|--------|
| Twilio | WhatsApp messaging | ✅ Implemented |
| Safaricom Daraja | M-Pesa payments | ✅ Implemented |
| Africa Talking | SMS notifications | ✅ Implemented |
| SendGrid | Email reports | ✅ Implemented |

---

## 📝 Changelog

### v1.0.0 (MVP) - Coming Soon
- WhatsApp bot core functionality
- Product management
- Sales recording
- Basic reporting

### v1.1.0 - Q2 2026
- M-Pesa integration
- Multiple shops
- Staff accounts

### v1.2.0 - Q3 2026
- Web dashboard
- AI predictions
- API access

---

## 🤝 Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

---

## 📄 License

MIT License - see [LICENSE](./LICENSE)
