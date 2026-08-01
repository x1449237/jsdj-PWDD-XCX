import request from '@/utils/request'

// 生成邀请码
export function generateInviteCodes(data) {
  return request({ url: '/invite-codes/generate', method: 'post', data })
}

// 邀请码列表
export function getInviteCodes(params) {
  return request({ url: '/invite-codes', method: 'get', params })
}

// 作废邀请码
export function invalidateInviteCode(id) {
  return request({ url: `/invite-codes/${id}/invalidate`, method: 'put' })
}

// 导出邀请码
export function exportInviteCode(id) {
  return request({ url: `/invite-codes/${id}/export`, method: 'get', responseType: 'blob' })
}
