import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useCamera } from '@/hooks/useCamera'
import type { Product } from '@/api/types'

interface ProductFormProps {
  product?: Product | null
  categories: string[]
  onSubmit: (data: ProductFormData) => Promise<void>
  onCancel: () => void
  isLoading?: boolean
}

export interface ProductFormData {
  name: string
  category: string
  unit: string
  cost_price: number
  selling_price: number
  current_stock: number
  low_stock_threshold: number
  barcode: string
  image_url: string
}

const defaultFormData: ProductFormData = {
  name: '',
  category: '',
  unit: 'pcs',
  cost_price: 0,
  selling_price: 0,
  current_stock: 0,
  low_stock_threshold: 10,
  barcode: '',
  image_url: ''
}

export function ProductForm({ product, categories, onSubmit, onCancel, isLoading }: ProductFormProps) {
  const [imagePreview, setImagePreview] = useState('')
  const [uploadingImage, setUploadingImage] = useState(false)
  const { isSupported: cameraSupported, isTakingPhoto, takePhoto, pickFromGallery } = useCamera()
  
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue
  } = useForm<ProductFormData>({
    defaultValues: defaultFormData
  })

  useEffect(() => {
    if (product) {
      reset({
        name: product.name,
        category: product.category || '',
        unit: product.unit,
        cost_price: product.cost_price,
        selling_price: product.selling_price,
        current_stock: product.current_stock,
        low_stock_threshold: product.low_stock_threshold,
        barcode: product.barcode || '',
        image_url: product.image_url || ''
      })
      setImagePreview(product.image_url || '')
    } else {
      reset(defaultFormData)
      setImagePreview('')
    }
  }, [product, reset])

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploadingImage(true)
    try {
      const reader = new FileReader()
      reader.onloadend = () => {
        const base64 = reader.result as string
        setValue('image_url', base64)
        setImagePreview(base64)
        setUploadingImage(false)
      }
      reader.readAsDataURL(file)
    } catch {
      setUploadingImage(false)
    }
  }

  const onFormSubmit = async (data: ProductFormData) => {
    await onSubmit(data)
  }

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Product Name *</label>
        <input
          type="text"
          {...register('name', { required: 'Product name is required', minLength: { value: 2, message: 'Name must be at least 2 characters' } })}
          className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
        />
        {errors.name && <p className="text-red-500 text-sm mt-1">{errors.name.message}</p>}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
          <input
            type="text"
            {...register('category')}
            list="categories"
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
          <datalist id="categories">
            {categories.map((cat) => (
              <option key={cat} value={cat} />
            ))}
          </datalist>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Unit</label>
          <select
            {...register('unit')}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          >
            <option value="pcs">Pieces</option>
            <option value="kg">Kilograms</option>
            <option value="g">Grams</option>
            <option value="L">Liters</option>
            <option value="ml">Milliliters</option>
            <option value="pack">Pack</option>
            <option value="box">Box</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Cost Price *</label>
          <input
            type="number"
            {...register('cost_price', { required: 'Cost price is required', min: { value: 0, message: 'Must be positive' } })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
          {errors.cost_price && <p className="text-red-500 text-sm mt-1">{errors.cost_price.message}</p>}
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Selling Price *</label>
          <input
            type="number"
            {...register('selling_price', { required: 'Selling price is required', min: { value: 0, message: 'Must be positive' } })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
          {errors.selling_price && <p className="text-red-500 text-sm mt-1">{errors.selling_price.message}</p>}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Current Stock *</label>
          <input
            type="number"
            {...register('current_stock', { required: 'Stock is required', min: { value: 0, message: 'Must be positive' } })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
          {errors.current_stock && <p className="text-red-500 text-sm mt-1">{errors.current_stock.message}</p>}
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Low Stock Alert</label>
          <input
            type="number"
            {...register('low_stock_threshold', { min: { value: 0, message: 'Must be positive' } })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Barcode</label>
        <input
          type="text"
          {...register('barcode')}
          className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary focus:border-transparent outline-none"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Product Image</label>
        <div className="flex items-center gap-4">
          <div className="w-20 h-20 bg-gray-100 rounded-xl flex items-center justify-center overflow-hidden">
            {imagePreview ? (
              <img src={imagePreview} alt="Preview" className="w-full h-full object-cover" />
            ) : (
              <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            )}
          </div>
          <div className="flex-1 flex gap-2">
            {cameraSupported && (
              <button
                type="button"
                onClick={async () => {
                  const photo = await takePhoto()
                  if (photo) {
                    setValue('image_url', photo)
                    setImagePreview(photo)
                  }
                }}
                disabled={isTakingPhoto}
                className="flex-1 inline-flex items-center justify-center px-4 py-2 bg-primary text-white rounded-xl text-sm font-medium transition disabled:opacity-50"
              >
                {isTakingPhoto ? 'Taking...' : 'Camera'}
              </button>
            )}
            <button
              type="button"
              onClick={async () => {
                const photo = await pickFromGallery()
                if (photo) {
                  setValue('image_url', photo)
                  setImagePreview(photo)
                }
              }}
              disabled={isTakingPhoto}
              className="flex-1 inline-flex items-center justify-center px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-xl text-sm font-medium transition disabled:opacity-50"
            >
              Gallery
            </button>
            <label className="flex-1 cursor-pointer">
              <input
                type="file"
                accept="image/*"
                onChange={handleImageUpload}
                className="hidden"
              />
              <span className="inline-flex items-center justify-center w-full px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-xl text-sm font-medium transition">
                {uploadingImage ? 'Uploading...' : 'Choose File'}
              </span>
            </label>
          </div>
        </div>
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
          {isLoading ? 'Saving...' : product ? 'Update' : 'Add'} Product
        </button>
      </div>
    </form>
  )
}
