import request from '@/utils/request'

// 处罚记录列表
export function getPunishmentRecords(params) {
  return request({ url: '/punishment/records', method: 'get', params })
}

// 撤销处罚
export function revokePunishment(id) {
  return request({ url: `/punishment/records/${id}/revoke`, method: 'post' })
}
