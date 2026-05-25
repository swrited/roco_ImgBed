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

export interface ApiKey {
  id: number
  name: string
  key: string
  last_used?: string
  revoked_at?: string
  created_at: string
}

export interface ApiUsageStats {
  total_requests: number
  total_success: number
  total_errors: number
  avg_latency_ms: number
  requests_by_day: { date: string; count: number }[]
  top_endpoints: { path: string; count: number }[]
}

export interface SystemSettings {
  app_name?: string
  app_version?: string
  site_description?: string
  site_keywords?: string
  is_enable_registration?: boolean | string
  is_enable_guest_upload?: boolean | string
  is_enable_gallery?: boolean | string
  is_enable_api?: boolean | string
  is_enable_ai_image?: boolean | string
  upload_max_size?: number | string
  user_initial_capacity?: number | string
  default_strategy_id?: string
  site_bg_image?: string
  site_bg_opacity?: number | string
  ai_image_provider?: string
  minimax_api_key?: string
  minimax_api_endpoint?: string
  minimax_model?: string
  openai_image_api_key?: string
  openai_image_api_endpoint?: string
  openai_image_model?: string
  siliconflow_image_api_key?: string
  siliconflow_image_api_endpoint?: string
  siliconflow_image_model?: string
  compatible_image_api_key?: string
  compatible_image_api_endpoint?: string
  compatible_image_model?: string
  ai_image_max_count?: number | string
  ai_image_rate_limit_seconds?: number | string
  ai_image_daily_limit?: number | string
  api_key_minute_limit?: number | string
  api_key_daily_limit?: number | string
  mail_host?: string
  mail_port?: string
  mail_username?: string
  mail_password?: string
}
