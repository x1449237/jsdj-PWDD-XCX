import request from '@/utils/request'

// ============ 俱乐部保证金 ============
export function getClubDepositList(params) {
  return request({ url: '/club/deposit_list', method: 'get', params })
}
export function confirmClubDeposit(data) {
  return request({ url: '/club/confirm_deposit', method: 'put', data })
}
export function refundClubDeposit(data) {
  return request({ url: '/club/refund_deposit', method: 'put', data })
}

// ============ 保证金阶梯 ============
export function getDepositTierList(params) {
  return request({ url: '/club/deposit-tier/list', method: 'get', params })
}
export function createDepositTier(data) {
  return request({ url: '/club/deposit-tier/create', method: 'post', data })
}
export function updateDepositTier(data) {
  return request({ url: '/club/deposit-tier/update', method: 'post', data })
}
export function deleteDepositTier(data) {
  return request({ url: '/club/deposit-tier/delete', method: 'post', data })
}

// ============ 服务保证金 ============
export function getServiceDepositList(params) {
  return request({ url: '/service_deposit/list', method: 'get', params })
}
export function getServiceDepositLogList(params) {
  return request({ url: '/service_deposit/log_list', method: 'get', params })
}
export function manualDepositServiceDeposit(data) {
  return request({ url: '/service_deposit/manual_deposit', method: 'post', data })
}
export function manualDeductServiceDeposit(data) {
  return request({ url: '/service_deposit/manual_deduct', method: 'post', data })
}
export function manualRefundServiceDeposit(data) {
  return request({ url: '/service_deposit/manual_refund', method: 'post', data })
}
export function freezeServiceDeposit(data) {
  return request({ url: '/service_deposit/freeze', method: 'post', data })
}
export function unfreezeServiceDeposit(data) {
  return request({ url: '/service_deposit/unfreeze', method: 'post', data })
}
