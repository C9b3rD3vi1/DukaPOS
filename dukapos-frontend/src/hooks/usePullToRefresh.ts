import { useState, useRef, useCallback, useEffect } from 'react'

interface UsePullToRefreshOptions {
  onRefresh: () => Promise<void>
  threshold?: number
  disabled?: boolean
}

export function usePullToRefresh({ 
  onRefresh, 
  threshold = 80,
  disabled = false 
}: UsePullToRefreshOptions) {
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [pullDistance, setPullDistance] = useState(0)
  const [isTouching, setIsTouching] = useState(false)
  const startY = useRef(0)

  const handleTouchStart = useCallback((e: TouchEvent) => {
    if (disabled || isRefreshing) return
    if (window.scrollY === 0) {
      setIsTouching(true)
      startY.current = e.touches[0].clientY
    }
  }, [disabled, isRefreshing])

  const handleTouchMove = useCallback((e: TouchEvent) => {
    if (!isTouching || disabled || isRefreshing) return
    if (window.scrollY > 0) {
      setIsTouching(false)
      setPullDistance(0)
      return
    }
    
    const currentY = e.touches[0].clientY
    const distance = Math.max(0, currentY - startY.current)
    setPullDistance(Math.min(distance * 0.5, threshold * 1.5))
  }, [isTouching, disabled, isRefreshing, threshold])

  const handleTouchEnd = useCallback(async () => {
    if (!isTouching || disabled) return
    
    setIsTouching(false)
    
    if (pullDistance >= threshold) {
      setIsRefreshing(true)
      try {
        await onRefresh()
      } finally {
        setIsRefreshing(false)
        setPullDistance(0)
      }
    } else {
      setPullDistance(0)
    }
  }, [isTouching, disabled, pullDistance, threshold, onRefresh])

  useEffect(() => {
    const options = { passive: true }
    
    window.addEventListener('touchstart', handleTouchStart, options)
    window.addEventListener('touchmove', handleTouchMove, options)
    window.addEventListener('touchend', handleTouchEnd, options)
    
    return () => {
      window.removeEventListener('touchstart', handleTouchStart)
      window.removeEventListener('touchmove', handleTouchMove)
      window.removeEventListener('touchend', handleTouchEnd)
    }
  }, [handleTouchStart, handleTouchMove, handleTouchEnd])

  const progress = Math.min(pullDistance / threshold, 1)
  
  return { isRefreshing, pullDistance, progress }
}
