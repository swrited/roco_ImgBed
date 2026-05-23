import apiClient from './client'

export interface AIImageGeneratePayload {
  prompt: string
  aspect_ratio: string
  count: number
  prompt_optimizer: boolean
}

export interface AIImageGenerateResponse {
  album: {
    id: number
    name: string
  }
  quota?: {
    limit: number
    used: number
    remaining: number
  }
  images: Array<{
    key: string
    name: string
    origin_name: string
    pathname: string
    size: number
    width: number
    height: number
    album_id: number
    links: {
      url: string
      html: string
      bbcode: string
      markdown: string
      markdown_with_link: string
      thumbnail_url: string
    }
  }>
}

export const aiImagesApi = {
  generate(data: AIImageGeneratePayload) {
    return apiClient.post<AIImageGenerateResponse>('/ai/images', data)
  },
}
