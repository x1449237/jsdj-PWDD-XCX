import request from '@/utils/request'

// ===== 违规处罚模板 =====
export function getPunishmentTemplates() {
  return request({ url: '/punishment-templates', method: 'get' })
}

export function savePunishmentTemplate(data) {
  return request({ url: '/punishment-templates', method: 'post', data })
}

// ===== 玩家意见反馈 =====
export function getAllFeedbacks(params) {
  return request({ url: '/feedbacks', method: 'get', params })
}

export function replyFeedback(id, data) {
  return request({ url: `/feedbacks/${id}/reply`, method: 'put', data })
}

// ===== 节日模板公告 =====
export function getFestivalTemplates() {
  return request({ url: '/festival-templates', method: 'get' })
}

export function saveFestivalTemplate(data) {
  return request({ url: '/festival-templates', method: 'post', data })
}

// ===== 推广渠道统计 =====
export function getPromoChannels(params) {
  return request({ url: '/promo-channels', method: 'get', params })
}

export function createPromoChannel(data) {
  return request({ url: '/promo-channels', method: 'post', data })
}

// ===== 会话举报 =====
export function getChatReports(params) {
  return request({ url: '/chat-reports', method: 'get', params })
}

export function handleChatReport(id, data) {
  return request({ url: `/chat-reports/${id}/handle`, method: 'put', data })
}
