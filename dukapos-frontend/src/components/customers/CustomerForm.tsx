import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import type { Customer } from '@/api/types'

interface CustomerFormProps {
  customer?: Customer | null
  onSubmit: (data: CustomerFormData) => Promise<void>
  onCancel: () => void
  isLoading?: boolean
}

export interface CustomerFormData {
  name: string
  phone: string
  email: string
  loyalty_points: number
  total_purchases: number
}

const defaultFormData: CustomerFormData = {
  name: '',
  phone: '',
  email: '',
  loyalty_points: 0,
  total_purchases: 0
}

export function CustomerForm({ customer, onSubmit, onCancel, isLoading }: CustomerFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset
  } = useForm<CustomerFormData>({
    defaultValues: defaultFormData
  })

  useEffect(() => {
    if (customer) {
      reset({
        name: customer.name,
        phone: customer.phone,
        email: customer.email || '',
        loyalty_points: customer.loyalty_points,
        total_purchases: customer.total_purchases
      })
    } else {
      reset(defaultFormData)
    }
  }, [customer, reset])

  const formatPhoneNumber = (value: string) => {
    const cleaned = value.replace(/\D/g, '')
    if (cleaned.startsWith('254')) {
      return '+' + cleaned
    } else if (cleaned.startsWith('0')) {
      return cleaned
    } else if (cleaned.length > 0) {
      return '0' + cleaned
    }
    return value
  }

  const onFormSubmit = async (data: CustomerFormData) => {
    await onSubmit({ ...data, phone: formatPhoneNumber(data.phone) })
  }

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Customer Name *
        </label>
        <input
          type="text"
          {...register('name', { required: 'Name is required', minLength: { value: 2, message: 'Name must be at least 2 characters' } })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.name ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="Enter customer name"
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-500">{errors.name.message}</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Phone Number *
        </label>
        <input
          type="tel"
          {...register('phone', { 
            required: 'Phone number is required',
            pattern: {
              value: /^(\+254|254|0)[1-9]\d{8}$/,
              message: 'Invalid Kenyan phone number'
            }
          })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.phone ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="0712345678 or +254712345678"
        />
        {errors.phone && (
          <p className="mt-1 text-sm text-red-500">{errors.phone.message}</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Email Address
        </label>
        <input
          type="email"
          {...register('email', {
            pattern: {
              value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
              message: 'Invalid email address'
            }
          })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.email ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="customer@email.com"
        />
        {errors.email && (
          <p className="mt-1 text-sm text-red-500">{errors.email.message}</p>
        )}
      </div>

      {customer && (
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Loyalty Points
            </label>
            <input
              type="number"
              {...register('loyalty_points', { min: { value: 0, message: 'Must be positive' } })}
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition"
              min="0"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Total Purchases
            </label>
            <input
              type="number"
              {...register('total_purchases', { min: { value: 0, message: 'Must be positive' } })}
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition"
              min="0"
              step="0.01"
            />
          </div>
        </div>
      )}

      <div className="flex gap-3 pt-4">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 px-4 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50 transition-all"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isLoading}
          className="flex-1 px-4 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all disabled:opacity-50"
        >
          {isLoading ? 'Saving...' : customer ? 'Update' : 'Add'} Customer
        </button>
      </div>
    </form>
  )
}
