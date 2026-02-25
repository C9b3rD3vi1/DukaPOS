import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import type { Supplier } from '@/api/types'

interface SupplierFormProps {
  supplier?: Supplier | null
  onSubmit: (data: SupplierFormData) => Promise<void>
  onCancel: () => void
  isLoading?: boolean
}

export interface SupplierFormData {
  name: string
  phone: string
  email: string
  address: string
}

const defaultFormData: SupplierFormData = {
  name: '',
  phone: '',
  email: '',
  address: ''
}

export function SupplierForm({ supplier, onSubmit, onCancel, isLoading }: SupplierFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset
  } = useForm<SupplierFormData>({
    defaultValues: defaultFormData
  })

  useEffect(() => {
    if (supplier) {
      reset({
        name: supplier.name,
        phone: supplier.phone || '',
        email: supplier.email || '',
        address: supplier.address || ''
      })
    } else {
      reset(defaultFormData)
    }
  }, [supplier, reset])

  const formatPhoneNumber = (value: string) => {
    const cleaned = value.replace(/\D/g, '')
    if (cleaned.startsWith('254')) return '+' + cleaned
    if (cleaned.startsWith('0')) return cleaned
    return cleaned
  }

  const onFormSubmit = async (data: SupplierFormData) => {
    await onSubmit({ ...data, phone: formatPhoneNumber(data.phone) })
  }

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Supplier Name *
        </label>
        <input
          type="text"
          {...register('name', { required: 'Supplier name is required', minLength: { value: 2, message: 'Name must be at least 2 characters' } })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.name ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="Enter supplier name"
        />
        {errors.name && <p className="mt-1 text-sm text-red-500">{errors.name.message}</p>}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Phone Number
        </label>
        <input
          type="tel"
          {...register('phone', {
            pattern: {
              value: /^(\+254|254|0)[1-9]\d{8}$/,
              message: 'Invalid Kenyan phone number'
            }
          })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.phone ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="0712345678"
        />
        {errors.phone && <p className="mt-1 text-sm text-red-500">{errors.phone.message}</p>}
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
          placeholder="supplier@email.com"
        />
        {errors.email && <p className="mt-1 text-sm text-red-500">{errors.email.message}</p>}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Address
        </label>
        <textarea
          {...register('address')}
          className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition resize-none"
          rows={3}
          placeholder="Enter supplier address"
        />
      </div>

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
          {isLoading ? 'Saving...' : supplier ? 'Update' : 'Add'} Supplier
        </button>
      </div>
    </form>
  )
}
