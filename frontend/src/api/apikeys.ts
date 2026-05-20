import apiClient from './client'

export interface ApiKeyItem {
  id: number
  user_id: number
  name: string
  key: string
  last_used: string | null
  created_at: string
  updated_at: string
}

export const apiKeysApi = {
  list() {
    return apiClient.get<ApiKeyItem[]>('/api-keys')
  },

  create(data: { name: string }) {
    return apiClient.post<ApiKeyItem>('/api-keys', data)
  },

  revoke(id: number) {
    return apiClient.delete(`/api-keys/${id}`)
  },
}
