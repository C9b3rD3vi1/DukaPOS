type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogEntry {
  level: LogLevel
  message: string
  timestamp: Date
  context?: Record<string, unknown>
  error?: Error
}

interface LoggerOptions {
  enableConsole?: boolean
  minLevel?: LogLevel
  prefix?: string
}

const logLevels: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3
}

class Logger {
  private enableConsole = true
  private minLevel: LogLevel = 'info'
  private prefix = '[DukaPOS]'
  private logs: LogEntry[] = []
  private maxLogs = 100

  constructor(options: LoggerOptions = {}) {
    this.enableConsole = options.enableConsole ?? true
    this.minLevel = options.minLevel ?? 'info'
    this.prefix = options.prefix ?? '[DukaPOS]'
  }

  private shouldLog(level: LogLevel): boolean {
    return logLevels[level] >= logLevels[this.minLevel]
  }

  private formatMessage(level: LogLevel, message: string): string {
    const timestamp = new Date().toISOString()
    return `${timestamp} ${this.prefix} [${level.toUpperCase()}] ${message}`
  }

  private addLog(entry: LogEntry): void {
    this.logs.push(entry)
    if (this.logs.length > this.maxLogs) {
      this.logs.shift()
    }
  }

  debug(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('debug')) return
    
    const entry: LogEntry = {
      level: 'debug',
      message,
      timestamp: new Date(),
      context
    }
    this.addLog(entry)
    
    if (this.enableConsole) {
      console.debug(this.formatMessage('debug', message), context || '')
    }
  }

  info(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('info')) return
    
    const entry: LogEntry = {
      level: 'info',
      message,
      timestamp: new Date(),
      context
    }
    this.addLog(entry)
    
    if (this.enableConsole) {
      console.info(this.formatMessage('info', message), context || '')
    }
  }

  warn(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('warn')) return
    
    const entry: LogEntry = {
      level: 'warn',
      message,
      timestamp: new Date(),
      context
    }
    this.addLog(entry)
    
    if (this.enableConsole) {
      console.warn(this.formatMessage('warn', message), context || '')
    }
  }

  error(message: string, error?: Error, context?: Record<string, unknown>): void {
    if (!this.shouldLog('error')) return
    
    const entry: LogEntry = {
      level: 'error',
      message,
      timestamp: new Date(),
      error,
      context
    }
    this.addLog(entry)
    
    if (this.enableConsole) {
      console.error(this.formatMessage('error', message), error || '', context || '')
    }
  }

  getLogs(level?: LogLevel): LogEntry[] {
    if (level) {
      return this.logs.filter(log => log.level === level)
    }
    return [...this.logs]
  }

  clearLogs(): void {
    this.logs = []
  }

  setLevel(level: LogLevel): void {
    this.minLevel = level
  }

  setEnableConsole(enable: boolean): void {
    this.enableConsole = enable
  }
}

export const logger = new Logger({
  enableConsole: true,
  minLevel: 'info',
  prefix: '[DukaPOS]'
})

// Convenience loggers for different contexts
export const authLogger = new Logger({ prefix: '[Auth]' })
export const apiLogger = new Logger({ prefix: '[API]' })
export const syncLogger = new Logger({ prefix: '[Sync]' })
export const storageLogger = new Logger({ prefix: '[Storage]' })

// Log uncaught errors
if (typeof window !== 'undefined') {
  window.addEventListener('error', (event) => {
    logger.error('Uncaught error', event.error, {
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno
    })
  })

  window.addEventListener('unhandledrejection', (event) => {
    logger.error('Unhandled promise rejection', event.reason)
  })
}
