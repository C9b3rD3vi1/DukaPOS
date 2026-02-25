import { useMemo } from 'react'

interface ChartDataPoint {
  label: string
  value: number
  color?: string
}

interface ChartProps {
  data: ChartDataPoint[]
  type: 'bar' | 'line' | 'pie' | 'doughnut'
  height?: number
  showLabels?: boolean
  showValues?: boolean
  title?: string
  colors?: string[]
}

const defaultColors = [
  '#00A650', '#3B82F6', '#F59E0B', '#EF4444', '#8B5CF6', 
  '#EC4899', '#14B8A6', '#F97316', '#6366F1', '#84CC16'
]

export function Chart({
  data,
  type,
  height = 200,
  showLabels = true,
  showValues = true,
  title,
  colors = defaultColors
}: ChartProps) {
  const maxValue = useMemo(() => Math.max(...data.map(d => d.value), 1), [data])
  
  const total = useMemo(() => data.reduce((sum, d) => sum + d.value, 0), [data])

  const formatValue = (value: number) => {
    if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
    if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
    return value.toString()
  }

  const renderBarChart = () => (
    <div className="flex items-end justify-between gap-2 h-full px-2">
      {data.map((item, index) => (
        <div key={index} className="flex-1 flex flex-col items-center">
          {showValues && (
            <span className="text-xs text-gray-500 mb-1">{formatValue(item.value)}</span>
          )}
          <div 
            className="w-full rounded-t-md transition-all hover:opacity-80"
            style={{ 
              height: `${(item.value / maxValue) * (height - 40)}px`,
              backgroundColor: item.color || colors[index % colors.length]
            }}
          />
          {showLabels && (
            <span className="text-xs text-gray-600 mt-2 truncate w-full text-center">
              {item.label}
            </span>
          )}
        </div>
      ))}
    </div>
  )

  const renderLineChart = () => {
    const points = data.map((item, index) => {
      const x = (index / (data.length - 1 || 1)) * 100
      const y = 100 - (item.value / maxValue) * 80
      return { x, y, ...item }
    })

    const pathD = points.map((p, i) => 
      i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`
    ).join(' ')

    return (
      <div className="relative h-full w-full">
        <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-full">
          {/* Grid lines */}
          {[0, 25, 50, 75, 100].map(y => (
            <line 
              key={y} 
              x1="0" y1={y} x2="100" y2={y} 
              stroke="#E5E7EB" 
              strokeWidth="0.5" 
              strokeDasharray="2,2" 
            />
          ))}
          
          {/* Line */}
          <path 
            d={pathD} 
            fill="none" 
            stroke={colors[0]} 
            strokeWidth="2" 
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          
          {/* Points */}
          {points.map((p, i) => (
            <circle 
              key={i} 
              cx={p.x} 
              cy={p.y} 
              r="2" 
              fill={colors[0]}
            />
          ))}
        </svg>
        
        {/* Labels */}
        {showLabels && (
          <div className="flex justify-between px-2 mt-2">
            {data.map((item, index) => (
              <span key={index} className="text-xs text-gray-600 truncate" style={{ maxWidth: '60px' }}>
                {item.label}
              </span>
            ))}
          </div>
        )}
      </div>
    )
  }

  const renderPieChart = () => {
    let currentAngle = 0
    
    const slices = data.map((item, index) => {
      const percentage = item.value / total
      const angle = percentage * 360
      const startAngle = currentAngle
      currentAngle += angle
      
      const startRad = (startAngle - 90) * Math.PI / 180
      const endRad = (startAngle + angle - 90) * Math.PI / 180
      
      const x1 = 50 + 40 * Math.cos(startRad)
      const y1 = 50 + 40 * Math.sin(startRad)
      const x2 = 50 + 40 * Math.cos(endRad)
      const y2 = 50 + 40 * Math.sin(endRad)
      
      const largeArc = angle > 180 ? 1 : 0
      
      return {
        path: `M 50 50 L ${x1} ${y1} A 40 40 0 ${largeArc} 1 ${x2} ${y2} Z`,
        color: item.color || colors[index % colors.length],
        percentage,
        ...item
      }
    })

    return (
      <div className="flex items-center gap-4">
        <svg viewBox="0 0 100 100" className="w-32 h-32">
          {slices.map((slice, i) => (
            <path 
              key={i} 
              d={slice.path} 
              fill={slice.color}
              className="hover:opacity-80 transition-opacity"
            />
          ))}
        </svg>
        
        {/* Legend */}
        <div className="flex-1 space-y-1">
          {data.map((item, index) => (
            <div key={index} className="flex items-center gap-2">
              <div 
                className="w-3 h-3 rounded-sm" 
                style={{ backgroundColor: item.color || colors[index % colors.length] }}
              />
              <span className="text-xs text-gray-600 flex-1 truncate">{item.label}</span>
              {showValues && (
                <span className="text-xs font-medium">{formatValue(item.value)}</span>
              )}
            </div>
          ))}
        </div>
      </div>
    )
  }

  const renderChart = () => {
    switch (type) {
      case 'bar': return renderBarChart()
      case 'line': return renderLineChart()
      case 'pie': 
      case 'doughnut': return renderPieChart()
      default: return renderBarChart()
    }
  }

  return (
    <div className="w-full">
      {title && (
        <h3 className="text-sm font-medium text-gray-700 mb-4">{title}</h3>
      )}
      <div style={{ height: `${height}px` }}>
        {data.length > 0 ? (
          renderChart()
        ) : (
          <div className="flex items-center justify-center h-full text-gray-400">
            No data available
          </div>
        )}
      </div>
    </div>
  )
}

// Helper to transform API data to chart format
export function transformSalesToChartData(
  sales: Array<{ date: string; total: number }>
): Array<{ label: string; value: number }> {
  return sales.map(sale => ({
    label: new Date(sale.date).toLocaleDateString('en-KE', {
      month: 'short',
      day: 'numeric'
    }),
    value: sale.total
  }))
}

export function transformTopProductsToChartData(
  products: Array<{ name: string; quantity: number; revenue: number }>
): Array<{ label: string; value: number }> {
  return products.map(product => ({
    label: product.name.length > 15 ? product.name.slice(0, 12) + '...' : product.name,
    value: product.revenue
  }))
}

export function transformCategorySalesToChartData(
  categories: Array<{ category: string; total: number }>
): Array<{ label: string; value: number }> {
  return categories.map(cat => ({
    label: cat.category || 'Uncategorized',
    value: cat.total
  }))
}
