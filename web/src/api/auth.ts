import request from '@/utils/request'

export interface LoginParams {
  email: string
  password: string
}

export interface RegisterParams {
  username: string
  email: string
  password: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface UserInfo {
  id: number
  username: string
  email: string
  role: string
  team_id: number | null
  avatar: string
  team?: TeamInfo
  created_at: string
}

export interface TeamInfo {
  id: number
  name: string
  owner_id: number
  owner?: UserInfo
  members?: UserInfo[]
  created_at: string
}

export function login(data: LoginParams) {
  return request.post<TokenResponse>('/auth/login', data)
}

export function register(data: RegisterParams) {
  return request.post<TokenResponse>('/auth/register', data)
}

export function refreshToken(refresh_token: string) {
  return request.post<TokenResponse>('/auth/refresh', { refresh_token })
}

export function getMe() {
  return request.get<UserInfo>('/auth/me')
}

export function updateMe(data: { username?: string; avatar?: string }) {
  return request.put<UserInfo>('/auth/me', data)
}
