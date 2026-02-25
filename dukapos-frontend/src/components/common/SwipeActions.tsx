import { useState, useRef, useCallback } from 'react'

interface SwipeAction {
  label: string
  icon: React.ReactNode
  onClick: () => void
  color: string
}

interface SwipeActionsProps {
  children: React.ReactNode
  leftActions?: SwipeAction[]
  rightActions?: SwipeAction[]
}

export function SwipeActions({ 
  children, 
  leftActions = [], 
  rightActions = []
}: SwipeActionsProps) {
  const [translateX, setTranslateX] = useState(0)
  const startX = useRef(0)
  const isSwiping = useRef(false)

  const handleSwipeOpen = (direction: 'left' | 'right') => {
    const actions = direction === 'left' ? leftActions : rightActions
    const width = actions.length * 80
    setTranslateX(direction === 'left' ? width : -width)
  }

  const handleSwipeClose = () => {
    setTranslateX(0)
  }

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    startX.current = e.touches[0].clientX
    isSwiping.current = true
  }, [])

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isSwiping.current) return
    
    const currentX = e.touches[0].clientX
    const diff = currentX - startX.current
    
    if (leftActions.length > 0 && diff > 0) {
      setTranslateX(Math.min(diff * 0.5, leftActions.length * 80))
    } else if (rightActions.length > 0 && diff < 0) {
      setTranslateX(Math.max(diff * 0.5, -rightActions.length * 80))
    }
  }, [leftActions, rightActions])

  const handleTouchEnd = useCallback(() => {
    isSwiping.current = false
    
    const threshold = 50
    if (translateX > threshold && leftActions.length > 0) {
      handleSwipeOpen('left')
    } else if (translateX < -threshold && rightActions.length > 0) {
      handleSwipeOpen('right')
    } else {
      handleSwipeClose()
    }
  }, [translateX, leftActions, rightActions])

  return (
    <div 
      className="relative overflow-hidden"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {/* Left Actions */}
      {leftActions.length > 0 && (
        <div 
          className="absolute inset-y-0 left-0 flex"
          style={{ transform: `translateX(${translateX > 0 ? 0 : -100}%)` }}
        >
          {leftActions.map((action, index) => (
            <button
              key={index}
              onClick={(e) => {
                e.stopPropagation()
                action.onClick()
                handleSwipeClose()
              }}
              className={`${action.color} px-4 h-full flex flex-col items-center justify-center gap-1 min-w-[80px] text-white`}
            >
              {action.icon}
              <span className="text-xs font-medium">{action.label}</span>
            </button>
          ))}
        </div>
      )}

      {/* Right Actions */}
      {rightActions.length > 0 && (
        <div 
          className="absolute inset-y-0 right-0 flex"
          style={{ transform: `translateX(${translateX < 0 ? 0 : 100}%)` }}
        >
          {rightActions.map((action, index) => (
            <button
              key={index}
              onClick={(e) => {
                e.stopPropagation()
                action.onClick()
                handleSwipeClose()
              }}
              className={`${action.color} px-4 h-full flex flex-col items-center justify-center gap-1 min-w-[80px] text-white`}
            >
              {action.icon}
              <span className="text-xs font-medium">{action.label}</span>
            </button>
          ))}
        </div>
      )}

      {/* Content */}
      <div
        style={{ transform: `translateX(${translateX}px)` }}
        className="bg-white transition-transform"
      >
        {children}
      </div>
    </div>
  )
}
