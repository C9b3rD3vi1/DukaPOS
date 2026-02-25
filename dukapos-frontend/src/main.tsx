import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'

function setupErrorHandling() {
  if (typeof window !== 'undefined') {
    window.addEventListener('error', (event) => {
      console.error('Global error:', event.error)
    })

    window.addEventListener('unhandledrejection', (event) => {
      console.error('Unhandled promise rejection:', event.reason)
    })
  }
}

function setupOnlineStatus() {
  if (typeof window !== 'undefined') {
    window.addEventListener('online', () => {
      console.log('Application is online')
      document.body.classList.remove('offline')
      document.body.classList.add('online')
    })

    window.addEventListener('offline', () => {
      console.log('Application is offline')
      document.body.classList.remove('online')
      document.body.classList.add('offline')
    })
  }
}

function bootstrap() {
  setupErrorHandling()
  setupOnlineStatus()

  const container = document.getElementById('root')
  
  if (container) {
    const root = createRoot(container)
    root.render(
      <StrictMode>
        <App />
      </StrictMode>
    )
  } else {
    console.error('Root element not found')
  }
}

bootstrap()
