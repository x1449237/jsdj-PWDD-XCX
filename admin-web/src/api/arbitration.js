import request from '@/utils/request'

// ============ 仲裁案件 ============
export function getArbitrationCaseList(params) {
  return request({ url: '/arbitration/case_list', method: 'get', params })
}
export function getArbitrationCaseDetail(params) {
  return request({ url: '/arbitration/case_detail', method: 'get', params })
}
export function processArbitrationCase(data) {
  return request({ url: '/arbitration/process_case', method: 'post', data })
}
export function resolveArbitrationCase(data) {
  return request({ url: '/arbitration/resolve_case', method: 'post', data })
}

// ============ 判责规则库 ============
export function getArbitrationRuleList(params) {
  return request({ url: '/arbitration/rule_list', method: 'get', params })
}
export function createArbitrationRule(data) {
  return request({ url: '/arbitration/rule_create', method: 'post', data })
}
export function updateArbitrationRule(data) {
  return request({ url: '/arbitration/rule_update', method: 'put', data })
}
export function deleteArbitrationRule(data) {
  return request({ url: '/arbitration/rule_delete', method: 'delete', data })
}

// ============ 举证模板 ============
export function getEvidenceTplList(params) {
  return request({ url: '/arbitration/evidence_tpl_list', method: 'get', params })
}
export function createEvidenceTpl(data) {
  return request({ url: '/arbitration/evidence_tpl_create', method: 'post', data })
}
export function updateEvidenceTpl(data) {
  return request({ url: '/arbitration/evidence_tpl_update', method: 'put', data })
}
export function deleteEvidenceTpl(data) {
  return request({ url: '/arbitration/evidence_tpl_delete', method: 'delete', data })
}
