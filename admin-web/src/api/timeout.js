import request from '@/utils/request'

// ============ 预置超时规则 ============
export function getPresetTimeoutRules() {
  return request({ url: '/timeout-rules/preset', method: 'get' })
}
export function togglePresetTimeoutRule(id, data) {
  return request({ url: `/timeout-rules/preset/${id}/toggle`, method: 'put', data })
}
export function updatePresetTimeoutRule(id, data) {
  return request({ url: `/timeout-rules/preset/${id}`, method: 'put', data })
}

// ============ 自定义超时规则 ============
export function getCustomTimeoutRules() {
  return request({ url: '/timeout-rules/custom', method: 'get' })
}
export function createCustomTimeoutRule(data) {
  return request({ url: '/timeout-rules/custom', method: 'post', data })
}
export function updateCustomTimeoutRule(id, data) {
  return request({ url: `/timeout-rules/custom/${id}`, method: 'put', data })
}
export function toggleCustomTimeoutRule(id, data) {
  return request({ url: `/timeout-rules/custom/${id}/toggle`, method: 'put', data })
}
export function deleteCustomTimeoutRule(id) {
  return request({ url: `/timeout-rules/custom/${id}`, method: 'delete' })
}
