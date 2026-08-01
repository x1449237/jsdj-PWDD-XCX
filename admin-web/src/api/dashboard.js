import request from '@/utils/request'

// 仪表盘统计
export function getStats() {
  return request({ url: '/dashboard/stats', method: 'get' })
}

// 数据大屏-实时数据
export function getRealtimeData() {
  return request({ url: '/data_dashboard/realtime', method: 'get' })
}

// 数据大屏-订单趋势
export function getOrderTrend(params) {
  return request({ url: '/data_dashboard/order_trend', method: 'get', params })
}

// 数据大屏-资金流水
export function getFundFlow(params) {
  return request({ url: '/data_dashboard/fund_flow', method: 'get', params })
}
