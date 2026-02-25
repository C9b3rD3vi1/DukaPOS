import { Link } from 'react-router-dom'

interface FABProps {
  to?: string
  onClick?: () => void
  icon?: React.ReactNode
}

export function FAB({ to, onClick, icon }: FABProps) {
  const content = (
    <button
      onClick={onClick}
      className="fixed bottom-24 md:bottom-8 right-6 z-50 w-14 h-14 md:w-16 md:h-16 bg-gradient-to-r from-primary to-primary-700 text-white rounded-2xl shadow-lg shadow-primary/30 hover:shadow-xl hover:shadow-primary/40 hover:scale-110 active:scale-95 transition-all flex items-center justify-center"
    >
      {icon || (
        <svg className="w-6 h-6 md:w-7 md:h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 4v16m8-8H4" />
        </svg>
      )}
    </button>
  )

  if (to) {
    return <Link to={to}>{content}</Link>
  }

  return content
}
