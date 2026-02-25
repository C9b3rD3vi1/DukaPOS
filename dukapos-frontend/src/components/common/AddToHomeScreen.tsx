import { useAddToHomeScreen } from '@/hooks/usePWA'

interface AddToHomeScreenProps {
  children?: React.ReactNode
}

export function AddToHomeScreenPrompt({ children }: AddToHomeScreenProps) {
  const { shouldShowPrompt, install, dismiss } = useAddToHomeScreen()

  if (!shouldShowPrompt) {
    return children || null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center p-4 sm:items-center">
      <div 
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={dismiss}
      />
      
      <div className="relative bg-white rounded-2xl shadow-2xl p-6 max-w-sm w-full animate-slide-up">
        <button
          onClick={dismiss}
          className="absolute top-3 right-3 p-1 text-gray-400 hover:text-gray-600 rounded-full hover:bg-gray-100"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        <div className="flex items-center gap-4 mb-4">
          <div className="w-14 h-14 bg-primary/10 rounded-2xl flex items-center justify-center">
            <svg className="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">Install DukaPOS</h3>
            <p className="text-sm text-gray-500">Add to home screen for quick access</p>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={dismiss}
            className="flex-1 px-4 py-2.5 text-gray-700 font-medium rounded-xl border border-gray-200 hover:bg-gray-50 transition"
          >
            Not now
          </button>
          <button
            onClick={install}
            className="flex-1 px-4 py-2.5 bg-primary text-white font-medium rounded-xl hover:bg-primary/90 transition"
          >
            Install
          </button>
        </div>
      </div>
    </div>
  )
}

export function InstallButton() {
  const { canInstall, isInstalled, install } = useAddToHomeScreen()

  if (isInstalled) {
    return (
      <span className="flex items-center gap-2 text-green-600 text-sm">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
        Installed
      </span>
    )
  }

  if (!canInstall) {
    return null
  }

  return (
    <button
      onClick={install}
      className="flex items-center gap-2 px-3 py-1.5 bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition text-sm font-medium"
    >
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
      </svg>
      Install App
    </button>
  )
}

export default AddToHomeScreenPrompt
