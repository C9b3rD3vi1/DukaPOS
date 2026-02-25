import { useState, useCallback, useRef } from 'react'
import { Capacitor } from '@capacitor/core'

interface UseCameraOptions {
  quality?: number
  saveToGallery?: boolean
}

interface UseCameraReturn {
  isSupported: boolean
  hasPermission: boolean | null
  isTakingPhoto: boolean
  capturedPhoto: string | null
  error: string | null
  requestPermission: () => Promise<boolean>
  takePhoto: (options?: UseCameraOptions) => Promise<string | null>
  pickFromGallery: () => Promise<string | null>
  clearPhoto: () => void
}

export function useCamera(): UseCameraReturn {
  const [isSupported] = useState(true)
  const [hasPermission, setHasPermission] = useState<boolean | null>(true)
  const [isTakingPhoto, setIsTakingPhoto] = useState(false)
  const [capturedPhoto, setCapturedPhoto] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement | null>(null)

  const requestPermission = useCallback(async (): Promise<boolean> => {
    setHasPermission(true)
    return true
  }, [])

  const processFile = useCallback((file: File): Promise<string | null> => {
    return new Promise((resolve) => {
      const reader = new FileReader()
      reader.onload = (e) => {
        const result = e.target?.result as string
        setCapturedPhoto(result)
        resolve(result)
      }
      reader.onerror = () => {
        setError('Failed to read file')
        resolve(null)
      }
      reader.readAsDataURL(file)
    })
  }, [])

  const takePhoto = useCallback(async (options: UseCameraOptions = {}): Promise<string | null> => {
    setIsTakingPhoto(true)
    setError(null)

    try {
      if (typeof window !== 'undefined' && Capacitor.isNativePlatform()) {
        const { Camera } = await import('@capacitor/camera')
        const { CameraResultType, CameraSource } = await import('@capacitor/camera')
        
        const result = await Camera.getPhoto({
          quality: options.quality || 80,
          resultType: CameraResultType.Base64,
          source: CameraSource.Camera,
          saveToGallery: options.saveToGallery || false,
          allowEditing: false
        })

        if (result.base64String) {
          const photoData = `data:image/jpeg;base64,${result.base64String}`
          setCapturedPhoto(photoData)
          return photoData
        }
        return null
      }
      
      if (!inputRef.current) {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = 'image/*'
        input.capture = 'environment'
        inputRef.current = input
      }
      
      const file = await new Promise<File>((resolve) => {
        inputRef.current!.onchange = (e) => {
          const file = (e.target as HTMLInputElement).files?.[0]
          if (file) resolve(file)
        }
        inputRef.current!.click()
      })
      
      return processFile(file)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to take photo'
      setError(message)
      return null
    } finally {
      setIsTakingPhoto(false)
    }
  }, [processFile])

  const pickFromGallery = useCallback(async (): Promise<string | null> => {
    setIsTakingPhoto(true)
    setError(null)

    try {
      if (typeof window !== 'undefined' && Capacitor.isNativePlatform()) {
        const { Camera } = await import('@capacitor/camera')
        const { CameraResultType, CameraSource } = await import('@capacitor/camera')
        
        const result = await Camera.getPhoto({
          quality: 80,
          resultType: CameraResultType.Base64,
          source: CameraSource.Photos,
          allowEditing: false
        })

        if (result.base64String) {
          const photoData = `data:image/jpeg;base64,${result.base64String}`
          setCapturedPhoto(photoData)
          return photoData
        }
        return null
      }
      
      if (!inputRef.current) {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = 'image/*'
        inputRef.current = input
      }
      
      const file = await new Promise<File>((resolve) => {
        inputRef.current!.onchange = (e) => {
          const file = (e.target as HTMLInputElement).files?.[0]
          if (file) resolve(file)
        }
        inputRef.current!.click()
      })
      
      return processFile(file)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to pick photo'
      setError(message)
      return null
    } finally {
      setIsTakingPhoto(false)
    }
  }, [processFile])

  const clearPhoto = useCallback(() => {
    setCapturedPhoto(null)
    setError(null)
    if (inputRef.current) {
      inputRef.current.value = ''
    }
  }, [])

  return {
    isSupported,
    hasPermission,
    isTakingPhoto,
    capturedPhoto,
    error,
    requestPermission,
    takePhoto,
    pickFromGallery,
    clearPhoto
  }
}

export function useGalleryPicker() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement | null>(null)

  const pickImage = useCallback(async (maxSize?: number): Promise<string | null> => {
    setIsLoading(true)
    setError(null)

    try {
      if (typeof window !== 'undefined' && Capacitor.isNativePlatform()) {
        const { Camera } = await import('@capacitor/camera')
        const { CameraResultType, CameraSource } = await import('@capacitor/camera')
        
        const result = await Camera.getPhoto({
          quality: 80,
          resultType: CameraResultType.Base64,
          source: CameraSource.Photos,
          width: maxSize,
          height: maxSize,
          allowEditing: true
        })

        if (result.base64String) {
          return `data:image/jpeg;base64,${result.base64String}`
        }
        return null
      }
      
      if (!inputRef.current) {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = 'image/*'
        inputRef.current = input
      }
      
      const file = await new Promise<File>((resolve, reject) => {
        inputRef.current!.onchange = (e) => {
          const file = (e.target as HTMLInputElement).files?.[0]
          if (file) resolve(file)
          else reject(new Error('No file selected'))
        }
        inputRef.current!.onerror = () => reject(new Error('Failed to select file'))
        inputRef.current!.click()
      })
      
      return new Promise((resolve) => {
        const reader = new FileReader()
        reader.onload = (ev) => resolve(ev.target?.result as string)
        reader.onerror = () => {
          setError('Failed to read file')
          resolve(null)
        }
        reader.readAsDataURL(file)
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to pick image'
      setError(message)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { pickImage, isLoading, error }
}
