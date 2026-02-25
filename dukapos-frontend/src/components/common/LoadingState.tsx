import { Loader } from './Loader'

interface LoadingStateProps {
  isLoading: boolean
  skeleton?: React.ReactNode
  children: React.ReactNode
  fallback?: React.ReactNode
  variant?: 'spinner' | 'skeleton' | 'pulse' | 'overlay'
  size?: 'sm' | 'md' | 'lg'
  fullPage?: boolean
}

export function LoadingState({
  isLoading,
  skeleton,
  children,
  fallback,
  variant = 'spinner',
  size = 'md',
  fullPage = false
}: LoadingStateProps) {
  if (isLoading) {
    if (fallback) return <>{fallback}</>
    
    if (variant === 'skeleton' && skeleton) {
      return <>{skeleton}</>
    }

    const spinner = (
      <div className={`flex items-center justify-center ${fullPage ? 'min-h-[50vh]' : 'py-8'}`}>
        <Loader size={size} />
      </div>
    )

    if (fullPage) {
      return (
        <div className="fixed inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-50">
          {spinner}
        </div>
      )
    }

    return spinner
  }

  return <>{children}</>
}

export function LoadingOverlay({ isLoading, children }: { isLoading: boolean; children: React.ReactNode }) {
  return (
    <div className="relative">
      {children}
      {isLoading && (
        <div className="absolute inset-0 bg-white/60 backdrop-blur-sm flex items-center justify-center z-10 rounded-inherit">
          <Loader size="md" />
        </div>
      )}
    </div>
  )
}

export function LoadingButton({
  isLoading,
  children,
  className = '',
  disabled,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { isLoading: boolean }) {
  return (
    <button
      className={`relative ${className} ${isLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading && (
        <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
          <Loader size="sm" />
        </span>
      )}
      <span className={isLoading ? 'invisible' : ''}>{children}</span>
    </button>
  )
}

export function LoadingCard({ isLoading: _isLoading, className = '' }: { isLoading: boolean; className?: string }) {
  return (
    <div className={`bg-white rounded-xl border border-gray-200 p-6 ${className}`}>
      <div className="animate-pulse space-y-4">
        <div className="h-4 bg-gray-200 rounded w-1/3"></div>
        <div className="h-8 bg-gray-200 rounded w-2/3"></div>
        <div className="h-4 bg-gray-200 rounded w-1/2"></div>
      </div>
    </div>
  )
}

export function LoadingTable({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) {
  return (
    <div className="animate-pulse space-y-3">
      <div className="flex gap-4">
        {Array.from({ length: columns }).map((_, i) => (
          <div key={i} className="h-4 bg-gray-200 rounded flex-1"></div>
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex gap-4">
          {Array.from({ length: columns }).map((_, j) => (
            <div key={j} className="h-8 bg-gray-100 rounded flex-1"></div>
          ))}
        </div>
      ))}
    </div>
  )
}

export function LoadingPage() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="h-8 bg-gray-200 rounded w-48"></div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-32 bg-gray-200 rounded-xl"></div>
        ))}
      </div>
      <div className="h-64 bg-gray-200 rounded-xl"></div>
    </div>
  )
}
