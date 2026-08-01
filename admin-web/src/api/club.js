import request from '@/utils/request'

// ============ 俱乐部列表/详情(与Go后端/admin/clubs/*路由对齐) ============
export function getClubList(params) {
  return request({ url: '/admin/clubs/audit', method: 'get', params })
}
export function getClubListFiltered(params) {
  return request({ url: '/admin/clubs/audit-filter', method: 'get', params })
}
export function getClubDetail(id) {
  return request({ url: `/admin/clubs/${id}`, method: 'get' })
}
export function approveClub(id) {
  return request({ url: `/admin/clubs/${id}/approve`, method: 'post' })
}
export function rejectClub(id, data) {
  return request({ url: `/admin/clubs/${id}/reject`, method: 'post', data })
}
export function freezeClub(data) {
  return request({ url: `/admin/clubs/${data.id}/freeze`, method: 'post', { reason: data.reason }) }
}
export function unfreezeClub(data) {
  return request({ url: `/admin/clubs/${data.id}/unfreeze`, method: 'post' })
}
export function cancelClub(data) {
  return request({ url: `/admin/clubs/${data.id}/cancel`, method: 'post', { reason: data.reason }) }
}

// ============ V标手动隐藏/恢复 ============
export function hideClubVBadge(id) {
  return request({ url: `/admin/clubs/${id}/vbadge/hide`, method: 'post' })
}
export function restoreClubVBadge(id) {
  return request({ url: `/admin/clubs/${id}/vbadge/restore`, method: 'post' })
}

// ============ 俱乐部资料修改日志 ============
export function getClubChangeLogs(id, params) {
  return request({ url: `/admin/clubs/${id}/change-logs`, method: 'get', params })
}

// ============ 对公打款验证 ============
export function getClubTransferList(params) {
  return request({ url: '/admin/corporate-transfers', method: 'get', params })
}
export function generateClubTransfer(data) {
  return request({ url: '/admin/corporate-transfers', method: 'post', data })
}
export function verifyClubTransfer(id, data) {
  return request({ url: `/admin/corporate-transfers/${id}/verify`, method: 'post', data })
}
export function exportClubTransferLedger(params) {
  return request({ url: '/admin/corporate-transfers/export', method: 'get', params, responseType: 'blob' })
}

// ============ 小额打款独立台账 ============
export function getClubTransferLedger(params) {
  return request({ url: '/admin/corporate-transfers', method: 'get', params })
}

// ============ 入驻附件管理 ============
export function downloadClubAttachment(params) {
  return request({ url: '/admin/clubs/attachment/download', method: 'get', params, responseType: 'blob' })
}
export function exportClubAttachment(params) {
  return request({ url: '/admin/clubs/attachment/export', method: 'get', params, responseType: 'blob' })
}

// ============ 保证金管理（个人/企业双参数配置） ============
export function getDepositConfig() {
  return request({ url: '/admin/deposits/config', method: 'get' })
}
export function updateDepositConfig(data) {
  return request({ url: '/admin/deposits/config', method: 'put', data })
}
export function getDepositDeductList(params) {
  return request({ url: '/admin/deposits', method: 'get', params })
}
export function getDepositRepayMonitor(params) {
  return request({ url: '/admin/deposits/repay-monitor', method: 'get', params })
}
export function confirmDeposit(clubId, data) {
  return request({ url: `/admin/deposits/${clubId}/confirm`, method: 'post', data })
}
export function refundDeposit(clubId, data) {
  return request({ url: `/admin/deposits/${clubId}/refund`, method: 'post', data })
}

// ============ 罚款规则备案管理 ============
export function getFineRuleList(params) {
  return request({ url: '/admin/fine-rules', method: 'get', params })
}
export function reviewFineRule(id, data) {
  return request({ url: `/admin/fine-rules/${id}/review`, method: 'post', data })
}
export function revokeFineRule(data) {
  return request({ url: '/admin/fine-rules/revoke', method: 'post', data })
}

// ============ 运营数据看板 ============
export function getClubOperationData(params) {
  return request({ url: '/admin/clubs/operation-data', method: 'get', params })
}

// ============ 内部订单监控 ============
export function getClubInternalOrderList(params) {
  return request({ url: '/admin/clubs/internal-order/list', method: 'get', params })
}
export function getClubInternalOrderDetail(params) {
  return request({ url: '/admin/clubs/internal-order/detail', method: 'get', params })
}
