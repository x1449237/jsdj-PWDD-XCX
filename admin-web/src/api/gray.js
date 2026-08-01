import request from '@/utils/request'

// 灰度发布列表
export function getGrayReleaseList(params) {
  return request({ url: '/gray-release', method: 'get', params })
}

// 创建灰度发布
export function createGrayRelease(data) {
  return request({ url: '/gray-release', method: 'post', data })
}

// 发布灰度
export function publishGrayRelease(id) {
  return request({ url: `/gray-release/${id}/publish`, method: 'put' })
}

// 全量发布
export function fullReleaseGray(id) {
  return request({ url: `/gray-release/${id}/full`, method: 'put' })
}

// 回滚灰度
export function rollbackGrayRelease(id) {
  return request({ url: `/gray-release/${id}/rollback`, method: 'put' })
}

// 删除灰度发布
export function deleteGrayRelease(id) {
  return request({ url: `/gray-release/${id}`, method: 'delete' })
}
