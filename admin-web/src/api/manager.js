import request from '@/utils/request'

// 管理员列表
export function getManagerList(params) {
  return request({ url: '/admins', method: 'get', params })
}

// 角色列表
export function getRoleList() {
  return request({ url: '/roles', method: 'get' })
}

// 新增管理员
export function createManager(data) {
  return request({ url: '/admins', method: 'post', data })
}

// 更新管理员
export function updateManager(id, data) {
  return request({ url: `/admins/${id}`, method: 'put', data })
}

// 删除管理员
export function deleteManager(id) {
  return request({ url: `/admins/${id}`, method: 'delete' })
}

// 修改管理员角色
export function updateManagerRole(id, data) {
  return request({ url: `/admins/${id}/role`, method: 'put', data })
}

// 获取管理员 passkey 列表
export function getManagerPasskeys(id) {
  return request({ url: `/admins/${id}/passkeys`, method: 'get' })
}

// 删除管理员 passkey
export function deleteManagerPasskey(userId, pkId) {
  return request({ url: `/admins/${userId}/passkeys/${pkId}`, method: 'delete' })
}
