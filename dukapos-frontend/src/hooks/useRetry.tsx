import { useState, useCallback, useRef } from 'react'

interface RetryOptions {
  maxRetries?: number
  initialDelay?: number
  maxDelay?: number
  onRetry?: (attempt: number, error: unknown) => void
  shouldRetry?: (error: unknown) => boolean
}

interface RetryReturn {
  execute: <T>(fn: () => Promise<T>) => Promise<T>
  isRetrying: boolean
  attempt: number
  error: unknown | null
  reset: () => void
}

const DEFAULT_OPTIONS: Required<RetryOptions> = {
  maxRetries: 3,
  initialDelay: 1000,
  maxDelay: 10000,
  onRetry: () => {},
  shouldRetry: (error: unknown) => {
    if (!error) return false
    const errorStr = String(error).toLowerCase()
    const retryableErrors = ['network', 'timeout', 'econnrefused', 'etimedout', 'enotfound', '500', '502', '503', '504']
    return retryableErrors.some(e => errorStr.includes(e.toLowerCase()))
  }
}

export function useRetry(options: RetryOptions = {}): RetryReturn {
  const opts = { ...DEFAULT_OPTIONS, ...options }
  const [isRetrying, setIsRetrying] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [error, setError] = useState<unknown | null>(null)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const calculateDelay = useCallback((attemptNumber: number) => {
    const delay = opts.initialDelay * Math.pow(2, attemptNumber)
    const jitter = Math.random() * 500
    return Math.min(delay + jitter, opts.maxDelay)
  }, [opts.initialDelay, opts.maxDelay])

  const reset = useCallback(() => {
    setIsRetrying(false)
    setAttempt(0)
    setError(null)
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
  }, [])

  const execute = useCallback(async <T,>(fn: () => Promise<T>): Promise<T> => {
    let lastError: unknown
    
    for (let attemptNum = 0; attemptNum <= opts.maxRetries; attemptNum++) {
      try {
        setAttempt(attemptNum)
        setIsRetrying(attemptNum > 0)
        setError(null)
        
        const result = await fn()
        setIsRetrying(false)
        return result
      } catch (err) {
        lastError = err
        setError(err)
        
        if (attemptNum < opts.maxRetries && opts.shouldRetry(err)) {
          opts.onRetry(attemptNum + 1, err)
          
          await new Promise<void>((resolve) => {
            timeoutRef.current = setTimeout(() => {
              resolve()
            }, calculateDelay(attemptNum))
          })
        } else {
          throw err
        }
      }
    }
    
    throw lastError
  }, [opts.maxRetries, opts.shouldRetry, opts.onRetry, calculateDelay])

  return { execute, isRetrying, attempt, error, reset }
}

export function useAsyncRetry<T>(asyncFn: () => Promise<T>, options: RetryOptions = {}) {
  const retry = useRetry(options)
  
  const execute = useCallback(async () => {
    return retry.execute(asyncFn)
  }, [asyncFn, retry.execute])
  
  return {
    ...retry,
    execute
  }
}

interface RetryButtonProps {
  onRetry: () => void
  attempt: number
  maxRetries: number
  error?: string | null
  className?: string
}

export function RetryButton({ onRetry, attempt, maxRetries, error, className = '' }: RetryButtonProps) {
  const canRetry = attempt < maxRetries
  
  return (
    <div className={`flex flex-col items-center gap-2 ${className}`}>
      {error && (
        <p className="text-sm text-red-600 text-center">{error}</p>
      )}
      {canRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
        >
          Retry ({attempt}/{maxRetries})
        </button>
      )}
    </div>
  )
}
