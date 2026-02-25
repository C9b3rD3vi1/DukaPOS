import React from 'react'

interface CardProps {
  children: React.ReactNode
  className?: string
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
  hover?: boolean
  onClick?: () => void
  variant?: 'default' | 'bordered' | 'elevated' | 'glass'
}

export function Card({
  children,
  className = '',
  padding = 'md',
  hover = false,
  onClick,
  variant = 'default'
}: CardProps) {
  const paddings = {
    none: '',
    sm: 'p-3',
    md: 'p-5',
    lg: 'p-6',
    xl: 'p-8'
  }

  const variants = {
    default: 'bg-white border border-surface-100 shadow-card',
    bordered: 'bg-white border-2 border-surface-200',
    elevated: 'bg-white shadow-soft',
    glass: 'bg-white/80 backdrop-blur-lg border border-white/20 shadow-card'
  }

  return (
    <div 
      className={`
        rounded-2xl transition-all duration-300 ease-smooth
        ${variants[variant]}
        ${paddings[padding]}
        ${hover || onClick ? 'cursor-pointer hover:shadow-card-hover hover:border-primary/20 hover:-translate-y-0.5 active:scale-[0.99]' : ''}
        ${className}
      `}
      onClick={onClick}
    >
      {children}
    </div>
  )
}

interface CardHeaderProps {
  title: string
  subtitle?: string
  action?: React.ReactNode
  icon?: React.ReactNode
}

export function CardHeader({ title, subtitle, action, icon }: CardHeaderProps) {
  return (
    <div className="flex items-start justify-between mb-4">
      <div className="flex items-center gap-3">
        {icon && (
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
            {icon}
          </div>
        )}
        <div>
          <h3 className="font-semibold text-surface-900">{title}</h3>
          {subtitle && <p className="text-sm text-surface-500 mt-0.5">{subtitle}</p>}
        </div>
      </div>
      {action && <div>{action}</div>}
    </div>
  )
}

interface StatCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon?: React.ReactNode
  trend?: {
    value: number
    isPositive: boolean
  }
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info'
  onClick?: () => void
}

export function StatCard({
  title,
  value,
  subtitle,
  icon,
  trend,
  variant = 'default',
  onClick
}: StatCardProps) {
  const variants = {
    default: {
      bg: 'from-surface-50 to-surface-100',
      icon: 'bg-surface-200 text-surface-600',
      trend: 'text-surface-600'
    },
    success: {
      bg: 'from-green-50 to-emerald-50',
      icon: 'bg-green-100 text-green-600',
      trend: 'text-green-600'
    },
    warning: {
      bg: 'from-amber-50 to-orange-50',
      icon: 'bg-amber-100 text-amber-600',
      trend: 'text-amber-600'
    },
    danger: {
      bg: 'from-red-50 to-rose-50',
      icon: 'bg-red-100 text-red-600',
      trend: 'text-red-600'
    },
    info: {
      bg: 'from-blue-50 to-indigo-50',
      icon: 'bg-blue-100 text-blue-600',
      trend: 'text-blue-600'
    }
  }

  const v = variants[variant]

  return (
    <Card hover={!!onClick} onClick={onClick} className="relative overflow-hidden">
      <div className={`absolute inset-0 bg-gradient-to-br ${v.bg} opacity-50`} />
      <div className="relative">
        <div className="flex items-start justify-between mb-3">
          <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${v.icon}`}>
            {icon}
          </div>
          {trend && (
            <div className={`flex items-center gap-1 text-sm font-semibold ${v.trend}`}>
              <span className={trend.isPositive ? '' : 'rotate-180'}>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 10l7-7m0 0l7 7m-7-7v18" />
                </svg>
              </span>
              {Math.abs(trend.value)}%
            </div>
          )}
        </div>
        <p className="text-sm font-medium text-surface-500">{title}</p>
        <p className="text-2xl font-bold text-surface-900 mt-1">{value}</p>
        {subtitle && (
          <p className="text-xs text-surface-400 mt-2">{subtitle}</p>
        )}
      </div>
    </Card>
  )
}

interface StatGridProps {
  children: React.ReactNode
  columns?: 2 | 3 | 4
}

export function StatGrid({ children, columns = 4 }: StatGridProps) {
  const cols = {
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4'
  }

  return (
    <div className={`grid ${cols[columns]} gap-4`}>
      {children}
    </div>
  )
}

export default Card
