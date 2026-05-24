import apiClient from './client'
import type { Tag } from '@/types'

export const tagsApi = {
  list() {
    return apiClient.get<Tag[]>('/tags')
  },
  delete(id: number) {
    return apiClient.delete(`/tags/${id}`)
  }
}
