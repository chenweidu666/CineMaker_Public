import request from '../utils/request'

export interface AIMessageLog {
  id: number
  request_id: string
  drama_id?: number
  service_type: string
  purpose: string
  provider: string
  model: string
  endpoint: string
  system_prompt: string
  user_prompt: string
  full_request?: Record<string, unknown>
  response: string
  status: string
  error_message: string
  duration_ms: number
  prompt_tokens: number
  output_tokens: number
  total_tokens: number
  created_at: string
}

export interface AIMessageLogQuery {
  page?: number
  page_size?: number
  service_type?: string
  purpose?: string
  status?: string
  drama_id?: number
  start_date?: string
  end_date?: string
  keyword?: string
}

export interface AIMessageLogListResult {
  items: AIMessageLog[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface AILogStats {
  total: number
  success: number
  failed: number
  today: number
  by_type: { service_type: string; count: number }[]
  by_purpose_top: { purpose: string; count: number }[]
}

export const aiLogAPI = {
  list(params?: AIMessageLogQuery) {
    return request.get<AIMessageLogListResult>('/ai-logs', { params })
  },

  getById(id: number) {
    return request.get<AIMessageLog>(`/ai-logs/${id}`)
  },

  getStats() {
    return request.get<AILogStats>('/ai-logs/stats')
  }
}
