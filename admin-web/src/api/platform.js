import request from '@/utils/request'

// 平台官方账号列表
export function getPlatformAccounts(params) {
  return request({ url: '/platform/accounts', method: 'get', params })
}

// 新增平台账号
export function createPlatformAccount(data) {
  return request({ url: '/platform/accounts', method: 'post', data })
}

// 更新平台账号
export function updatePlatformAccount(id, data) {
  return request({ url: `/platform/accounts/${id}`, method: 'put', data })
}

// 禁用平台账号
export function disablePlatformAccount(id) {
  return request({ url: `/platform/accounts/${id}/disable`, method: 'post' })
}

// 启用平台账号
export function enablePlatformAccount(id) {
  return request({ url: `/platform/accounts/${id}/enable`, method: 'post' })
}
