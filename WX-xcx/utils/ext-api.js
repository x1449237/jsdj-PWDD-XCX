/**
 * 俱乐部小助手APP对标扩展 - 玩家端 API 聚合
 * 所有路径相对 baseURL(/api/v1)，request 已 resolve(data.data)
 */
const request = require('./request');

// ===== 订单预填模板(#42) =====
const listOrderTemplates = () => request.get('/order-ext/templates');
const createOrderTemplate = (data) => request.post('/order-ext/templates', data);
const deleteOrderTemplate = (id) => request.del(`/order-ext/templates/${id}`);

// ===== 订单补充需求/备注(#43 #54) =====
const createSupplement = (data) => request.post('/order-ext/supplements', data);
const listSupplements = (orderId) => request.get('/order-ext/supplements', { order_id: orderId });
const addOrderRemark = (data) => request.post('/order-ext/remarks', data);
const listOrderRemarks = (orderId) => request.get('/order-ext/remarks', { order_id: orderId });

// ===== 部分退款(#46) =====
const createPartialRefund = (data) => request.post('/order-ext/partial-refunds', data);
const listPartialRefunds = (orderId) => request.get('/order-ext/partial-refunds', { order_id: orderId });

// ===== 订单改价/转单/延时申请(#27 #28 #58) =====
const createOrderExtension = (data) => request.post('/order-ext/extensions', data);
const createOrderTransfer = (data) => request.post('/order-ext/transfers', data);
const createOrderPriceChange = (data) => request.post('/order-ext/price-changes', data);

// ===== 退会/请假/资料变更申报(#9 #23 #24) =====
const createResignation = (data) => request.post('/user/club-resignation', data);
const createLeave = (data) => request.post('/user/club-leave', data);
const createChangeRequest = (data) => request.post('/user/club-change-request', data);

// ===== 收藏俱乐部(#75) =====
const favoriteClub = (clubId) => request.post('/user/favorite-clubs', { club_id: clubId });
const unfavoriteClub = (clubId) => request.del(`/user/favorite-clubs/${clubId}`);
const listFavoriteClubs = (page = 1, size = 20) => request.get('/user/favorite-clubs', { page, size });

// ===== 钱包变动记录(#121) =====
const listWalletLogs = (params) => request.get('/user/wallet-logs', params);

// ===== 预存存单(#111) =====
const listDeposits = (page = 1, size = 20) => request.get('/user/deposits', { page, size });
const createDeposit = (data) => request.post('/user/deposits', data);

// ===== 意见反馈(#160) =====
const createFeedback = (data) => request.post('/user/feedbacks', data);
const listMyFeedbacks = (page = 1, size = 20) => request.get('/user/feedbacks', { page, size });

// ===== 拉黑打手(#175) =====
const blockPlayer = (blockedUid) => request.post('/user/blocklist', { blocked_uid: blockedUid });
const unblockPlayer = (playerId) => request.del(`/user/blocklist/${playerId}`);

// ===== 通知设置(#162) =====
const getNotificationSettings = () => request.get('/user/notification-settings');
const updateNotificationSettings = (data) => request.put('/user/notification-settings', data);

// ===== 活动弹窗(#179) =====
const listActivityPopups = () => request.get('/activity-popups');

// ===== 聊天扩展(#102 会话举报) =====
const createChatReport = (data) => request.post('/chat-ext/reports', data);
const togglePinSession = (sessionId) => request.put(`/chat-ext/sessions/${sessionId}/pin`);
const listGroupFiles = (groupId, page = 1, size = 20) => request.get('/chat-ext/group-files', { groupId, page, size });

module.exports = {
  listOrderTemplates, createOrderTemplate, deleteOrderTemplate,
  createSupplement, listSupplements, addOrderRemark, listOrderRemarks,
  createPartialRefund, listPartialRefunds,
  createOrderExtension, createOrderTransfer, createOrderPriceChange,
  createResignation, createLeave, createChangeRequest,
  favoriteClub, unfavoriteClub, listFavoriteClubs,
  listWalletLogs,
  listDeposits, createDeposit,
  createFeedback, listMyFeedbacks,
  blockPlayer, unblockPlayer,
  getNotificationSettings, updateNotificationSettings,
  listActivityPopups,
  createChatReport, togglePinSession, listGroupFiles
};
