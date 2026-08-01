import request from '@/utils/request'

// 文档列表
export function getDocumentList(params) {
  return request({ url: '/document/list', method: 'get', params })
}

// 上传文档
export function uploadDocument(data) {
  return request({ url: '/document/upload', method: 'post', data })
}

// 替换文档
export function replaceDocument(data) {
  return request({ url: '/document/replace', method: 'put', data })
}

// 文档版本列表
export function getDocumentVersions(params) {
  return request({ url: '/document/versions', method: 'get', params })
}

// 切换文档启用状态
export function toggleDocument(data) {
  return request({ url: '/document/toggle', method: 'put', data })
}

// 删除文档
export function deleteDocument(data) {
  return request({ url: '/document/delete', method: 'delete', data })
}
