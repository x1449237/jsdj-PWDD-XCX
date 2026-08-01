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

// ============ 对公打款验证 ============
export function getClubTransferList(params) {
  return request({ url: '/club/transfer_list', method: 'get', params })
}
export function verifyClubTransfer(data) {
  return request({ url: '/club/verify_transfer', method: 'put', data })
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
