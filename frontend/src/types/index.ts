export interface User {
  id: number
  name: string
  email: string
  avatar?: string
  capacity: number
  url?: string
  image_num: number
  album_num: number
  used_capacity?: number
  is_adminer: boolean
  group_id: number
  created_at: string
}

export interface Image {
  key: string
  name: string
  origin_name: string
  path: string
  size: number
  mimetype: string
  extension: string
  md5: string
  sha1: string
  width: number
  height: number
  url: string
  permission: number
  created_at: string
  album_id?: number
  strategy_id?: number
}

export interface Album {
  id: number
  name: string
  intro?: string
  image_num: number
  created_at: string
}

export interface Strategy {
  id: number
  name: string
  intro?: string
  key: number
  configs: Record<string, any>
}

export interface Group {
  id: number
  name: string
  is_default: boolean
  configs: Record<string, any>
  strategies?: Strategy[]
}

export interface PaginatedResponse<T> {
  data: T[]
  current_page: number
  last_page: number
  per_page: number
  total: number
  from: number
  to: number
}

export interface ApiResponse<T = any> {
  status: boolean
  message: string
  data: T
}

export interface LoginData {
  email: string
  password: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
  password_confirmation: string
}
