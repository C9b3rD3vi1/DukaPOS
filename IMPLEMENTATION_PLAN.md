# DukaPOS PWA + Capacitor Implementation Plan

## Overview

This document outlines the complete implementation plan for rewriting DukaPOS frontend using PWA + Capacitor architecture to support Android, iOS, and Web platforms.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DukaPOS Architecture                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     FRONTEND (React + PWA)                         │    │
│  │                                                                      │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │    │
│  │  │   Android     │  │     iOS      │  │     Web      │           │    │
│  │  │   (APK)       │  │    (IPA)     │  │   (PWA)      │           │    │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │    │
│  │         │                  │                  │                    │    │
│  │         └──────────────────┼──────────────────┘                    │    │
│  │                            │                                          │    │
│  │                  ┌─────────▼─────────┐                              │    │
│  │                  │    Capacitor      │                              │    │
│  │                  │  ┌─────────────┐  │                              │    │
│  │                  │  │ Native APIs │  │                              │    │
│  │                  │  │ • Camera    │  │                              │    │
│  │                  │  │ • Storage   │  │                              │    │
│  │                  │  │ • Bluetooth │  │                              │    │
│  │                  │  │ • Push      │  │                              │    │
│  │                  │  └─────────────┘  │                              │    │
│  │                  └─────────┬─────────┘                              │    │
│  │                            │                                          │    │
│  │                  ┌─────────▼─────────┐                              │    │
│  │                  │   React SPA      │                              │    │
│  │                  │  ┌─────────────┐ │  ┌─────────────┐            │    │
│  │                  │  │ PWA Engine  │ │  │ Offline DB  │            │    │
│  │                  │  │ • SW        │ │  │ (Dexie.js)  │            │    │
│  │                  │  │ • Manifest   │ │  │ • Products  │            │    │
│  │                  │  │ • Sync      │ │  │ • Sales     │            │    │
│  │                  │  └─────────────┘ │  │ • Queue     │            │    │
│  │                  │                  │  └─────────────┘            │    │
│  │                  └──────────────────┘                               │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                       │                                     │
│                                       │ REST API                            │
│                                       ▼                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     BACKEND (Go + Fiber)                            │    │
│  │                                                                      │    │
│  │   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │    │
│  │   │ Products │  │  Sales   │  │  M-Pesa  │  │  WhatsApp │        │    │
│  │   │  API    │  │   API    │  │   API    │  │   Bot    │        │    │
│  │   └──────────┘  └──────────┘  └──────────┘  └──────────┘        │    │
│  │                                                                      │    │
│  │   ┌──────────────────────────────────────────────────────────┐     │    │
│  │   │                    PostgreSQL / SQLite                    │     │    │
│  │   └──────────────────────────────────────────────────────────┘     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Technology Stack

| Layer | Technology | Version | Purpose |
|-------|------------|---------|---------|
| **Framework** | React | 18.x | UI Framework |
| **Build Tool** | Vite | 5.x | Fast builds & HMR |
| **Language** | TypeScript | 5.x | Type safety |
| **State** | Zustand | 4.x | Lightweight state |
| **Offline DB** | Dexie.js | 4.x | IndexedDB wrapper |
| **HTTP** | Axios | 1.x | API client |
| **UI** | Tailwind CSS | 3.x | Styling |
| **PWA** | vite-plugin-pwa | 0.x | PWA generation |
| **Mobile** | Capacitor | 6.x | Native wrapper |
| **Barcode** | @capacitor-mlkit/barcode-scanning | 6.x | Barcode scanning |
| **Push** | @capacitor/push-notifications | 6.x | Push notifications |

## Project Structure

```
dukapos-frontend/
├── public/
│   ├── manifest.json          # Auto-generated by vite-plugin-pwa
│   ├── icons/                 # App icons (192, 512, favicon)
│   ├── splash/               # Splash screens
│   └── locales/              # i18n files (en, sw)
├── src/
│   ├── main.tsx              # Entry point
│   ├── App.tsx               # Root component
│   ├── index.css             # Global styles
│   │
│   ├── api/                  # API Layer
│   │   ├── client.ts         # Axios instance
│   │   ├── auth.ts           # Auth endpoints
│   │   ├── products.ts       # Product endpoints
│   │   ├── sales.ts          # Sales endpoints
│   │   ├── mpesa.ts          # M-Pesa endpoints
│   │   └── types.ts          # TypeScript interfaces
│   │
│   ├── db/                   # Offline Database
│   │   ├── db.ts             # Dexie configuration
│   │   ├── products.ts       # Product sync
│   │   ├── sales.ts          # Sales sync
│   │   └── sync.ts           # Sync engine
│   │
│   ├── stores/               # State Management (Zustand)
│   │   ├── authStore.ts      # Auth state
│   │   ├── shopStore.ts      # Shop state
│   │   ├── productStore.ts   # Products
│   │   ├── saleStore.ts      # Sales
│   │   └── syncStore.ts      # Offline sync status
│   │
│   ├── hooks/                # Custom React Hooks
│   │   ├── useAuth.ts        # Authentication
│   │   ├── useOnline.ts      # Online status
│   │   ├── useSync.ts        # Sync logic
│   │   ├── useBarcode.ts     # Barcode scanner
│   │   └── useCamera.ts      # Camera access
│   │
│   ├── components/           # Reusable Components
│   │   ├── common/
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Loader.tsx
│   │   │   ├── OfflineBadge.tsx
│   │   │   └── SyncStatus.tsx
│   │   ├── layout/
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── BottomNav.tsx
│   │   │   └── Layout.tsx
│   │   ├── dashboard/
│   │   │   ├── StatsGrid.tsx
│   │   │   ├── RecentSales.tsx
│   │   │   └── LowStockAlert.tsx
│   │   ├── products/
│   │   │   ├── ProductCard.tsx
│   │   │   ├── ProductList.tsx
│   │   │   ├── ProductForm.tsx
│   │   │   └── BarcodeScanner.tsx
│   │   ├── sales/
│   │   │   ├── SaleForm.tsx
│   │   │   ├── Cart.tsx
│   │   │   ├── Receipt.tsx
│   │   │   └── PaymentModal.tsx
│   │   └── reports/
│   │       ├── Chart.tsx
│   │       └── ReportCard.tsx
│   │
│   ├── pages/                # Page Components
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Products.tsx
│   │   ├── ProductDetail.tsx
│   │   ├── Sales.tsx
│   │   ├── NewSale.tsx
│   │   ├── Customers.tsx
│   │   ├── Suppliers.tsx
│   │   ├── Mpesa.tsx
│   │   ├── Reports.tsx
│   │   ├── Settings.tsx
│   │   └── Staff.tsx
│   │
│   └── utils/                # Utilities
│       ├── format.ts         # Number/date formatting
│       ├── validation.ts     # Form validation
│       ├── storage.ts        # Capacitor storage
│       └── logger.ts         # Logging
│
├── android/                  # Generated by Capacitor
├── ios/                      # Generated by Capacitor
│
├── capacitor.config.ts
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
└── package.json
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)

| Task | Description | Status |
|------|-------------|--------|
| 1.1 | Initialize React + TypeScript + Vite | [ ] |
| 1.2 | Configure TypeScript | [ ] |
| 1.3 | Setup Tailwind CSS | [ ] |
| 1.4 | Configure PWA (vite-plugin-pwa) | [ ] |
| 1.5 | Generate app icons | [ ] |
| 1.6 | Setup Capacitor | [ ] |
| 1.7 | Create folder structure | [ ] |
| 1.8 | Test build | [ ] |

### Phase 2: Core Features (Week 2)

| Task | Description | Status |
|------|-------------|--------|
| 2.1 | API Client & Types | [ ] |
| 2.2 | Authentication | [ ] |
| 2.3 | Layout Components | [ ] |
| 2.4 | Dashboard Page | [ ] |

### Phase 3: Business Logic (Week 3)

| Task | Description | Status |
|------|-------------|--------|
| 3.1 | Products Module | [ ] |
| 3.2 | Sales Module | [ ] |
| 3.3 | Reports Module | [ ] |
| 3.4 | Other Modules | [ ] |

### Phase 4: Offline-First (Week 4)

| Task | Description | Status |
|------|-------------|--------|
| 4.1 | Setup Dexie.js | [ ] |
| 4.2 | Implement Offline Logic | [ ] |
| 4.3 | Sync Engine | [ ] |
| 4.4 | Test Offline Mode | [ ] |

### Phase 5: Mobile Native (Week 5)

| Task | Description | Status |
|------|-------------|--------|
| 5.1 | Setup Capacitor Plugins | [ ] |
| 5.2 | Barcode Scanner | [ ] |
| 5.3 | Push Notifications | [ ] |
| 5.4 | Build APK/IPA | [ ] |

### Phase 6: Polish & Deploy (Week 6)

| Task | Description | Status |
|------|-------------|--------|
| 6.1 | Performance Optimization | [ ] |
| 6.2 | User Experience Polish | [ ] |
| 6.3 | PWA Enhancements | [ ] |
| 6.4 | Deployment | [ ] |

---

## Key Features to Implement

### Offline-First Architecture
- Cache API responses in IndexedDB
- Queue offline mutations for sync
- Background sync when online
- Conflict resolution strategy

### PWA Features
- Service Worker with caching strategies
- App manifest for installation
- Push notifications
- Add to Home Screen prompt
- Offline fallback page

### Mobile Native Features (via Capacitor)
- Barcode scanning
- Camera for product photos
- Push notifications
- Haptic feedback
- Local storage
- File system access

### Sync Strategy
```
┌─────────────────────────────────────────────────────────────┐
│                     Sync Strategy                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐      ┌─────────────┐      ┌───────────┐  │
│  │   Offline   │ ───► │   Queue     │ ───► │  Server   │  │
│  │   Action    │      │  (Dexie)    │      │   Sync    │  │
│  └─────────────┘      └─────────────┘      └───────────┘  │
│                              │                    │         │
│                              │                    │         │
│                              ▼                    ▼         │
│                      ┌─────────────┐      ┌─────────────┐  │
│                      │  Conflict  │      │   Pull      │  │
│                      │ Resolution │◄─────│   Updates   │  │
│                      └─────────────┘      └─────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## API Endpoints to Integrate

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/auth/login` | POST | Login |
| `/api/auth/register` | POST | Register |
| `/api/v1/shop/profile` | GET | Shop profile |
| `/api/v1/shop/dashboard` | GET | Dashboard data |
| `/api/v1/products` | GET/POST | Products CRUD |
| `/api/v1/sales` | GET/POST | Sales CRUD |
| `/api/v1/customers` | GET/POST | Customers CRUD |
| `/api/v1/suppliers` | GET/POST | Suppliers CRUD |
| `/api/v1/staff` | GET/POST | Staff CRUD |
| `/api/v1/mpesa/stk-push` | POST | M-Pesa payment |
| `/api/v1/ai/predictions/:shop_id` | GET | AI predictions |

---

## Testing Strategy

### Unit Tests
- Jest + React Testing Library
- Test components in isolation
- Mock API calls

### Integration Tests
- Test user flows
- Test offline/online sync

### E2E Tests
- Cypress for critical paths
- Test on real devices

### Manual Testing Checklist
- [ ] Test on Chrome (desktop)
- [ ] Test on Safari (iOS simulator)
- [ ] Test on Android emulator
- [ ] Test PWA installation
- [ ] Test offline mode
- [ ] Test barcode scanner

---

## Deployment

### Web (PWA)
- Platform: Vercel or Netlify
- CI/CD: GitHub Actions
- Domain: dukapos.com

### Android
- Platform: Google Play Store
- Package: com.dukapos.app
- Build: APK/AAB

### iOS
- Platform: Apple App Store
- Bundle ID: com.dukapos.app
- Build: IPA

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Lighthouse Score | >90 |
| First Contentful Paint | <1.5s |
| Time to Interactive | <3s |
| Offline Functionality | 100% |
| PWA Install Rate | >30% |

---

## Timeline

| Phase | Duration | Start | End |
|-------|----------|-------|-----|
| Phase 1: Foundation | 1 week | Week 1 | Week 1 |
| Phase 2: Core Features | 1 week | Week 2 | Week 2 |
| Phase 3: Business Logic | 1 week | Week 3 | Week 3 |
| Phase 4: Offline-First | 1 week | Week 4 | Week 4 |
| Phase 5: Mobile Native | 1 week | Week 5 | Week 5 |
| Phase 6: Polish & Deploy | 1 week | Week 6 | Week 6 |

**Total: 6 weeks**

---

## Notes

- This plan replaces the existing Go server-side templates
- Backend (Go/Fiber) remains unchanged
- Frontend will be a standalone React application
- Capacitor will wrap the React app as native mobile apps
- All API calls will proxy through the existing backend
