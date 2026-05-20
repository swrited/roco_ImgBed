import apiClient from './client'
import type { Strategy } from '@/types'

export const strategiesApi = {
  list() {
    return apiClient.get<Strategy[]>('/strategies')
  },
}
