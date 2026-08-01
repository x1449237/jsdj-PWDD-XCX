import request from '@/utils/request'

// 用户列表
export function getUserList(params) {
  return request({ url: '/users', method: 'get', params })
}

// 用户详情
export function getUserDetail(id) {
  return request({ url: `/users/${id}`, method: 'get' })
}

// 用户操作（冻结/解冻等）
export function handleUserAction(id, action) {
  return request({ url: `/users/${id}/${action}`, method: 'post' })
}

// 解绑邀请关系
export function unbindInvite(id) {
  return request({ url: `/users/${id}/unbind-invite`, method: 'post' })
}

// 导出用户列表
export function exportUsers(params) {
  return request({ url: '/users/export', method: 'get', params, responseType: 'blob' })
}
