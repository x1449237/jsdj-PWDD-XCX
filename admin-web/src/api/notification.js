import request from '@/utils/request'

// ============ 订阅消息模板 ============
export function getSubscribeTemplateList(params) {
  return request({ url: '/subscribe/template/list', method: 'get', params })
}
export function createSubscribeTemplate(data) {
  return request({ url: '/subscribe/template/create', method: 'post', data })
}
export function updateSubscribeTemplate(id, data) {
  return request({ url: `/subscribe/template/update/${id}`, method: 'put', data })
}
export function toggleSubscribeTemplate(id, data) {
  return request({ url: `/subscribe/template/toggle/${id}`, method: 'put', data })
}
export function deleteSubscribeTemplate(id) {
  return request({ url: `/subscribe/template/delete/${id}`, method: 'delete' })
}

// ============ 推送日志 ============
export function getSubscribeLogList(params) {
  return request({ url: '/subscribe/log/list', method: 'get', params })
}
