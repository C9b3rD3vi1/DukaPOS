interface SkeletonProps {
  className?: string
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded'
  width?: string | number
  height?: string | number
  animation?: 'pulse' | 'wave' | 'none'
}

export function Skeleton({ 
  className = '', 
  variant = 'rectangular',
  width,
  height,
  animation = 'pulse'
}: SkeletonProps) {
  const variants = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-none',
    rounded: 'rounded-xl',
  }

  const animations = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer bg-gradient-to-r from-surface-200 via-surface-100 to-surface-200 bg-[length:200%_100%]',
    none: '',
  }

  return (
    <div
      className={`
        bg-surface-200 
        ${variants[variant]} 
        ${animations[animation]}
        ${className}
      `}
      style={{
        width: width,
        height: height,
      }}
    />
  )
}

export function SkeletonCard() {
  return (
    <div className="bg-white rounded-2xl p-4 border border-surface-100 shadow-card">
      <Skeleton variant="rounded" className="aspect-square mb-4" />
      <Skeleton variant="text" className="h-4 w-3/4 mb-2" />
      <Skeleton variant="text" className="h-3 w-1/2 mb-3" />
      <Skeleton variant="rounded" className="h-6 w-full" />
    </div>
  )
}

export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="bg-white rounded-2xl border border-surface-100 shadow-card overflow-hidden">
      <div className="bg-surface-50 border-b border-surface-100">
        <div className="flex items-center gap-4 p-4">
          <Skeleton variant="text" className="h-4 flex-1" />
          <Skeleton variant="text" className="h-4 w-24" />
          <Skeleton variant="text" className="h-4 w-24" />
          <Skeleton variant="text" className="h-4 w-24" />
        </div>
      </div>
      <div className="divide-y divide-surface-100">
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 p-4">
            <Skeleton variant="circular" className="w-10 h-10" />
            <Skeleton variant="text" className="h-4 flex-1" />
            <Skeleton variant="rounded" className="h-6 w-20" />
            <Skeleton variant="text" className="h-4 w-24" />
          </div>
        ))}
      </div>
    </div>
  )
}

export function SkeletonList({ items = 5 }: { items?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 p-4 bg-white rounded-xl border border-surface-100">
          <Skeleton variant="circular" className="w-12 h-12" />
          <div className="flex-1">
            <Skeleton variant="text" className="h-4 w-1/3 mb-2" />
            <Skeleton variant="text" className="h-3 w-1/4" />
          </div>
          <Skeleton variant="rounded" className="h-8 w-20" />
        </div>
      ))}
    </div>
  )
}

export function SkeletonGrid({ columns = 4, items = 8 }: { columns?: number, items?: number }) {
  const gridCols = {
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-4',
    5: 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-5',
    6: 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-6',
  }

  return (
    <div className={`grid ${gridCols[columns as keyof typeof gridCols] || gridCols[4]} gap-4`}>
      {Array.from({ length: items }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  )
}
