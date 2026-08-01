import request from '@/utils/request'

// ============ 聊天审计会话 ============
export function getChatSessions(params) {
  return request({ url: '/chat-audit/sessions', method: 'get', params })
}
export function getChatMessages(sessionId, params) {
  return request({ url: `/chat-audit/sessions/${sessionId}/messages`, method: 'get', params })
}
export function getRiskUsers(params) {
  return request({ url: '/chat-audit/risk-users', method: 'get', params })
}
export function processRiskUser(id, data) {
  return request({ url: `/chat-audit/risk-users/${id}/process`, method: 'post', data })
}

// ============ 飞单风控规则 ============
export function getAntiFraudRules(params) {
  return request({ url: '/chat/anti_fraud_rules', method: 'get', params })
}
export function createAntiFraudRule(data) {
  return request({ url: '/chat/anti_fraud_rule', method: 'post', data })
}
export function updateAntiFraudRule(data) {
  return request({ url: '/chat/anti_fraud_rule', method: 'put', data })
}
export function deleteAntiFraudRule(data) {
  return request({ url: '/chat/anti_fraud_rule', method: 'delete', data })
}
export function getAntiFraudLogs(params) {
  return request({ url: '/chat/anti_fraud_logs', method: 'get', params })
}
export function handleAntiFraudLog(data) {
  return request({ url: '/chat/anti_fraud_log_handle', method: 'put', data })
}

// ============ 快捷卡片 ============
export function getQuickCards(params) {
  return request({ url: '/chat/quick_cards', method: 'get', params })
}
export function createQuickCard(data) {
  return request({ url: '/chat/quick_card', method: 'post', data })
}
export function updateQuickCard(data) {
  return request({ url: '/chat/quick_card', method: 'put', data })
}
export function deleteQuickCard(data) {
  return request({ url: '/chat/quick_card', method: 'delete', data })
}

// ============ 群聊监察 ============
export function getGroupList(params) {
  return request({ url: '/group-monitor/groups', method: 'get', params })
}
export function getGroupDetail(id) {
  return request({ url: `/group-monitor/groups/${id}`, method: 'get' })
}
export function getGroupMessages(id, params) {
  return request({ url: `/group-monitor/groups/${id}/messages`, method: 'get', params })
}
export function disbandGroup(id) {
  return request({ url: `/group-monitor/groups/${id}/disband`, method: 'post' })
}
export function muteGroup(id, data) {
  return request({ url: `/group-monitor/groups/${id}/mute`, method: 'post', data })
}
export function unmuteGroup(id, data) {
  return request({ url: `/group-monitor/groups/${id}/unmute`, method: 'post', data })
}
export function removeGroupMember(id, data) {
  return request({ url: `/group-monitor/groups/${id}/remove-member`, method: 'post', data })
}
export function banGroupUser(userId) {
  return request({ url: `/group-monitor/users/${userId}/ban`, method: 'post' })
}
export function freezeGroupUserFunds(userId) {
  return request({ url: `/group-monitor/users/${userId}/freeze-funds`, method: 'post' })
}

// ============ 售后关键词 ============
export function getAfterSaleKeywords(params) {
  return request({ url: '/after-sale/keywords', method: 'get', params })
}
export function getAfterSaleKeywordsSwitchStatus() {
  return request({ url: '/after-sale/keywords/switch-status', method: 'get' })
}
export function setAfterSaleKeywordsGlobalSwitch(data) {
  return request({ url: '/after-sale/keywords/global-switch', method: 'post', data })
}
export function setAfterSaleKeywordsTestMode(data) {
  return request({ url: '/after-sale/keywords/test-mode', method: 'post', data })
}
export function createAfterSaleKeyword(data) {
  return request({ url: '/after-sale/keywords', method: 'post', data })
}
export function updateAfterSaleKeyword(id, data) {
  return request({ url: `/after-sale/keywords/${id}`, method: 'put', data })
}
export function deleteAfterSaleKeyword(id) {
  return request({ url: `/after-sale/keywords/${id}`, method: 'delete' })
}
export function importAfterSaleKeywords(data) {
  return request({ url: '/after-sale/keywords/import', method: 'post', data })
}
export function testAfterSaleMatch(data) {
  return request({ url: '/after-sale/keywords/test-match', method: 'post', data })
}

// ============ 售后介入 ============
export function getAfterSaleSessions(params) {
  return request({ url: '/after-sale/sessions', method: 'get', params })
}
export function getAfterSaleSessionDetail(sessionId) {
  return request({ url: `/after-sale/sessions/${sessionId}`, method: 'get' })
}
export function getAfterSaleSessionRecords(sessionId) {
  return request({ url: `/after-sale/sessions/${sessionId}/records`, method: 'get' })
}
export function processAfterSaleSession(sessionId, data) {
  return request({ url: `/after-sale/sessions/${sessionId}/process`, method: 'post', data })
}
export function exportAfterSaleSessions(params) {
  return request({ url: '/after-sale/sessions/export', method: 'get', params, responseType: 'blob' })
}
