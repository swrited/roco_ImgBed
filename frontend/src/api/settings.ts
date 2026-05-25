import apiClient from './client'

interface PublicSettings {
  app_name?: string
  app_version?: string
  site_description?: string
  site_keywords?: string
  is_enable_registration?: string
  is_enable_gallery?: string
  is_enable_api?: string
  is_enable_ai_image?: string
  site_bg_image?: string
  site_bg_opacity?: string
  [key: string]: unknown
}

export const settingsApi = {
  /** 获取公开系统设置（无需登录） */
  getPublicSettings() {
    return apiClient.get<PublicSettings>('/settings/public')
  },

  /** 上传背景图片（管理员） */
  uploadBgImage(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post<{ url: string }>('/admin/settings/bg-upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}
