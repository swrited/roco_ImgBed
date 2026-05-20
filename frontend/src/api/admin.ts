import apiClient from './client'
import type { Group, PaginatedResponse, Strategy, User } from '@/types'

interface ConsoleData {
  stats: {
    users: number
    images: number
    albums: number
    recent_uploads: number
  }
  daily: Array<{ date: string; count: number }>
}

interface SettingsData {
  version?: string
  [key: string]: unknown
}

export const adminApi = {
  // Console
  getConsole() {
    return apiClient.get<ConsoleData>('/admin/console')
  },

  // Users
  listUsers(params?: Record<string, any>) {
    return apiClient.get<PaginatedResponse<User>>('/admin/users', { params })
  },
  getUser(id: number) {
    return apiClient.get<User>(`/admin/users/${id}`)
  },
  updateUser(id: number, data: any) {
    return apiClient.put(`/admin/users/${id}`, data)
  },
  deleteUser(id: number) {
    return apiClient.delete(`/admin/users/${id}`)
  },

  // Images
  listImages(params?: Record<string, any>) {
    return apiClient.get<PaginatedResponse<any>>('/admin/images', { params })
  },
  deleteImages(keys: string[]) {
    return apiClient.delete('/admin/images', { data: { keys } })
  },

  // Groups
  listGroups() {
    return apiClient.get<Group[]>('/admin/groups')
  },
  createGroup(data: any) {
    return apiClient.post('/admin/groups', data)
  },
  updateGroup(id: number, data: any) {
    return apiClient.put(`/admin/groups/${id}`, data)
  },
  deleteGroup(id: number) {
    return apiClient.delete(`/admin/groups/${id}`)
  },
  clearGroupCache() {
    return apiClient.delete('/admin/groups/clear-cache')
  },

  // Strategies
  listStrategies() {
    return apiClient.get<Strategy[]>('/admin/strategies')
  },
  createStrategy(data: any) {
    return apiClient.post('/admin/strategies', data)
  },
  updateStrategy(id: number, data: any) {
    return apiClient.put(`/admin/strategies/${id}`, data)
  },
  deleteStrategy(id: number) {
    return apiClient.delete(`/admin/strategies/${id}`)
  },

  // Settings
  getSettings() {
    return apiClient.get<SettingsData>('/admin/settings')
  },
  updateSettings(data: any) {
    return apiClient.put('/admin/settings', data)
  },
  testMail(data: any) {
    return apiClient.post('/admin/settings/mail-test', data)
  },
  upgrade() {
    return apiClient.post('/admin/upgrade')
  },
  checkUpgrade() {
    return apiClient.get<SettingsData>('/admin/upgrade')
  },
}
