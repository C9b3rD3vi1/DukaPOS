import { useEffect, useRef, useState } from 'react'

interface UseInViewOptions {
  threshold?: number
  rootMargin?: string
  triggerOnce?: boolean
}

export function useInView({
  threshold = 0.1,
  rootMargin = '0px',
  triggerOnce = true
}: UseInViewOptions = {}) {
  const ref = useRef<HTMLDivElement>(null)
  const [isInView, setIsInView] = useState(false)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true)
          if (triggerOnce) {
            observer.unobserve(element)
          }
        } else if (!triggerOnce) {
          setIsInView(false)
        }
      },
      { threshold, rootMargin }
    )

    observer.observe(element)

    return () => observer.unobserve(element)
  }, [threshold, rootMargin, triggerOnce])

  return { ref, isInView }
}

interface RevealOnScrollProps {
  children: React.ReactNode
  className?: string
  animation?: 'fade' | 'fade-up' | 'fade-down' | 'slide-left' | 'slide-right' | 'scale' | 'float'
  delay?: number
  duration?: number
}

export function RevealOnScroll({ 
  children, 
  className = '',
  animation = 'fade-up',
  delay = 0,
  duration = 500
}: RevealOnScrollProps) {
  const { ref, isInView } = useInView({ threshold: 0.1, triggerOnce: true })

  const animations: Record<string, string> = {
    fade: 'animate-fade-in',
    'fade-up': 'animate-fade-in-up',
    'fade-down': 'animate-fade-in-down',
    'slide-left': 'animate-slide-left',
    'slide-right': 'animate-slide-right',
    scale: 'animate-scale-in-up',
    float: 'animate-float',
  }

  return (
    <div
      ref={ref}
      className={`${className} ${isInView ? animations[animation] : 'opacity-0'}`}
      style={{
        animationDelay: `${delay}ms`,
        animationDuration: `${duration}ms`,
        animationFillMode: 'both',
      }}
    >
      {children}
    </div>
  )
}

interface StaggeredRevealProps {
  children: React.ReactNode
  className?: string
  staggerDelay?: number
  animation?: 'fade-up' | 'scale' | 'slide-left' | 'slide-right'
}

export function StaggeredReveal({
  children,
  className = '',
  staggerDelay = 50,
  animation = 'fade-up'
}: StaggeredRevealProps) {
  const childrenArray = Array.isArray(children) ? children : [children]
  const { ref, isInView } = useInView({ threshold: 0.1, triggerOnce: true })

  const animations: Record<string, string> = {
    'fade-up': 'animate-fade-in-up',
    scale: 'animate-scale-in-up',
    'slide-left': 'animate-slide-left',
    'slide-right': 'animate-slide-right',
  }

  return (
    <div ref={ref} className={className}>
      {childrenArray.map((child, index) => (
        <div
          key={index}
          className={isInView ? animations[animation] : 'opacity-0'}
          style={{
            animationDelay: `${index * staggerDelay}ms`,
            animationDuration: '400ms',
            animationFillMode: 'both',
          }}
        >
          {child}
        </div>
      ))}
    </div>
  )
}

export function AnimatedList({ 
  children, 
  className = '',
  animation = 'fade-up',
  stagger = 30 
}: { 
  children: React.ReactNode
  className?: string
  animation?: 'fade' | 'fade-up' | 'slide-left' | 'slide-right' | 'scale'
  stagger?: number
}) {
  const { ref, isInView } = useInView({ threshold: 0.05, triggerOnce: true })
  const childrenArray = Array.isArray(children) ? children : [children]

  const animations: Record<string, string> = {
    fade: 'animate-fade-in',
    'fade-up': 'animate-fade-in-up',
    'slide-left': 'animate-slide-left',
    'slide-right': 'animate-slide-right',
    scale: 'animate-scale-in-up',
  }

  return (
    <div ref={ref} className={className}>
      {childrenArray.map((child, i) => (
        <div
          key={i}
          className={isInView ? animations[animation] : 'opacity-0 translate-y-4'}
          style={{
            animationDelay: `${i * stagger}ms`,
            animationDuration: '300ms',
            animationFillMode: 'both',
          }}
        >
          {child}
        </div>
      ))}
    </div>
  )
}
