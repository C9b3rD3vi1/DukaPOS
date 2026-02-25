import { Link } from 'react-router-dom'
import { Button } from '@/components/common/Button'

type EmptyStateVariant = 'products' | 'sales' | 'customers' | 'orders' | 'search' | 'generic'

interface EmptyStateProps {
  variant?: EmptyStateVariant
  title: string
  description?: string
  action?: {
    label: string
    to?: string
    onClick?: () => void
  }
}

const illustrations = {
  products: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <rect x="20" y="30" width="80" height="70" rx="8" fill="#0D9488" fillOpacity="0.1" stroke="#0D9488" strokeWidth="2"/>
      <rect x="30" y="40" width="25" height="20" rx="4" fill="#0D9488" fillOpacity="0.2"/>
      <rect x="60" y="40" width="25" height="20" rx="4" fill="#0D9488" fillOpacity="0.2"/>
      <rect x="30" y="65" width="25" height="20" rx="4" fill="#0D9488" fillOpacity="0.2"/>
      <rect x="60" y="65" width="25" height="20" rx="4" fill="#0D9488" fillOpacity="0.2"/>
      <circle cx="60" cy="25" r="8" fill="#0D9488" fillOpacity="0.3"/>
      <path d="M56 25L58 27L64 21" stroke="#0D9488" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  sales: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <rect x="15" y="25" width="90" height="70" rx="8" fill="#0D9488" fillOpacity="0.1" stroke="#0D9488" strokeWidth="2"/>
      <path d="M25 50h70M25 65h50M25 80h30" stroke="#0D9488" strokeWidth="3" strokeLinecap="round"/>
      <circle cx="85" cy="75" r="12" fill="#0D9488" fillOpacity="0.2"/>
      <path d="M80 75L88 75M85 70L85 80" stroke="#0D9488" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  customers: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <circle cx="60" cy="40" r="20" fill="#0D9488" fillOpacity="0.1" stroke="#0D9488" strokeWidth="2"/>
      <path d="M25 90c0-19.33 15.67-35 35-35s35 15.67 35 35" stroke="#0D9488" strokeWidth="2" strokeLinecap="round"/>
      <circle cx="45" cy="35" r="4" fill="#0D9488" fillOpacity="0.3"/>
      <circle cx="75" cy="35" r="4" fill="#0D9488" fillOpacity="0.3"/>
      <circle cx="60" cy="48" r="3" fill="#0D9488" fillOpacity="0.3"/>
    </svg>
  ),
  orders: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <rect x="20" y="20" width="80" height="80" rx="8" fill="#0D9488" fillOpacity="0.1" stroke="#0D9488" strokeWidth="2"/>
      <rect x="30" y="35" width="60" height="8" rx="2" fill="#0D9488" fillOpacity="0.2"/>
      <rect x="30" y="50" width="40" height="8" rx="2" fill="#0D9488" fillOpacity="0.2"/>
      <rect x="30" y="65" width="50" height="8" rx="2" fill="#0D9488" fillOpacity="0.2"/>
      <circle cx="80" cy="80" r="12" fill="#0D9488"/>
      <path d="M80 74v12M74 80h12" stroke="white" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
  search: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <circle cx="50" cy="50" r="25" stroke="#0D9488" strokeWidth="3" fill="none"/>
      <path d="M70 70l25 25" stroke="#0D9488" strokeWidth="3" strokeLinecap="round"/>
      <circle cx="50" cy="50" r="8" fill="#0D9488" fillOpacity="0.2"/>
      <path d="M45 50l3 3 7-7" stroke="#0D9488" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
  ),
  generic: (
    <svg className="w-32 h-32" viewBox="0 0 120 120" fill="none">
      <circle cx="60" cy="60" r="40" stroke="#0D9488" strokeWidth="2" fill="none" strokeDasharray="8 4"/>
      <path d="M50 55v15M60 50v20M70 55v15" stroke="#0D9488" strokeWidth="2" strokeLinecap="round"/>
    </svg>
  ),
}

export function EmptyState({ variant = 'generic', title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4">
      <div className="mb-6 animate-fade-in">
        {illustrations[variant]}
      </div>
      
      <h3 className="text-xl font-bold text-surface-900 mb-2 text-center">
        {title}
      </h3>
      
      {description && (
        <p className="text-surface-500 text-center max-w-sm mb-6">
          {description}
        </p>
      )}
      
      {action && (
        <div className="animate-slide-up">
          {action.to ? (
            <Link to={action.to}>
              <Button onClick={action.onClick}>
                {action.label}
              </Button>
            </Link>
          ) : (
            <Button onClick={action.onClick}>
              {action.label}
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
