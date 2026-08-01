import request from '@/utils/request'

// ============ 系统配置 ============
export function getSystemConfig() {
  return request({ url: '/system/config', method: 'get' })
}
export function updateSystemConfig(data) {
  return request({ url: '/system/config', method: 'put', data })
}

// ============ 操作日志 ============
export function getOperationLogModules() {
  return request({ url: '/operation_log/modules', method: 'get' })
}
export function getOperationLogList(params) {
  return request({ url: '/operation_log/list', method: 'get', params })
}
export function exportOperationLog(params) {
  return request({ url: '/operation_log/export', method: 'get', params, responseType: 'blob' })
}

// ============ 接口监控 ============
export function getApiMonitorIndex() {
  return request({ url: '/api_monitor/index', method: 'get' })
}
export function getApiMonitorTrend(params) {
  return request({ url: '/api_monitor/trend', method: 'get', params })
}
export function setApiMonitorThreshold(data) {
  return request({ url: '/api_monitor/threshold', method: 'post', data })
}
export function resetApiMonitor(data) {
  return request({ url: '/api_monitor/reset', method: 'post', data })
}
export function getApiMonitorSlowQueryList(params) {
  return request({ url: '/api_monitor/slow_query/list', method: 'get', params })
}
