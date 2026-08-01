import request from '@/utils/request'

// UP主认证列表
export function getUpMasterList(params) {
  return request({ url: '/up_master/list', method: 'get', params })
}

// UP主认证详情
export function getUpMasterDetail(params) {
  return request({ url: '/up_master/detail', method: 'get', params })
}

// 通过 UP主认证
export function approveUpMaster(data) {
  return request({ url: '/up_master/approve', method: 'post', data })
}

// 拒绝 UP主认证
export function rejectUpMaster(data) {
  return request({ url: '/up_master/reject', method: 'post', data })
}

// 撤销 UP主认证
export function revokeUpMaster(data) {
  return request({ url: '/up_master/revoke', method: 'post', data })
}
