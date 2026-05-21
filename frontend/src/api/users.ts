import apiClient from './client'
import type { User } from '@/types'

export const usersApi = {
  profile() {
    return apiClient.get<User>('/profile')
  },

  updateProfile(data: { name?: string; url?: string; old_password?: string; password?: string }) {
    return apiClient.put<User>('/profile', data)
  },

  setStrategy(strategyId: number) {
    return apiClient.put('/user/settings/strategy', { strategy_id: strategyId })
  },
}
