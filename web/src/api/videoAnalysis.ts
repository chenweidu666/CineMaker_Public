import request from '../utils/request'

export interface StageData {
  download?: { status: string; title?: string; file_path?: string; duration?: number }
  detect?: { shot_count: number; duration: number; has_audio: boolean; shots?: { index: number; start_time: number; end_time: number; frame_url: string }[] }
  transcribe?: {
    status: string; error?: string; reason?: string; message?: string;
    done?: number; total?: number;
    shots?: { shot_index: number; start_time: number; end_time: number; status: string; text?: string; error?: string; segments?: { start: number; end: number; text: string }[] }[];
    segments?: { start: number; end: number; text: string }[];
  }
  analyze?: { status: string; done?: number; total?: number; frames?: { index: number; description: string; frame_url?: string }[] }
  synthesize?: { status: string }
}

export interface VideoAnalysisTask {
  id: number
  task_id: string
  video_url: string
  video_path: string
  title: string
  duration: number
  status: string
  progress: number
  stage: string
  shot_count: number
  result?: AnalysisResult
  stage_data?: StageData
  error_msg: string
  imported_drama_id?: number
  created_at: string
  updated_at: string
}

export interface AnalysisResult {
  title: string
  summary: string
  tags?: string[]
  characters: AnalysisChar[]
  shots: AnalysisShot[]
  dialogues: AnalysisLine[]
}

export interface AnalysisChar {
  name: string
  description: string
  role: string
}

export interface AnalysisShot {
  index: number
  start_time: number
  end_time: number
  title?: string
  description: string
  location: string
  time?: string
  characters: string[]
  dialogue: string
  mood: string
  shot_type?: string
  angle?: string
  movement?: string
  first_frame_desc?: string
  middle_action_desc?: string
  last_frame_desc?: string
  video_prompt?: string
  bgm_prompt?: string
  sound_effect?: string
  frame_path?: string
}

export interface AnalysisLine {
  start_time: number
  end_time: number
  speaker: string
  text: string
}

export const videoAnalysisAPI = {
  list() {
    return request.get<{ items: VideoAnalysisTask[] }>('/video-analysis')
  },

  upload(file: File) {
    const formData = new FormData()
    formData.append('video', file)
    return request.post<{ task_id: string; status: string; message: string }>(
      '/video-analysis/upload',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 600000 }
    )
  },

  analyzeFromURL(url: string) {
    return request.post<{ task_id: string; status: string; message: string }>(
      '/video-analysis/from-url',
      { url }
    )
  },

  getStatus(taskId: string) {
    return request.get<VideoAnalysisTask>(`/video-analysis/${taskId}/status`)
  },

  retry(taskId: string) {
    return request.post<{ task_id: string; status: string; message: string }>(
      `/video-analysis/${taskId}/retry`
    )
  },

  resynthesize(taskId: string, includeAudio: boolean = true) {
    return request.post<{ task_id: string; status: string; message: string }>(
      `/video-analysis/${taskId}/resynthesize`,
      { include_audio: includeAudio }
    )
  },

  deleteTask(taskId: string) {
    return request.delete<{ message: string }>(`/video-analysis/${taskId}`)
  },

  importToDrama(taskId: string, title?: string) {
    return request.post<{ message: string; drama_id: number; title: string }>(
      `/video-analysis/${taskId}/import`,
      { title }
    )
  },

  exportAnalysis(taskId: string) {
    const baseURL = request.defaults.baseURL || '/api/v1'
    window.open(`${baseURL}/video-analysis/${taskId}/export`, '_blank')
  }
}

export const dramaExportImportAPI = {
  exportDrama(dramaId: number) {
    const baseURL = request.defaults.baseURL || '/api/v1'
    window.open(`${baseURL}/dramas/${dramaId}/export`, '_blank')
  },

  importDrama(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<{ message: string; drama_id: number }>(
      '/dramas/import',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 600000 }
    )
  }
}
