# DukaPOS 🛒

**WhatsApp POS for Kenyan Duka & Kiosk Owners**

Simple stock & sales management via WhatsApp. No app download, no training needed.

---

## 🎯 What is DukaPOS?

DukaPOS is a WhatsApp-based Point of Sale system designed for Kenyan duka and kiosk owners who already use WhatsApp daily. Manage inventory, track sales, and see profits - all from WhatsApp.

---

## 🚀 Features

### MVP (Available Now)
- [x] Add inventory/stock via WhatsApp
- [x] Record sales via WhatsApp
- [x] Daily sales summary
- [x] Low stock alerts
- [x] Multiple product support
- [x] Product pricing management
- [x] Stock removal/deduction

### Pro (Available Now)
- [x] Multiple shops support
- [x] Weekly/monthly reports
- [x] Supplier management
- [x] Order management
- [x] Staff management
- [x] Product categories
- [x] Barcode support
- [x] Threshold alerts
- [x] M-Pesa integration (STK Push, callbacks)

### Enterprise
- [x] Customer loyalty program
- [x] API for third-party integrations
- [x] QR code payments
- [x] Multi-currency support
- [x] AI restock predictions
- [ ] Mobile app (iOS/Android) - Future

---

## 📱 How It Works

### For Shop Owners:
1. Save DukaPOS WhatsApp number
2. Send commands like:
   - `add bread 50 30` (add 30 bread at KSh 50)
   - `sell bread 2` (sold 2 bread)
   - `stock` (check current inventory)
   - `report` (get daily summary)
3. Receive instant reports and alerts

### Example Commands:
```
add milk 60 20          → Add 20 packets milk @ KSh 60
sell milk 5             → Sold 5 packets milk  
stock                   → Show current inventory
report                  → Today's sales summary
low                     → Show items below threshold
profit                   → Calculate today's profit
```

---

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   WhatsApp      │────▶│   Go Backend    │────▶│    SQLite DB    │
│   (Twilio)      │     │   (Fiber)       │     │    (GORM)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                    │
                                    ▼
                             ┌─────────────────┐
                             │   M-Pesa API    │
                             │   (Daraja)      │
                             └─────────────────┘
```

### Technology Stack

| Layer | Technology |
|-------|------------|
| **Language** | Go 1.21+ |
| **Web Framework** | Fiber (Fasthttp) |
| **Database** | SQLite or PostgreSQL (GORM) |
| **WhatsApp** | Twilio API |
| **Payments** | Safaricom Daraja API |
| **SMS** | Africa Talking API |
| **Email** | SendGrid API |
| **Hosting** | Linux VPS (Ubuntu) |
| **Deployment** | Docker, Systemd + Nginx |

### Design Pattern: Clean Architecture

```
├── cmd/              # Entry points
├── internal/         # Core business logic
│   ├── handlers/    # HTTP handlers
│   ├── services/    # Business logic
│   └── models/      # Data models
├── pkg/             # Reusable packages
├── migrations/      # Database migrations
└── configs/        # Configuration
```

---

## 📦 Installation

### Prerequisites
- Go 1.21+
- SQLite3
- Twilio Account (WhatsApp Sandbox)

### Setup

1. Clone the repo:
```bash
git clone https://github.com/C9b3rD3vi1/DukaPOS.git
cd DukaPOS
```

2. Install dependencies:
```bash
go mod download
```

3. Set environment variables:
```bash
export TWILIO_ACCOUNT_SID=your_sid
export TWILIO_AUTH_TOKEN=your_token
export TWILIO_WHATSAPP_NUMBER=whatsapp:+14155238886
export DATABASE_PATH=./dukapos.db
export PORT=8080
```

4. Run:
```bash
go run cmd/server/main.go
```

### Docker (Alternative)

```bash
docker build -t dukapos .
docker run -p 8080:8080 dukapos
```

---

## 🔧 Configuration

| Variable | Description | Required |
|----------|-------------|----------|
| `TWILIO_ACCOUNT_SID` | Twilio Account SID | Yes |
| `TWILIO_AUTH_TOKEN` | Twilio Auth Token | Yes |
| `TWILIO_WHATSAPP_NUMBER` | Twilio WhatsApp number | Yes |
| `DATABASE_PATH` | Path to SQLite database | No |
| `DB_TYPE` | Database type (sqlite/postgres) | No |
| `DB_HOST` | PostgreSQL host | No |
| `DB_PORT` | PostgreSQL port | No |
| `DB_USER` | PostgreSQL user | No |
| `DB_PASSWORD` | PostgreSQL password | No |
| `DB_NAME` | PostgreSQL database name | No |
| `PORT` | Server port (default: 8080) | No |
| `MPESA_CONSUMER_KEY` | M-Pesa Daraja Consumer Key | No |
| `MPESA_CONSUMER_SECRET` | M-Pesa Daraja Consumer Secret | No |
| `MPESA_SHORTCODE` | M-Pesa Shortcode | No |
| `MPESA_PASSKEY` | M-Pesa Passkey | No |
| `AFRICA_TALKING_API_KEY` | Africa Talking API Key | No |
| `SENDGRID_API_KEY` | SendGrid API Key | No |
| `JWT_SECRET` | JWT Secret (change in production!) | No |

---

## 📁 Project Structure

```
DukaPOS/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration
│   ├── database/
│   │   └── db.go               # Database connection & migrations
│   ├── handlers/
│   │   ├── whatsapp.go         # WhatsApp webhook handler
│   │   ├── auth.go             # Authentication handlers
│   │   ├── customer.go          # Customer handlers
│   │   ├── api/                # API handlers
│   │   ├── mpesa/              # M-Pesa handlers
│   │   ├── staff/              # Staff handlers
│   │   ├── supplier/            # Supplier/Order handlers
│   │   ├── ussd/               # USSD handlers
│   │   └── webhook/             # Webhook handlers
│   ├── models/
│   │   └── models.go           # All data models
│   ├── repository/
│   │   └── repository.go       # Database repositories
│   ├── services/
│   │   ├── commands.go          # WhatsApp command handler
│   │   ├── auth.go             # Auth service
│   │   ├── mpesa/               # M-Pesa service
│   │   ├── ussd/               # USSD service
│   │   ├── ai/                 # AI predictions
│   │   ├── loyalty/             # Loyalty service
│   │   └── ...
│   └── middleware/
│       ├── middleware.go        # Auth, CORS, Rate limit
│       └── validation/          # Input validation
├── static/
│   └── index.html              # Web dashboard (static)
├── .env.example               # Environment template
├── go.mod
├── go.sum
└── README.md
```

---

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

---

## 📄 API Endpoints

### Webhooks (External)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /webhook/twilio | Twilio WhatsApp webhook |
| GET | /webhook/twilio/verify | Twilio webhook verification |
| POST | /webhook/twilio/status | WhatsApp message status |
| POST | /webhook/mpesa/stk | M-Pesa STK callback |
| POST | /webhook/mpesa/b2c | M-Pesa B2C callback |

### Public API
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/health | Health check |
| POST | /api/auth/register | Register new shop |
| POST | /api/auth/login | Login |

### Protected API (Requires JWT)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/v1/shop/profile | Get shop profile |
| PUT | /api/v1/shop/profile | Update shop profile |
| GET | /api/v1/shop/dashboard | Get dashboard data |
| GET | /api/v1/products | List products |
| POST | /api/v1/products | Create product |
| GET | /api/v1/products/:id | Get product |
| PUT | /api/v1/products/:id | Update product |
| DELETE | /api/v1/products/:id | Delete product |
| GET | /api/v1/sales | List sales |
| POST | /api/v1/sales | Record sale |
| GET | /api/v1/sales/:id | Get sale |
| GET | /api/v1/staff | List staff (Pro) |
| POST | /api/v1/staff | Add staff (Pro) |
| PUT | /api/v1/staff/:id | Update staff (Pro) |
| DELETE | /api/v1/staff/:id | Delete staff (Pro) |
| GET | /api/v1/suppliers | List suppliers (Pro) |
| POST | /api/v1/suppliers | Add supplier (Pro) |
| GET | /api/v1/orders | List orders (Pro) |
| POST | /api/v1/orders | Create order (Pro) |
| POST | /api/v1/mpesa/stk-push | Initiate STK push (Pro) |
| GET | /api/v1/mpesa/status/:id | Check payment status |
| GET | /api/v1/customers | List customers (Business) |
| POST | /api/v1/customers | Add customer (Business) |
| GET | /api/v1/api-keys | List API keys (Business) |
| POST | /api/v1/api-keys | Create API key (Business) |
| GET | /api/v1/webhooks | List webhooks (Business) |
| POST | /api/v1/webhooks | Create webhook (Business) |
| GET | /api/v1/ai/predictions/:shop_id | AI restock predictions |
| GET | /api/v1/ai/trends/:shop_id | Sales trends |
| POST | /api/v1/qr/generate | Generate QR payment |
| POST | /api/v1/sms/send | Send SMS |
| POST | /api/v1/email/send | Send email |

### API Documentation
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/docs/json | OpenAPI JSON documentation |
| GET | /api/docs/markdown | Markdown API documentation |

---

## 🤝 Contributing

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

---

## 📝 License

MIT License - see LICENSE file

---

## 👤 Author

**Nickson Wekongo**
- Security Engineer & Systems Developer
- Email: nicksonwekongo@gmail.com
- GitHub: [@C9b3rD3vi1](https://github.com/C9b3rD3vi1)
- Website: [simuxtech.com](https://simuxtech.com)

---

## 🙏 Acknowledgments

- Twilio for WhatsApp API
- Safaricom for M-Pesa Daraja API
- Go community

---

## 🔗 Related Projects

- [LinkBio.ke](https://linkbio.ke) - Link-in-bio SaaS platform
- [Cashflow Tracker](https://cashflow.simuxtech.com) - M-Pesa tracking bot

---

**Built with ❤️ in Kenya for Kenyan Businesses 🇰🇪**
