import React from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger' | 'success' | 'warning'
  size?: 'sm' | 'md' | 'lg' | 'xl'
  isLoading?: boolean
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
  fullWidth?: boolean
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  isLoading = false,
  leftIcon,
  rightIcon,
  fullWidth = false,
  className = '',
  disabled,
  ...props
}: ButtonProps) {
  const baseStyles = `
    inline-flex items-center justify-center font-semibold 
    transition-all duration-200 ease-smooth
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed
    active:scale-[0.98] transform
  `

  const variants = {
    primary: `
      bg-primary text-white hover:bg-primary-dark
      focus:ring-primary/50 shadow-button hover:shadow-button-hover
      hover:shadow-glow
    `,
    secondary: `
      bg-surface-100 text-surface-700 hover:bg-surface-200
      focus:ring-surface-300
    `,
    outline: `
      border-2 border-surface-200 text-surface-700 hover:border-primary hover:text-primary
      focus:ring-primary/30 bg-transparent
    `,
    ghost: `
      text-surface-600 hover:bg-surface-100 hover:text-surface-900
      focus:ring-surface-200 bg-transparent
    `,
    danger: `
      bg-danger text-white hover:bg-red-600
      focus:ring-danger/50 shadow-button hover:shadow-button-hover hover:shadow-glow-danger
    `,
    success: `
      bg-success text-white hover:bg-emerald-600
      focus:ring-success/50 shadow-button hover:shadow-button-hover
    `,
    warning: `
      bg-warning text-white hover:bg-amber-600
      focus:ring-warning/50 shadow-button hover:shadow-button-hover hover:shadow-glow-warning
    `,
  }

  const sizes = {
    sm: 'px-3 py-1.5 text-sm rounded-lg gap-1.5',
    md: 'px-4 py-2.5 text-sm rounded-xl gap-2',
    lg: 'px-6 py-3 text-base rounded-xl gap-2.5',
    xl: 'px-8 py-4 text-lg rounded-2xl gap-3',
  }

  return (
    <button
      className={`
        ${baseStyles}
        ${variants[variant]}
        ${sizes[size]}
        ${fullWidth ? 'w-full' : ''}
        ${isLoading ? 'cursor-wait' : ''}
        ${className}
      `}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? (
        <>
          <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
              fill="none"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          <span>Loading...</span>
        </>
      ) : (
        <>
          {leftIcon && <span className="shrink-0">{leftIcon}</span>}
          <span>{children}</span>
          {rightIcon && <span className="shrink-0">{rightIcon}</span>}
        </>
      )}
    </button>
  )
}

interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  isLoading?: boolean
  icon: React.ReactNode
  label: string
}

export function IconButton({
  variant = 'ghost',
  size = 'md',
  isLoading = false,
  icon,
  label,
  className = '',
  disabled,
  ...props
}: IconButtonProps) {
  const baseStyles = `
    inline-flex items-center justify-center rounded-full
    transition-all duration-200 ease-smooth
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed
    active:scale-90
  `

  const variants = {
    primary: 'bg-primary text-white hover:bg-primary-dark focus:ring-primary/50',
    secondary: 'bg-surface-100 text-surface-600 hover:bg-surface-200 focus:ring-surface-300',
    outline: 'border border-surface-200 text-surface-600 hover:border-primary hover:text-primary focus:ring-primary/30',
    ghost: 'text-surface-500 hover:bg-surface-100 hover:text-surface-700 focus:ring-surface-200',
    danger: 'text-danger hover:bg-red-50 focus:ring-danger/50',
  }

  const sizes = {
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-12 h-12',
  }

  return (
    <button
      className={`${baseStyles} ${variants[variant]} ${sizes[size]} ${className}`}
      disabled={disabled || isLoading}
      aria-label={label}
      title={label}
      {...props}
    >
      {isLoading ? (
        <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      ) : (
        icon
      )}
    </button>
  )
}

export default Button
