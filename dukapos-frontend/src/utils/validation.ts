export interface ValidationRule {
  required?: boolean
  minLength?: number
  maxLength?: number
  min?: number
  max?: number
  pattern?: RegExp
  email?: boolean
  phone?: boolean
  numeric?: boolean
  decimal?: boolean
  custom?: (value: string) => boolean | string
}

export interface ValidationError {
  field: string
  message: string
}

export interface ValidationResult {
  isValid: boolean
  errors: ValidationError[]
}

export function validate(value: string, rules: ValidationRule, fieldName: string): ValidationError | null {
  // Required check
  if (rules.required && (!value || value.trim() === '')) {
    return { field: fieldName, message: `${fieldName} is required` }
  }

  // Skip other checks if empty and not required
  if (!value || value.trim() === '') {
    return null
  }

  // Min length
  if (rules.minLength && value.length < rules.minLength) {
    return { field: fieldName, message: `${fieldName} must be at least ${rules.minLength} characters` }
  }

  // Max length
  if (rules.maxLength && value.length > rules.maxLength) {
    return { field: fieldName, message: `${fieldName} must be at most ${rules.maxLength} characters` }
  }

  // Min value (for numbers)
  if (rules.min !== undefined) {
    const num = parseFloat(value)
    if (isNaN(num) || num < rules.min) {
      return { field: fieldName, message: `${fieldName} must be at least ${rules.min}` }
    }
  }

  // Max value (for numbers)
  if (rules.max !== undefined) {
    const num = parseFloat(value)
    if (isNaN(num) || num > rules.max) {
      return { field: fieldName, message: `${fieldName} must be at most ${rules.max}` }
    }
  }

  // Email
  if (rules.email) {
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailPattern.test(value)) {
      return { field: fieldName, message: 'Invalid email address' }
    }
  }

  // Phone (Kenyan format)
  if (rules.phone) {
    const phonePattern = /^(\+254|254|0)[1-9]\d{8}$/
    if (!phonePattern.test(value)) {
      return { field: fieldName, message: 'Invalid phone number' }
    }
  }

  // Numeric
  if (rules.numeric && !/^\d+$/.test(value)) {
    return { field: fieldName, message: `${fieldName} must be a number` }
  }

  // Decimal
  if (rules.decimal && !/^\d+(\.\d+)?$/.test(value)) {
    return { field: fieldName, message: `${fieldName} must be a decimal number` }
  }

  // Custom validation
  if (rules.custom) {
    const result = rules.custom(value)
    if (result !== true) {
      return { 
        field: fieldName, 
        message: typeof result === 'string' ? result : `${fieldName} is invalid` 
      }
    }
  }

  return null
}

export function validateForm(
  data: Record<string, string>,
  schema: Record<string, ValidationRule>
): ValidationResult {
  const errors: ValidationError[] = []

  for (const [fieldName, rules] of Object.entries(schema)) {
    const value = data[fieldName] || ''
    const error = validate(value, rules, fieldName)
    if (error) {
      errors.push(error)
    }
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

// Common validation schemas
export const loginSchema = {
  phone: { required: true, phone: true },
  password: { required: true, minLength: 6 }
}

export const registerSchema = {
  name: { required: true, minLength: 2, maxLength: 50 },
  phone: { required: true, phone: true },
  email: { required: true, email: true },
  password: { required: true, minLength: 6, maxLength: 20 }
}

export const productSchema = {
  name: { required: true, minLength: 2, maxLength: 100 },
  category: { maxLength: 50 },
  cost_price: { required: true, numeric: true, min: 0 },
  selling_price: { required: true, numeric: true, min: 0 },
  current_stock: { required: true, numeric: true, min: 0 },
  low_stock_threshold: { required: true, numeric: true, min: 0 }
}

export const customerSchema = {
  name: { required: true, minLength: 2, maxLength: 100 },
  phone: { required: true, phone: true },
  email: { email: true }
}

export const supplierSchema = {
  name: { required: true, minLength: 2, maxLength: 100 },
  phone: { phone: true },
  email: { email: true }
}

// Simple validation helpers for quick checks
export function validateRequired(value: string | null | undefined): true | string {
  if (!value || value.trim() === '') {
    return 'This field is required'
  }
  return true
}

export function validateMinLength(value: string, min: number): true | string {
  if (value.length < min) {
    return `Must be at least ${min} characters`
  }
  return true
}

export function validateMaxLength(value: string, max: number): true | string {
  if (value.length > max) {
    return `Must be at most ${max} characters`
  }
  return true
}

export function validatePositiveNumber(value: number): true | string {
  if (value < 0) {
    return 'Must be a positive number'
  }
  return true
}

export function validateEmail(value: string): true | string {
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailPattern.test(value)) {
    return 'Invalid email address'
  }
  return true
}

export function validatePhone(value: string): true | string {
  const phonePattern = /^(\+254|254|0)[1-9]\d{8}$/
  if (!phonePattern.test(value)) {
    return 'Invalid phone number'
  }
  return true
}

export function validateKenyanPhone(value: string): true | string {
  const patterns = [
    /^\+254[1-9]\d{8}$/,
    /^254[1-9]\d{8}$/,
    /^0[1-9]\d{8}$/
  ]
  
  if (!patterns.some(p => p.test(value))) {
    return 'Invalid Kenyan phone number'
  }
  return true
}
