import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import type { Staff } from '@/api/types'

interface StaffFormProps {
  staff?: Staff | null
  onSubmit: (data: StaffFormData) => Promise<void>
  onCancel: () => void
  isLoading?: boolean
}

export interface StaffFormData {
  name: string
  phone: string
  role: string
  is_active: boolean
}

const defaultFormData: StaffFormData = {
  name: '',
  phone: '',
  role: 'staff',
  is_active: true
}

const roleOptions = [
  { value: 'admin', label: 'Admin' },
  { value: 'manager', label: 'Manager' },
  { value: 'staff', label: 'Staff' },
  { value: 'cashier', label: 'Cashier' },
  { value: 'viewer', label: 'Viewer' }
]

export function StaffForm({ staff, onSubmit, onCancel, isLoading }: StaffFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset
  } = useForm<StaffFormData>({
    defaultValues: defaultFormData
  })

  useEffect(() => {
    if (staff) {
      reset({
        name: staff.name,
        phone: staff.phone,
        role: staff.role,
        is_active: staff.is_active
      })
    } else {
      reset(defaultFormData)
    }
  }, [staff, reset])

  const formatPhoneNumber = (value: string) => {
    const cleaned = value.replace(/\D/g, '')
    if (cleaned.startsWith('254')) return '+' + cleaned
    if (cleaned.startsWith('0')) return cleaned
    return cleaned
  }

  const getRoleDescription = (role: string) => {
    const descriptions: Record<string, string> = {
      admin: 'Full access to all features and settings',
      manager: 'Can manage products, sales, and reports',
      cashier: 'Can process sales and view products',
      staff: 'Basic access to products and sales',
      viewer: 'View-only access to dashboard and reports'
    }
    return descriptions[role] || ''
  }

  const onFormSubmit = async (data: StaffFormData) => {
    await onSubmit({ ...data, phone: formatPhoneNumber(data.phone) })
  }

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Staff Name *
        </label>
        <input
          type="text"
          {...register('name', { required: 'Name is required', minLength: { value: 2, message: 'Name must be at least 2 characters' } })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition ${
            errors.name ? 'border-red-500' : 'border-gray-200'
          }`}
          placeholder="Enter staff name"
        />
        {errors.name && <p className="mt-1 text-sm text-red-500">{errors.name.message}</p>}
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
          placeholder="0712345678"
        />
        {errors.phone && <p className="mt-1 text-sm text-red-500">{errors.phone.message}</p>}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Role *
        </label>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {roleOptions.map((option) => (
            <label
              key={option.value}
              className={`flex items-start p-3 border-2 rounded-xl cursor-pointer transition ${
                option.value === 'admin' || option.value === 'manager' || option.value === 'cashier' || option.value === 'staff' || option.value === 'viewer'
                  ? 'border-gray-200 hover:border-gray-300'
                  : ''
              }`}
            >
              <input
                type="radio"
                {...register('role', { required: 'Role is required' })}
                value={option.value}
                className="mt-1 mr-3"
              />
              <div>
                <span className="font-medium text-gray-900">{option.label}</span>
                <p className="text-xs text-gray-500 mt-0.5">
                  {getRoleDescription(option.value)}
                </p>
              </div>
            </label>
          ))}
        </div>
        {errors.role && <p className="mt-1 text-sm text-red-500">{errors.role.message}</p>}
      </div>

      {staff && (
        <div className="flex items-center gap-3 p-4 bg-gray-50 rounded-xl">
          <input
            type="checkbox"
            id="is_active"
            {...register('is_active')}
            className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
          />
          <label htmlFor="is_active" className="text-sm text-gray-700">
            Staff account is active
          </label>
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
          {isLoading ? 'Saving...' : staff ? 'Update' : 'Add'} Staff
        </button>
      </div>
    </form>
  )
}
