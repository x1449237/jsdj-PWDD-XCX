import request from '@/utils/request'

// 风险预警统计
export function getRiskAlertStats() {
  return request({ url: '/risk_alert/stats', method: 'get' })
}

// 风险预警列表
export function getRiskAlertList(params) {
  return request({ url: '/risk_alert/list', method: 'get', params })
}

// 处理风险预警
export function handleRiskAlert(data) {
  return request({ url: '/risk_alert/handle', method: 'post', data })
}

// 批量处理风险预警
export function batchHandleRiskAlert(data) {
  return request({ url: '/risk_alert/batch_handle', method: 'post', data })
}

// 批量封禁用户
export function batchBanRiskAlert(data) {
  return request({ url: '/risk_alert/batch_ban', method: 'post', data })
}
