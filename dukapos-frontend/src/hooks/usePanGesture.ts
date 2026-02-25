import { useRef, useCallback } from 'react'

interface PanInfo {
  offsetX: number
  offsetY: number
  velocityX: number
  velocityY: number
}

interface UsePanGestureOptions {
  onPanStart?: () => void
  onPanEnd?: (info: PanInfo) => void
  onPan?: (info: PanInfo) => void
}

export function usePanGesture({ 
  onPanStart, 
  onPanEnd, 
  onPan 
}: UsePanGestureOptions = {}) {
  const startX = useRef(0)
  const startY = useRef(0)
  const offsetX = useRef(0)
  const offsetY = useRef(0)
  const isPanning = useRef(false)

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    startX.current = e.touches[0].clientX
    startY.current = e.touches[0].clientY
    offsetX.current = 0
    offsetY.current = 0
    isPanning.current = true
    onPanStart?.()
  }, [onPanStart])

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isPanning.current) return
    
    const currentX = e.touches[0].clientX
    const currentY = e.touches[0].clientY
    
    offsetX.current = currentX - startX.current
    offsetY.current = currentY - startY.current
    
    onPan?.({
      offsetX: offsetX.current,
      offsetY: offsetY.current,
      velocityX: 0,
      velocityY: 0
    })
  }, [onPan])

  const handleTouchEnd = useCallback(() => {
    if (!isPanning.current) return
    
    isPanning.current = false
    onPanEnd?.({
      offsetX: offsetX.current,
      offsetY: offsetY.current,
      velocityX: 0,
      velocityY: 0
    })
  }, [onPanEnd])

  return {
    bind: {
      onTouchStart: handleTouchStart,
      onTouchMove: handleTouchMove,
      onTouchEnd: handleTouchEnd
    }
  }
}
