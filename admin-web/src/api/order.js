import request from '@/utils/request'

// ============ 订单管理 ============
export function getOrderList(params) {
  return request({ url: '/orders', method: 'get', params })
}
export function getOrderDetail(id) {
  return request({ url: `/orders/${id}`, method: 'get' })
}
export function forceOrderStatus(id, data) {
  return request({ url: `/orders/${id}/force-status`, method: 'post', data })
}
export function refundOrder(id, data) {
  return request({ url: `/orders/${id}/refund`, method: 'post', data })
}

// 批量订单
export function batchOrders(data) {
  return request({ url: '/orders/batch', method: 'post', data })
}
export function getBatchConfirm(batchId) {
  return request({ url: `/orders/batch/confirm/${batchId}`, method: 'get' })
}

// 竞价订单
export function getBidOrderList(params) {
  return request({ url: '/order/bid/list', method: 'get', params })
}

// 游戏列表（下拉用）
export function getGameList(params) {
  return request({ url: '/game/list', method: 'get', params })
}

// ============ 订单套餐 ============
export function getOrderPackageList(params) {
  return request({ url: '/order/package/list', method: 'get', params })
}
export function createOrderPackage(data) {
  return request({ url: '/order/package/create', method: 'post', data })
}
export function updateOrderPackage(data) {
  return request({ url: '/order/package/update', method: 'put', data })
}
export function toggleOrderPackage(id) {
  return request({ url: '/order/package/toggle', method: 'put', data: { id } })
}
export function deleteOrderPackage(id) {
  return request({ url: '/order/package/delete', method: 'delete', data: { id } })
}

// ============ 退单规则 ============
export function getRefundRuleList(params) {
  return request({ url: '/order/refund_rule/list', method: 'get', params })
}
export function createRefundRule(data) {
  return request({ url: '/order/refund_rule/create', method: 'post', data })
}
export function updateRefundRule(data) {
  return request({ url: '/order/refund_rule/update', method: 'put', data })
}
export function toggleRefundRule(id) {
  return request({ url: '/order/refund_rule/toggle', method: 'put', data: { id } })
}
export function deleteRefundRule(id) {
  return request({ url: '/order/refund_rule/delete', method: 'delete', data: { id } })
}
