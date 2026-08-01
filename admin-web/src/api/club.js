import request from '@/utils/request'

// ============ 俱乐部列表/详情 ============
export function getClubList(params) {
  return request({ url: '/club/list', method: 'get', params })
}
export function getClubDetail(params) {
  return request({ url: '/club/detail', method: 'get', params })
}
export function freezeClub(data) {
  return request({ url: '/club/freeze', method: 'put', data })
}
export function unfreezeClub(data) {
  return request({ url: '/club/unfreeze', method: 'put', data })
}
export function cancelClub(data) {
  return request({ url: '/club/cancel', method: 'put', data })
}

// ============ V标手动隐藏/恢复 ============
export function hideClubVBadge(id, data) {
  return request({ url: `/club/${id}/hide-vbadge`, method: 'post', data })
}
export function restoreClubVBadge(id, data) {
  return request({ url: `/club/${id}/restore-vbadge`, method: 'post', data })
}

// ============ 对公打款验证 ============
export function getClubTransferList(params) {
  return request({ url: '/club/transfer-list', method: 'get', params })
}
export function generateClubTransfer(data) {
  return request({ url: '/club/generate-transfer', method: 'post', data })
}
export function verifyClubTransfer(data) {
  return request({ url: '/club/verify-transfer', method: 'post', data })
}
export function exportClubTransferLedger(params) {
  return request({ url: '/club/transfer-ledger/export', method: 'get', params, responseType: 'blob' })
}

// ============ 小额打款独立台账 ============
export function getClubTransferLedger(params) {
  return request({ url: '/club/transfer-ledger/list', method: 'get', params })
}

// ============ 入驻附件管理 ============
export function downloadClubAttachment(params) {
  return request({ url: '/club/attachment/download', method: 'get', params, responseType: 'blob' })
}
export function exportClubAttachment(params) {
  return request({ url: '/club/attachment/export', method: 'get', params, responseType: 'blob' })
}

// ============ 保证金管理（个人/企业双参数配置） ============
export function getDepositConfig() {
  return request({ url: '/deposit/config', method: 'get' })
}
export function updateDepositConfig(data) {
  return request({ url: '/deposit/config', method: 'put', data })
}
export function getDepositDeductList(params) {
  return request({ url: '/deposit/deduct-list', method: 'get', params })
}
export function getDepositRepayMonitor(params) {
  return request({ url: '/deposit/repay-monitor', method: 'get', params })
}

// ============ 罚款规则备案管理 ============
export function getFineRuleList(params) {
  return request({ url: '/club/fine-rules', method: 'get', params })
}
export function getFineRuleDetail(params) {
  return request({ url: '/club/fine-rules/detail', method: 'get', params })
}
export function revokeFineRule(data) {
  return request({ url: '/club/fine-rules/revoke', method: 'post', data })
}

// ============ 运营数据看板 ============
export function getClubOperationData(params) {
  return request({ url: '/club/operation_data', method: 'get', params })
}

// ============ 内部订单监控 ============
export function getClubInternalOrderList(params) {
  return request({ url: '/club/internal-order/list', method: 'get', params })
}
export function getClubInternalOrderDetail(params) {
  return request({ url: '/club/internal-order/detail', method: 'get', params })
}
