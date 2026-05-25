import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api/auth'
import { usersApi } from '@/api/users'
import type { User, LoginData, RegisterData } from '@/types'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))
  const isAuthenticated = ref(!!token.value)
  const isAdmin = ref(user.value?.is_adminer || false)

  async function login(data: LoginData) {
    const res = await authApi.login(data) as any
    token.value = res.token
    user.value = res.user
    isAuthenticated.value = true
    isAdmin.value = res.user?.is_adminer || false
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  async function register(data: RegisterData) {
    await authApi.register(data)
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // ignore
    }
    token.value = ''
    user.value = null
    isAuthenticated.value = false
    isAdmin.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    router.push('/login')
  }

  async function fetchProfile() {
    const userData = await usersApi.profile() as any
    user.value = userData
    isAdmin.value = userData.is_adminer
    localStorage.setItem('user', JSON.stringify(userData))
    return userData
  }

  return { token, user, isAuthenticated, isAdmin, login, register, logout, fetchProfile, fetchUser: fetchProfile }
})
