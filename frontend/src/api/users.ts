import apiClient from './client'
import type { User } from '@/types'

export interface DayStat {
  date: string
  count: number
}

export interface DashboardData {
  user: User
  image_count: number
  album_count: number
  used_capacity: number
  today_count: number
  month_count: number
  daily_stats: DayStat[]
}

export const usersApi = {
  profile() {
    return apiClient.get<User>('/profile')
  },

  dashboard() {
    return apiClient.get<DashboardData>('/dashboard')
  },

  updateProfile(data: { name?: string; url?: string; old_password?: string; password?: string }) {
    return apiClient.put<User>('/profile', data)
  },

  setStrategy(strategyId: number) {
    return apiClient.put('/user/settings/strategy', { strategy_id: strategyId })
  },
}
