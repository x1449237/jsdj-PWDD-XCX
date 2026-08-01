import request from '@/utils/request'

// ============ 打手审核 ============
export function getPlayerAuditList(params) {
  return request({ url: '/audit/players', method: 'get', params })
}
export function getPlayerAuditDetail(id) {
  return request({ url: `/audit/players/${id}`, method: 'get' })
}
export function approvePlayer(id) {
  return request({ url: `/audit/players/${id}/approve`, method: 'post' })
}
export function rejectPlayer(id, data) {
  return request({ url: `/audit/players/${id}/reject`, method: 'post', data })
}
export function delistPlayer(id) {
  return request({ url: `/audit/players/${id}/delist`, method: 'post' })
}

// ============ 分销商审核 ============
export function getDistributorAuditList(params) {
  return request({ url: '/audit/distributors', method: 'get', params })
}
export function getDistributorAuditDetail(id) {
  return request({ url: `/audit/distributors/${id}`, method: 'get' })
}
export function approveDistributor(id) {
  return request({ url: `/audit/distributors/${id}/approve`, method: 'post' })
}
export function rejectDistributor(id, data) {
  return request({ url: `/audit/distributors/${id}/reject`, method: 'post', data })
}
export function delistDistributor(id) {
  return request({ url: `/audit/distributors/${id}/delist`, method: 'post' })
}

// ============ 派单员审核 ============
export function getDispatcherAuditList(params) {
  return request({ url: '/audit/dispatchers', method: 'get', params })
}
export function getDispatcherAuditDetail(id) {
  return request({ url: `/audit/dispatchers/${id}`, method: 'get' })
}
export function approveDispatcher(id) {
  return request({ url: `/audit/dispatchers/${id}/approve`, method: 'post' })
}
export function rejectDispatcher(id, data) {
  return request({ url: `/audit/dispatchers/${id}/reject`, method: 'post', data })
}
export function delistDispatcher(id) {
  return request({ url: `/audit/dispatchers/${id}/delist`, method: 'post' })
}

// ============ 内置管理员审核 ============
export function getAdminAuditList(params) {
  return request({ url: '/audit/admins', method: 'get', params })
}
export function getAdminAuditDetail(id) {
  return request({ url: `/audit/admins/${id}`, method: 'get' })
}
export function approveAdmin(id) {
  return request({ url: `/audit/admins/${id}/approve`, method: 'post' })
}
export function rejectAdmin(id, data) {
  return request({ url: `/audit/admins/${id}/reject`, method: 'post', data })
}
export function delistAdmin(id) {
  return request({ url: `/audit/admins/${id}/delist`, method: 'post' })
}

// ============ 俱乐部入驻审核 ============
export function getClubAuditList(params) {
  return request({ url: '/audit/club_list', method: 'get', params })
}
export function approveAudit(data) {
  return request({ url: '/audit/approve', method: 'put', data })
}
export function rejectAudit(data) {
  return request({ url: '/audit/reject', method: 'put', data })
}
export function forceOfflineAudit(data) {
  return request({ url: '/audit/force_offline', method: 'put', data })
}
