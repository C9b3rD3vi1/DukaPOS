import { useState, useEffect, useCallback } from 'react'

interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[]
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed'
    platform: string
  }>
  prompt(): Promise<void>
}

interface PushNotificationConfig {
  onRegister?: (token: string) => void
  onNotification?: (notification: unknown) => void
  onError?: (error: unknown) => void
}

function isPlatform(platform: string): boolean {
  if (typeof window === 'undefined') return false
  const userAgent = navigator.userAgent.toLowerCase()
  if (platform === 'ios') return userAgent.includes('iphone') || userAgent.includes('ipad')
  if (platform === 'android') return userAgent.includes('android')
  return false
}

export function useAddToHomeScreen() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [isInstalled, setIsInstalled] = useState(false)
  const [canInstall, setCanInstall] = useState(false)

  useEffect(() => {
    const checkInstalled = () => {
      if (window.matchMedia('(display-mode: standalone)').matches) {
        setIsInstalled(true)
      }
    }

    checkInstalled()

    const handleBeforeInstall = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
      setCanInstall(true)
    }

    window.addEventListener('beforeinstallprompt', handleBeforeInstall)

    window.addEventListener('appinstalled', () => {
      setIsInstalled(true)
      setCanInstall(false)
      setDeferredPrompt(null)
    })

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstall)
    }
  }, [])

  const install = useCallback(async () => {
    if (!deferredPrompt) return false

    try {
      await deferredPrompt.prompt()
      const { outcome } = await deferredPrompt.userChoice
      
      if (outcome === 'accepted') {
        setIsInstalled(true)
        setCanInstall(false)
      }
      
      setDeferredPrompt(null)
      return outcome === 'accepted'
    } catch (error) {
      console.error('Install error:', error)
      return false
    }
  }, [deferredPrompt])

  const dismiss = useCallback(() => {
    setCanInstall(false)
    localStorage.setItem('pwa-install-dismissed', 'true')
  }, [])

  const shouldShowPrompt = canInstall && !isInstalled && 
    !localStorage.getItem('pwa-install-dismissed')

  return {
    canInstall,
    isInstalled,
    shouldShowPrompt,
    install,
    dismiss
  }
}

export function usePushNotifications(config?: PushNotificationConfig) {
  const [token, setToken] = useState<string | null>(null)
  const [permission, setPermission] = useState<NotificationPermission>('default')
  const [isSupported, setIsSupported] = useState(false)

  useEffect(() => {
    const checkSupport = async () => {
      if (typeof window === 'undefined' || !('Notification' in window)) {
        setIsSupported(false)
        return
      }

      if (isPlatform('ios') || isPlatform('android')) {
        try {
          const { PushNotifications } = await import('@capacitor/push-notifications')
          const result = await PushNotifications.requestPermissions()
          setIsSupported(result.receive === 'granted')
        } catch {
          setIsSupported(false)
        }
      } else {
        setIsSupported('Notification' in window)
      }
    }

    checkSupport()
  }, [])

  useEffect(() => {
    if (!isSupported) return

    const initPush = async () => {
      if (isPlatform('ios') || isPlatform('android')) {
        await initCapacitorPush(config)
      } else {
        await initWebPush(config)
      }
    }

    initPush()
  }, [isSupported])

  const initCapacitorPush = async (cfg?: PushNotificationConfig) => {
    try {
      const { PushNotifications } = await import('@capacitor/push-notifications')

      await PushNotifications.register()
      
      PushNotifications.addListener('registration', (event) => {
        setToken(event.value)
        cfg?.onRegister?.(event.value)
      })

      PushNotifications.addListener('registrationError', (event) => {
        cfg?.onError?.(event.error)
      })

      PushNotifications.addListener('pushNotificationReceived', (event) => {
        cfg?.onNotification?.(event)
      })

      PushNotifications.addListener('pushNotificationActionPerformed', (event) => {
        cfg?.onNotification?.(event)
      })
    } catch (error) {
      console.error('Push init error:', error)
      cfg?.onError?.(error)
    }
  }

  const initWebPush = async (cfg?: PushNotificationConfig) => {
    try {
      const perm = Notification.permission
      setPermission(perm)

      if (perm === 'granted') {
        const registration = await navigator.serviceWorker.ready
        const existingToken = await registration.pushManager.getSubscription()
        if (existingToken) {
          setToken(existingToken.endpoint)
          cfg?.onRegister?.(existingToken.endpoint)
        }
      }

      navigator.serviceWorker.addEventListener('push', (evt) => {
        const pushEvt = evt as unknown as { data?: { json: () => unknown } }
        if (pushEvt.data) {
          try {
            const data = pushEvt.data.json()
            cfg?.onNotification?.(data)
          } catch (e) {
            console.error('Failed to parse push data:', e)
          }
        }
      })
    } catch (error) {
      console.error('Web push init error:', error)
    }
  }

  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (isPlatform('ios') || isPlatform('android')) {
      try {
        const { PushNotifications } = await import('@capacitor/push-notifications')
        const result = await PushNotifications.requestPermissions()
        return result.receive === 'granted'
      } catch {
        return false
      }
    } else {
      const perm = await Notification.requestPermission()
      setPermission(perm)
      return perm === 'granted'
    }
  }, [])

  const showLocalNotification = useCallback(async (
    title: string,
    body: string,
    _options?: NotificationOptions
  ) => {
    if (isPlatform('ios') || isPlatform('android')) {
      console.log('Local notification:', title, body)
    } else if (permission === 'granted') {
      new Notification(title, { body })
    }
  }, [permission])

  return {
    token,
    permission,
    isSupported,
    requestPermission,
    showLocalNotification
  }
}
