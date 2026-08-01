import request from '@/utils/request'

// 申诉列表
export function getAppealList(params) {
  return request({ url: '/appeals', method: 'get', params })
}

// 申诉详情
export function getAppealDetail(id) {
  return request({ url: `/appeals/${id}`, method: 'get' })
}

// 完成申诉
export function completeAppeal(id) {
  return request({ url: `/appeals/${id}/complete`, method: 'put' })
}

// 新增申诉沟通
export function addAppealCommunication(id, data) {
  return request({ url: `/appeals/${id}/communications`, method: 'post', data })
}
