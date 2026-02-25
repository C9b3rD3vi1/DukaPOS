import { Card } from '@/components/common'

interface ReportCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon: React.ReactNode
  trend?: {
    value: number
    isPositive: boolean
  }
  variant?: 'default' | 'success' | 'warning' | 'danger'
  className?: string
}

export function ReportCard({
  title,
  value,
  subtitle,
  icon,
  trend,
  variant = 'default',
  className = ''
}: ReportCardProps) {
  const variants = {
    default: 'border-gray-100',
    success: 'border-green-100 bg-green-50/50',
    warning: 'border-amber-100 bg-amber-50/50',
    danger: 'border-red-100 bg-red-50/50'
  }

  return (
    <Card className={`${variants[variant]} ${className}`} padding="md">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
          {subtitle && (
            <p className="text-xs text-gray-400 mt-1">{subtitle}</p>
          )}
          {trend && (
            <div className={`flex items-center gap-1 mt-2 text-xs font-medium ${
              trend.isPositive ? 'text-green-600' : 'text-red-600'
            }`}>
              <span>{trend.isPositive ? '↑' : '↓'}</span>
              <span>{Math.abs(trend.value)}%</span>
              <span className="text-gray-400 font-normal">vs last period</span>
            </div>
          )}
        </div>
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
          variant === 'success' ? 'bg-green-100 text-green-600' :
          variant === 'warning' ? 'bg-amber-100 text-amber-600' :
          variant === 'danger' ? 'bg-red-100 text-red-600' :
          'bg-gray-100 text-gray-600'
        }`}>
          {icon}
        </div>
      </div>
    </Card>
  )
}

interface ReportGridProps {
  children: React.ReactNode
  columns?: 2 | 3 | 4
}

export function ReportGrid({ children, columns = 4 }: ReportGridProps) {
  const gridCols = {
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4'
  }

  return (
    <div className={`grid ${gridCols[columns]} gap-4`}>
      {children}
    </div>
  )
}

interface ReportSectionProps {
  title: string
  subtitle?: string
  action?: React.ReactNode
  children: React.ReactNode
}

export function ReportSection({ title, subtitle, action, children }: ReportSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          {subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}
        </div>
        {action}
      </div>
      {children}
    </div>
  )
}

interface PeriodSelectorProps {
  periods: Array<{ label: string; value: string }>
  selected: string
  onChange: (value: string) => void
}

export function PeriodSelector({ periods, selected, onChange }: PeriodSelectorProps) {
  return (
    <div className="flex items-center gap-1 p-1 bg-gray-100 rounded-lg">
      {periods.map(period => (
        <button
          key={period.value}
          onClick={() => onChange(period.value)}
          className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
            selected === period.value
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          {period.label}
        </button>
      ))}
    </div>
  )
}

export default ReportCard
