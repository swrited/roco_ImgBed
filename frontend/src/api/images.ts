import apiClient from './client'
import type { Image, PaginatedResponse } from '@/types'

export const imagesApi = {
  list(params?: Record<string, any>) {
    return apiClient.get<PaginatedResponse<Image>>('/images', { params })
  },

  upload(formData: FormData, onProgress?: (progress: number) => void) {
    return apiClient.post('/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },

  delete(keys: string[]) {
    return apiClient.delete('/images', { data: { keys } })
  },

  rename(key: string, alias_name: string) {
    return apiClient.put('/images/rename', { key, alias_name })
  },

  move(keys: string[], album_id: number | null) {
    return apiClient.put('/images/movement', { keys, album_id })
  },

  updateTags(key: string, tags: string[]) {
    return apiClient.put(`/images/${key}/tags`, { tags })
  },

  getTrash(params?: Record<string, any>) {
    return apiClient.get<PaginatedResponse<Image>>('/images/trash', { params })
  },

  restore(keys: string[]) {
    return apiClient.put('/images/trash/restore', { keys })
  },

  forceDelete(keys: string[]) {
    return apiClient.delete('/images/trash/force', { data: { keys } })
  },
}
