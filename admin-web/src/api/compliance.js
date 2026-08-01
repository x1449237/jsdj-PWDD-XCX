import request from '@/utils/request'

// ============ 代练拦截规则 ============
export function getAntiBoostingRuleList(params) {
  return request({ url: '/compliance/anti_boosting_rule_list', method: 'get', params })
}
export function createAntiBoostingRule(data) {
  return request({ url: '/compliance/anti_boosting_rule_create', method: 'post', data })
}
export function updateAntiBoostingRule(data) {
  return request({ url: '/compliance/anti_boosting_rule_update', method: 'put', data })
}
export function deleteAntiBoostingRule(data) {
  return request({ url: '/compliance/anti_boosting_rule_delete', method: 'delete', data })
}
export function expandSensitiveWords() {
  return request({ url: '/compliance/expand_sensitive_words', method: 'post' })
}
export function getAntiBoostingLogList(params) {
  return request({ url: '/compliance/anti_boosting_log_list', method: 'get', params })
}
export function handleAntiBoostingLog(data) {
  return request({ url: '/compliance/anti_boosting_log_handle', method: 'post', data })
}

// ============ 协议版本 ============
export function getAgreementVersionList(params) {
  return request({ url: '/compliance/agreement_version_list', method: 'get', params })
}
export function publishAgreementVersion(data) {
  return request({ url: '/compliance/agreement_version_publish', method: 'post', data })
}
export function createAgreementVersion(data) {
  return request({ url: '/compliance/agreement_version_create', method: 'post', data })
}

// ============ 未成年保护 ============
export function getGuardianList(params) {
  return request({ url: '/minor/guardian_list', method: 'get', params })
}
export function forceUnbindGuardian(data) {
  return request({ url: '/minor/force_unbind', method: 'post', data })
}
export function getMinorWarningLog(params) {
  return request({ url: '/minor/warning_log', method: 'get', params })
}
export function getCurfewConfig() {
  return request({ url: '/minor/curfew_config', method: 'get' })
}
export function getCurfewStats(params) {
  return request({ url: '/minor/curfew_stats', method: 'get', params })
}
export function updateCurfewConfig(data) {
  return request({ url: '/minor/curfew_config', method: 'put', data })
}
