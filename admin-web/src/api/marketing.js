import request from '@/utils/request'

// ============ 活动运营 ============
export function getActivityIndex() {
  return request({ url: '/activity/index', method: 'get' })
}
export function getActivityCouponList(params) {
  return request({ url: '/activity/coupon/list', method: 'get', params })
}
export function createActivityCoupon(data) {
  return request({ url: '/activity/coupon/create', method: 'post', data })
}
export function updateActivityCoupon(data) {
  return request({ url: '/activity/coupon/update', method: 'post', data })
}
export function toggleActivityCoupon(data) {
  return request({ url: '/activity/coupon/toggle', method: 'post', data })
}
export function deleteActivityCoupon(id) {
  return request({ url: '/activity/coupon/delete', method: 'delete', params: { id } })
}

// ============ 营销-优惠券 ============
export function getCouponList(params) {
  return request({ url: '/marketing/coupon/list', method: 'get', params })
}
export function createCoupon(data) {
  return request({ url: '/marketing/coupon/create', method: 'post', data })
}
export function updateCoupon(data) {
  return request({ url: '/marketing/coupon/update', method: 'post', data })
}
export function toggleCoupon(id) {
  return request({ url: '/marketing/coupon/toggle', method: 'put', data: { id } })
}
export function deleteCoupon(id) {
  return request({ url: '/marketing/coupon/delete', method: 'delete', data: { id } })
}

// ============ 营销-充值活动 ============
export function getRechargeList(params) {
  return request({ url: '/marketing/recharge/list', method: 'get', params })
}
export function createRecharge(data) {
  return request({ url: '/marketing/recharge/create', method: 'post', data })
}
export function updateRecharge(data) {
  return request({ url: '/marketing/recharge/update', method: 'post', data })
}
export function toggleRecharge(id) {
  return request({ url: '/marketing/recharge/toggle', method: 'put', data: { id } })
}
export function deleteRecharge(id) {
  return request({ url: '/marketing/recharge/delete', method: 'delete', data: { id } })
}

// ============ 营销-抽奖活动 ============
export function getLotteryList(params) {
  return request({ url: '/marketing/lottery/list', method: 'get', params })
}
export function getLotteryDetail(params) {
  return request({ url: '/marketing/lottery/detail', method: 'get', params })
}
export function createLottery(data) {
  return request({ url: '/marketing/lottery/create', method: 'post', data })
}
export function updateLottery(data) {
  return request({ url: '/marketing/lottery/update', method: 'post', data })
}
export function saveLotteryPrizes(data) {
  return request({ url: '/marketing/lottery/save_prizes', method: 'post', data })
}
export function toggleLottery(id) {
  return request({ url: '/marketing/lottery/toggle', method: 'put', data: { id } })
}
export function deleteLottery(id) {
  return request({ url: '/marketing/lottery/delete', method: 'delete', data: { id } })
}

// ============ 营销-拼团活动 ============
export function getGroupBuyList(params) {
  return request({ url: '/marketing/group_buy/list', method: 'get', params })
}
export function createGroupBuy(data) {
  return request({ url: '/marketing/group_buy/create', method: 'post', data })
}
export function updateGroupBuy(data) {
  return request({ url: '/marketing/group_buy/update', method: 'post', data })
}
export function toggleGroupBuy(id) {
  return request({ url: '/marketing/group_buy/toggle', method: 'put', data: { id } })
}
export function deleteGroupBuy(id) {
  return request({ url: '/marketing/group_buy/delete', method: 'delete', data: { id } })
}

// ============ 营销-邀请奖励 ============
export function getInviteRewardList(params) {
  return request({ url: '/marketing/invite_reward/list', method: 'get', params })
}
export function createInviteReward(data) {
  return request({ url: '/marketing/invite_reward/create', method: 'post', data })
}
export function updateInviteReward(data) {
  return request({ url: '/marketing/invite_reward/update', method: 'post', data })
}
export function toggleInviteReward(id) {
  return request({ url: '/marketing/invite_reward/toggle', method: 'put', data: { id } })
}
export function deleteInviteReward(id) {
  return request({ url: '/marketing/invite_reward/delete', method: 'delete', data: { id } })
}
