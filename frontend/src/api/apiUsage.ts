import apiClient from './client'

export interface ApiUsageDay {
  date: string
  count: number
}

export interface ApiUsageEndpoint {
  method: string
  path: string
  count: number
}

export interface ApiUsageUser {
  user_id: number
  name: string
  email: string
  count: number
}

export interface ApiUsageLog {
  id: number
  user_id?: number
  name?: string
  email?: string
  api_key_id?: number
  api_key_name?: string
  method: string
  path: string
  status: number
  latency_ms: number
  ip: string
  auth_type: string
  created_at: string
  user?: { id: number; name: string; email: string }
  api_key?: { id: number; name: string }
}

export interface ApiUsageStats {
  total: number
  today: number
  last_7_days: number
  daily: ApiUsageDay[]
  endpoints: ApiUsageEndpoint[]
  by_user?: ApiUsageUser[]
  recent?: ApiUsageLog[]
  range?: { start: string; end: string }
}

export const apiUsageApi = {
  mine(params?: Record<string, any>) {
    return apiClient.get<ApiUsageStats>('/api-usage', { params })
  },
  admin(params?: Record<string, any>) {
    return apiClient.get<ApiUsageStats>('/admin/api-usage', { params })
  },
}
