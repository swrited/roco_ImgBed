export interface User {
  id: number
  name: string
  email: string
  avatar?: string
  token?: string
  capacity: number
  url?: string
  image_num: number
  album_num: number
  used_capacity?: number
  remaining_capacity?: number
  capacity_percent?: number
  is_adminer: boolean
  group_id: number
  configs?: Record<string, any>
  created_at: string
}

export interface Image {
  id: number
  key: string
  name: string
  alias_name: string
  origin_name: string
  url: string
  path: string
  size: number
  mimetype: string
  extension: string
  md5: string
  sha1: string
  width: number
  height: number
  is_unhealthy: boolean
  uploaded_ip: string
  deleted_at?: string
  created_at: string
  updated_at: string
  links: ImageLinks
  album_id?: number
  tags?: Tag[]
}

export interface ImageLinks {
  url: string
  html: string
  bbcode: string
  markdown: string
  markdown_with_link: string
  thumbnail_url: string
}

export interface Tag {
  id: number
  name: string
  image_num: number
  created_at: string
}

export interface Album {
  id: number
  name: string
  intro: string
  image_num: number
  permission: number
  cover_image_id?: number | null
  user_name?: string
  cover_url?: string
  created_at: string
  updated_at: string
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
  is_guest: boolean
  configs: Record<string, any>
  strategies?: Strategy[]
  users_count?: number
  strategies_count?: number
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
