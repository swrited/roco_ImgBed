import apiClient from './client'
import type { LoginData, RegisterData, User } from '@/types'

export const authApi = {
  login(data: LoginData) {
    return apiClient.post<{ token: string; user: User }>('/tokens', data)
  },

  register(data: RegisterData) {
    return apiClient.post('/register', data)
  },

  logout() {
    return apiClient.delete('/tokens')
  },

  forgotPassword(email: string) {
    return apiClient.post('/forgot-password', { email })
  },

  resetPassword(data: { token: string; email: string; password: string; password_confirmation: string }) {
    return apiClient.post('/reset-password', data)
  },
}
