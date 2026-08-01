import request from '@/utils/request'

// ============ 提现管理 ============
export function getWithdrawList(params) {
  return request({ url: '/finance/withdraws', method: 'get', params })
}
export function approveWithdraw(id) {
  return request({ url: `/finance/withdraws/${id}/approve`, method: 'post' })
}
export function rejectWithdraw(id, data) {
  return request({ url: `/finance/withdraws/${id}/reject`, method: 'post', data })
}
export function verifyBankCard(data) {
  return request({ url: '/finance/withdraws/verify-bank-card', method: 'post', data })
}

// ============ 批量提现 ============
export function getWithdrawBatchList(params) {
  return request({ url: '/finance/withdraw_batch_list', method: 'get', params })
}
export function createWithdrawBatch(data) {
  return request({ url: '/finance/withdraw_batch_create', method: 'post', data })
}
export function processWithdrawBatch(data) {
  return request({ url: '/finance/withdraw_batch_process', method: 'put', data })
}
export function completeWithdrawBatch(data) {
  return request({ url: '/finance/withdraw_batch_complete', method: 'put', data })
}

// ============ 财务配置 ============
export function getFinanceConfig() {
  return request({ url: '/finance/config', method: 'get' })
}
export function updateFinanceConfig(data) {
  return request({ url: '/finance/config', method: 'put', data })
}

// ============ 分账规则/记录 ============
export function getProfitShareRuleList(params) {
  return request({ url: '/profit_share/rule_list', method: 'get', params })
}
export function createProfitShareRule(data) {
  return request({ url: '/profit_share/rule_create', method: 'post', data })
}
export function updateProfitShareRule(data) {
  return request({ url: '/profit_share/rule_update', method: 'put', data })
}
export function toggleProfitShareRule(id) {
  return request({ url: '/profit_share/rule_toggle', method: 'put', data: { id } })
}
export function deleteProfitShareRule(id) {
  return request({ url: '/profit_share/rule_delete', method: 'delete', params: { id } })
}
export function getProfitShareRecordList(params) {
  return request({ url: '/profit_share/record_list', method: 'get', params })
}
export function settleProfitShareRecord(id) {
  return request({ url: '/profit_share/record_settle', method: 'put', data: { id } })
}
export function batchSettleProfitShareRecord(data) {
  return request({ url: '/profit_share/record_batch_settle', method: 'post', data })
}

// ============ 个税配置 ============
export function getTaxConfigList() {
  return request({ url: '/tax/config_list', method: 'get' })
}
export function updateTaxConfig(data) {
  return request({ url: '/tax/config_update', method: 'put', data })
}
export function getTaxRecordList(params) {
  return request({ url: '/tax/record_list', method: 'get', params })
}
export function completeTaxRecord(data) {
  return request({ url: '/tax/record_complete', method: 'put', data })
}
