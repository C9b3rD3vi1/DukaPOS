import { useForm, useFieldArray } from 'react-hook-form'
import type { Product, Supplier } from '@/api/types'

interface OrderItem {
  productId: number
  productName: string
  quantity: number
  unitPrice: number
}

interface OrderFormProps {
  products: Product[]
  suppliers: Supplier[]
  onSubmit: (data: OrderFormData) => Promise<void>
  onCancel: () => void
  isLoading?: boolean
}

export interface OrderFormData {
  supplier_id: number
  items: OrderItem[]
  notes: string
  expected_delivery?: string
}

export function OrderForm({ products, suppliers, onSubmit, onCancel, isLoading }: OrderFormProps) {
  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
    watch
  } = useForm<OrderFormData>({
    defaultValues: {
      supplier_id: 0,
      items: [],
      notes: '',
      expected_delivery: ''
    }
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items'
  })

  const watchedItems = watch('items') || []

  const addItem = () => {
    if (products.length === 0) return
    const firstProduct = products[0]
    append({
      productId: firstProduct.id,
      productName: firstProduct.name,
      quantity: 1,
      unitPrice: firstProduct.cost_price
    })
  }

  const totalAmount = watchedItems.reduce(
    (sum: number, item) => sum + ((item?.quantity || 0) * (item?.unitPrice || 0)), 
    0
  )

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Supplier *
        </label>
        <select
          {...register('supplier_id', { required: 'Please select a supplier', valueAsNumber: true })}
          className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition bg-white ${
            errors.supplier_id ? 'border-red-500' : 'border-gray-200'
          }`}
        >
          <option value={0}>Select supplier</option>
          {suppliers.map((supplier) => (
            <option key={supplier.id} value={supplier.id}>
              {supplier.name}
            </option>
          ))}
        </select>
        {errors.supplier_id && <p className="mt-1 text-sm text-red-500">{errors.supplier_id.message}</p>}
      </div>

      <div>
        <div className="flex items-center justify-between mb-2">
          <label className="block text-sm font-medium text-gray-700">
            Order Items *
          </label>
          <button
            type="button"
            onClick={addItem}
            className="text-sm text-primary hover:underline"
          >
            + Add Item
          </button>
        </div>

        {errors.items && <p className="mt-1 text-sm text-red-500 mb-2">{errors.items?.message || 'Please add at least one item'}</p>}

        <div className="space-y-2">
          {fields.length === 0 ? (
            <div className="text-center py-8 text-gray-500 border-2 border-dashed border-gray-200 rounded-xl">
              No items added. Click "Add Item" to add products.
            </div>
          ) : (
            fields.map((field, index) => (
              <div key={field.id} className="flex items-center gap-2 p-3 bg-gray-50 rounded-xl">
                <select
                  {...register(`items.${index}.productId` as const, { valueAsNumber: true })}
                  className="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm"
                >
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.name}
                    </option>
                  ))}
                </select>
                
                <input
                  type="number"
                  {...register(`items.${index}.quantity` as const, { valueAsNumber: true, min: { value: 1, message: 'Min 1' } })}
                  min="1"
                  className="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm"
                  placeholder="Qty"
                />
                
                <input
                  type="number"
                  {...register(`items.${index}.unitPrice` as const, { valueAsNumber: true, min: { value: 0, message: 'Min 0' } })}
                  min="0"
                  className="w-24 px-3 py-2 border border-gray-200 rounded-lg text-sm"
                  placeholder="Price"
                />
                
                <button
                  type="button"
                  onClick={() => remove(index)}
                  className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Expected Delivery
          </label>
          <input
            type="date"
            {...register('expected_delivery')}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Total Amount
          </label>
          <div className="w-full px-4 py-3 bg-gray-100 rounded-xl font-bold text-lg text-gray-900">
            {formatCurrency(totalAmount)}
          </div>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Notes
        </label>
        <textarea
          {...register('notes')}
          className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition resize-none"
          rows={3}
          placeholder="Add any notes for this order..."
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
          {isLoading ? 'Creating...' : 'Create Order'}
        </button>
      </div>
    </form>
  )
}
