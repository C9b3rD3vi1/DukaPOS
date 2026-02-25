import { describe, it, expect } from 'vitest'
import { 
  validatePhone, 
  validateEmail, 
  validateRequired, 
  validateMinLength, 
  validateMaxLength,
  validatePositiveNumber,
  validateKenyanPhone
} from '@/utils/validation'

describe('Validation Utils', () => {
  describe('validateRequired', () => {
    it('should fail for empty values', () => {
      expect(validateRequired('')).toBe('This field is required')
      expect(validateRequired(null)).toBe('This field is required')
      expect(validateRequired(undefined)).toBe('This field is required')
    })

    it('should pass for non-empty values', () => {
      expect(validateRequired('hello')).toBe(true)
      expect(validateRequired('0')).toBe(true)
    })
  })

  describe('validateMinLength', () => {
    it('should fail for short strings', () => {
      expect(validateMinLength('ab', 3)).toBe('Must be at least 3 characters')
      expect(validateMinLength('', 1)).toBe('Must be at least 1 characters')
    })

    it('should pass for strings of minimum length', () => {
      expect(validateMinLength('abc', 3)).toBe(true)
      expect(validateMinLength('hello', 3)).toBe(true)
    })
  })

  describe('validateMaxLength', () => {
    it('should fail for long strings', () => {
      expect(validateMaxLength('abcdef', 5)).toBe('Must be at most 5 characters')
    })

    it('should pass for strings within limit', () => {
      expect(validateMaxLength('abc', 5)).toBe(true)
      expect(validateMaxLength('', 5)).toBe(true)
    })
  })

  describe('validatePositiveNumber', () => {
    it('should fail for negative numbers', () => {
      expect(validatePositiveNumber(-1)).toBe('Must be a positive number')
      expect(validatePositiveNumber(-0.01)).toBe('Must be a positive number')
    })

    it('should pass for positive numbers', () => {
      expect(validatePositiveNumber(0)).toBe(true)
      expect(validatePositiveNumber(1)).toBe(true)
      expect(validatePositiveNumber(100.50)).toBe(true)
    })
  })

  describe('validateEmail', () => {
    it('should fail for invalid emails', () => {
      expect(validateEmail('invalid')).toBe('Invalid email address')
      expect(validateEmail('invalid@')).toBe('Invalid email address')
      expect(validateEmail('@domain.com')).toBe('Invalid email address')
    })

    it('should pass for valid emails', () => {
      expect(validateEmail('test@example.com')).toBe(true)
      expect(validateEmail('user.name@domain.co.ke')).toBe(true)
    })
  })

  describe('validatePhone', () => {
    it('should fail for invalid phone numbers', () => {
      expect(validatePhone('123')).toBe('Invalid phone number')
      expect(validatePhone('abc')).toBe('Invalid phone number')
    })

    it('should pass for valid phone numbers', () => {
      expect(validatePhone('254712345678')).toBe(true)
      expect(validatePhone('0712345678')).toBe(true)
    })
  })

  describe('validateKenyanPhone', () => {
    it('should fail for non-Kenyan phone numbers', () => {
      expect(validateKenyanPhone('+254712345678')).toBe(true)
      expect(validateKenyanPhone('254712345678')).toBe(true)
      expect(validateKenyanPhone('0712345678')).toBe(true)
      expect(validateKenyanPhone('+254712345678')).toBe(true)
    })

    it('should pass for valid Kenyan phone numbers', () => {
      expect(validateKenyanPhone('254700000000')).toBe(true)
      expect(validateKenyanPhone('254800000000')).toBe(true)
    })
  })
})
