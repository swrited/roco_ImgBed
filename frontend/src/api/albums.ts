import apiClient from './client'
import type { Album } from '@/types'

export const albumsApi = {
  list() {
    return apiClient.get<Album[]>('/albums')
  },

  create(data: { name: string; intro?: string; permission: number; cover_image_id?: number | null }) {
    return apiClient.post<Album>('/albums', data)
  },

  update(id: number, data: { name?: string; intro?: string; permission: number; cover_image_id?: number | null }) {
    return apiClient.put<Album>(`/albums/${id}`, data)
  },

  delete(id: number) {
    return apiClient.delete(`/albums/${id}`)
  },
}
