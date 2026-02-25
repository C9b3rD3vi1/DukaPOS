import type { ReactNode } from 'react'

type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'

interface BadgeProps {
  children: ReactNode
  variant?: BadgeVariant
  size?: 'sm' | 'md' | 'lg'
  pulse?: boolean
  dot?: boolean
  className?: string
}

const variants: Record<BadgeVariant, string> = {
  default: 'bg-surface-100 text-surface-700',
  primary: 'bg-primary-100 text-primary-700',
  success: 'bg-green-100 text-green-700',
  warning: 'bg-amber-100 text-amber-700',
  danger: 'bg-red-100 text-red-700',
  info: 'bg-blue-100 text-blue-700',
}

const sizes = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-sm',
  lg: 'px-3 py-1.5 text-sm',
}

export function Badge({ 
  children, 
  variant = 'default', 
  size = 'md',
  pulse = false,
  dot = false,
  className = '' 
}: BadgeProps) {
  return (
    <span className={`
      inline-flex items-center gap-1.5 font-medium rounded-full
      ${variants[variant]}
      ${sizes[size]}
      ${className}
    `}>
      {dot && (
        <span className={`w-1.5 h-1.5 rounded-full ${pulse ? 'animate-pulse' : ''} ${
          variant === 'default' ? 'bg-surface-500' :
          variant === 'primary' ? 'bg-primary-500' :
          variant === 'success' ? 'bg-green-500' :
          variant === 'warning' ? 'bg-amber-500' :
          variant === 'danger' ? 'bg-red-500' :
          'bg-blue-500'
        }`} />
      )}
      {pulse && variant !== 'default' && (
        <span className="relative flex h-2 w-2">
          <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
            variant === 'primary' ? 'bg-primary-400' :
            variant === 'success' ? 'bg-green-400' :
            variant === 'warning' ? 'bg-amber-400' :
            variant === 'danger' ? 'bg-red-400' :
            'bg-blue-400'
          }`} />
          <span className={`relative inline-flex rounded-full h-2 w-2 ${
            variant === 'primary' ? 'bg-primary-500' :
            variant === 'success' ? 'bg-green-500' :
            variant === 'warning' ? 'bg-amber-500' :
            variant === 'danger' ? 'bg-red-500' :
            'bg-blue-500'
          }`} />
        </span>
      )}
      {children}
    </span>
  )
}

interface StatusBadgeProps {
  status: 'online' | 'offline' | 'pending' | 'syncing' | 'error' | 'success'
  label?: string
}

const statusConfig: Record<string, { variant: BadgeVariant; pulse?: boolean; dot?: boolean; label: string }> = {
  online: { variant: 'success', dot: true, label: 'Online' },
  offline: { variant: 'default', dot: true, label: 'Offline' },
  pending: { variant: 'warning', pulse: true, dot: true, label: 'Pending' },
  syncing: { variant: 'info', pulse: true, dot: true, label: 'Syncing' },
  error: { variant: 'danger', dot: true, label: 'Error' },
  success: { variant: 'success', dot: true, label: 'Success' },
}

export function StatusBadge({ status, label }: StatusBadgeProps) {
  const config = statusConfig[status]
  return (
    <Badge variant={config.variant} pulse={!!config.pulse} dot={config.dot}>
      {label || config.label}
    </Badge>
  )
}

interface NotificationBadgeProps {
  count: number
  max?: number
  className?: string
}

export function NotificationBadge({ count, max = 99, className = '' }: NotificationBadgeProps) {
  if (count <= 0) return null
  
  const displayCount = count > max ? `${max}+` : count
  
  return (
    <span className={`
      absolute -top-1 -right-1 min-w-[18px] h-[18px] 
      flex items-center justify-center
      bg-red-500 text-white text-xs font-bold
      rounded-full px-1
      animate-bounce-soft
      ${className}
    `}>
      {displayCount}
    </span>
  )
}

interface TabProps {
  active?: boolean
  children: ReactNode
  count?: number
  onClick?: () => void
}

export function Tab({ active = false, children, count, onClick }: TabProps) {
  return (
    <button
      onClick={onClick}
      className={`
        relative px-4 py-2 rounded-full text-sm font-medium transition-all
        ${active 
          ? 'bg-primary text-white shadow-lg shadow-primary/25' 
          : 'text-surface-600 hover:bg-surface-100 hover:text-surface-900'
        }
      `}
    >
      {children}
      {count !== undefined && count > 0 && (
        <span className={`
          ml-1.5 px-1.5 py-0.5 text-xs rounded-full
          ${active ? 'bg-white/20 text-white' : 'bg-surface-200 text-surface-600'}
        `}>
          {count}
        </span>
      )}
    </button>
  )
}
