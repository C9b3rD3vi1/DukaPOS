import { useState, useEffect, useCallback, createContext, useContext } from 'react'

type TranslationKey = string
type Translations = Record<string, unknown>

interface I18nContextType {
  locale: string
  setLocale: (locale: string) => void
  t: (key: TranslationKey, params?: Record<string, string>) => string
  availableLocales: { code: string; name: string }[]
}

const translations: Record<string, Translations> = {}

const availableLocales = [
  { code: 'en', name: 'English' },
  { code: 'sw', name: 'Kiswahili' }
]

async function loadTranslations(locale: string): Promise<Translations> {
  if (translations[locale]) {
    return translations[locale]
  }
  
  try {
    const response = await fetch(`/locales/${locale}.json`)
    const data = await response.json()
    translations[locale] = data
    return data
  } catch (error) {
    console.error(`Failed to load translations for ${locale}:`, error)
    if (locale !== 'en') {
      return loadTranslations('en')
    }
    return {}
  }
}

const I18nContext = createContext<I18nContextType | null>(null)

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState('en')
  const [translations, setTranslations] = useState<Translations>({})

  useEffect(() => {
    const initLocale = async () => {
      const savedLocale = localStorage.getItem('locale') || 'en'
      setLocaleState(savedLocale)
      await loadTranslations(savedLocale).then(setTranslations)
    }
    
    initLocale()
  }, [])

  const setLocale = useCallback(async (newLocale: string) => {
    setLocaleState(newLocale)
    localStorage.setItem('locale', newLocale)
    await loadTranslations(newLocale).then(setTranslations)
  }, [])

  const t = useCallback((key: TranslationKey, params?: Record<string, string>): string => {
    const keys = key.split('.')
    let value: unknown = translations
    
    for (const k of keys) {
      if (value && typeof value === 'object') {
        value = (value as Record<string, unknown>)[k]
      } else {
        return key
      }
    }
    
    if (typeof value !== 'string') {
      return key
    }
    
    if (params) {
      return Object.entries(params).reduce(
        (str, [paramKey, paramValue]) => str.replace(`{${paramKey}}`, paramValue),
        value
      )
    }
    
    return value
  }, [translations])

  return (
    <I18nContext.Provider value={{ locale, setLocale, t, availableLocales }}>
      {children}
    </I18nContext.Provider>
  )
}

export function useI18n() {
  const context = useContext(I18nContext)
  if (!context) {
    throw new Error('useI18n must be used within I18nProvider')
  }
  return context
}

export function useTranslation() {
  const { t, locale, setLocale, availableLocales } = useI18n()
  return { t, locale, setLocale, availableLocales }
}
