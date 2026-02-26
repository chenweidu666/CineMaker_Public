import request from '@/utils/request'
import type { TeamInfo } from './auth'

export function getTeam() {
  return request.get<TeamInfo>('/team')
}

export function updateTeam(data: { name: string }) {
  return request.put<TeamInfo>('/team', data)
}

export function inviteMember(data: { email: string }) {
  return request.post<any>('/team/invite', data)
}

export function acceptInvitation(data: { token: string }) {
  return request.post<any>('/team/invite/accept', data)
}

export function removeMember(memberId: number) {
  return request.delete<any>(`/team/members/${memberId}`)
}
