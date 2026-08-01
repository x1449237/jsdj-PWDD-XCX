import { createRouter, createWebHashHistory } from 'vue-router'
import { getToken, getNeedInit } from '@/utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', noAuth: true }
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('@/views/init/index.vue'),
    meta: { title: '初始化', noAuth: true }
  },
  {
    path: '/',
    component: () => import('@/layout/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' }
      },
      {
        path: 'dashboard/bigscreen',
        name: 'Bigscreen',
        component: () => import('@/views/dashboard/bigscreen.vue'),
        meta: { title: '数据大屏', icon: 'DataAnalysis', hidden: true }
      },
      {
        path: 'user/list',
        name: 'UserList',
        component: () => import('@/views/user/list.vue'),
        meta: { title: '用户管理', icon: 'User' }
      },
      {
        path: 'user/detail/:id',
        name: 'UserDetail',
        component: () => import('@/views/user/detail.vue'),
        meta: { title: '用户详情', hidden: true }
      },
      {
        path: 'admin/list',
        name: 'AdminList',
        component: () => import('@/views/admin/list.vue'),
        meta: { title: '管理员管理', icon: 'UserFilled' }
      },
      {
        path: 'audit/player',
        name: 'AuditPlayer',
        component: () => import('@/views/audit/player.vue'),
        meta: { title: '打手审核', icon: 'Checked' }
      },
      {
        path: 'audit/distributor',
        name: 'AuditDistributor',
        component: () => import('@/views/audit/distributor.vue'),
        meta: { title: '分销商审核', icon: 'Checked' }
      },
      {
        path: 'audit/dispatcher',
        name: 'AuditDispatcher',
        component: () => import('@/views/audit/dispatcher.vue'),
        meta: { title: '派单员审核', icon: 'Checked' }
      },
      {
        path: 'audit/admin',
        name: 'AuditAdmin',
        component: () => import('@/views/audit/admin.vue'),
        meta: { title: '内置管理员审核', icon: 'Checked' }
      },
      {
        path: 'audit/club',
        name: 'AuditClub',
        component: () => import('@/views/audit/club.vue'),
        meta: { title: '俱乐部入驻审核', icon: 'Checked' }
      },
      {
        path: 'invite',
        name: 'Invite',
        component: () => import('@/views/invite/index.vue'),
        meta: { title: '邀请码管理', icon: 'Link' }
      },
      {
        path: 'order/list',
        name: 'OrderList',
        component: () => import('@/views/order/list.vue'),
        meta: { title: '订单管理', icon: 'Document' }
      },
      {
        path: 'order/detail/:id',
        name: 'OrderDetail',
        component: () => import('@/views/order/detail.vue'),
        meta: { title: '订单详情', hidden: true }
      },
      {
        path: 'order/package',
        name: 'OrderPackage',
        component: () => import('@/views/order/package.vue'),
        meta: { title: '订单套餐', icon: 'Goods', group: 'order' }
      },
      {
        path: 'order/refund-rule',
        name: 'OrderRefundRule',
        component: () => import('@/views/order/refund-rule.vue'),
        meta: { title: '退单规则', icon: 'RefreshLeft', group: 'order' }
      },
      {
        path: 'order/bid-list',
        name: 'OrderBidList',
        component: () => import('@/views/order/bid-list.vue'),
        meta: { title: '竞价订单', icon: 'Money', group: 'order' }
      },
      {
        path: 'finance/withdraw',
        name: 'FinanceWithdraw',
        component: () => import('@/views/finance/withdraw.vue'),
        meta: { title: '提现管理', icon: 'Money' }
      },
      {
        path: 'finance/config',
        name: 'FinanceConfig',
        component: () => import('@/views/finance/config.vue'),
        meta: { title: '财务配置', icon: 'Setting', group: 'finance' }
      },
      {
        path: 'finance/profit-share-rule',
        name: 'ProfitShareRule',
        component: () => import('@/views/finance/profit-share-rule.vue'),
        meta: { title: '分账规则配置', icon: 'Money', group: 'finance' }
      },
      {
        path: 'finance/profit-share-record',
        name: 'ProfitShareRecord',
        component: () => import('@/views/finance/profit-share-record.vue'),
        meta: { title: '分账记录', icon: 'Document', group: 'finance' }
      },
      {
        path: 'finance/tax-config',
        name: 'TaxConfig',
        component: () => import('@/views/finance/tax-config.vue'),
        meta: { title: '个税配置', icon: 'Wallet', group: 'finance' }
      },
      {
        path: 'finance/withdraw-batch',
        name: 'WithdrawBatch',
        component: () => import('@/views/finance/withdraw-batch.vue'),
        meta: { title: '批量提现', icon: 'Money', group: 'finance' }
      },
      {
        path: 'timeout-rule',
        name: 'TimeoutRule',
        component: () => import('@/views/timeout-rule/index.vue'),
        meta: { title: '超时规则', icon: 'Timer' }
      },
      {
        path: 'chat-audit',
        name: 'ChatAudit',
        component: () => import('@/views/chat-audit/index.vue'),
        meta: { title: '聊天审计', icon: 'ChatDotSquare', group: 'chat' }
      },
      {
        path: 'anti-fraud',
        name: 'AntiFraud',
        component: () => import('@/views/chat-audit/anti-fraud.vue'),
        meta: { title: '飞单风控', icon: 'Warning', group: 'chat' }
      },
      {
        path: 'quick-card',
        name: 'QuickCard',
        component: () => import('@/views/chat-audit/quick-card.vue'),
        meta: { title: '快捷卡片', icon: 'Postcard', group: 'chat' }
      },
      {
        path: 'system/config',
        name: 'SystemConfig',
        component: () => import('@/views/system/config.vue'),
        meta: { title: '系统配置', icon: 'Tools', group: 'system' }
      },
      {
        path: 'system/operation-log',
        name: 'OperationLog',
        component: () => import('@/views/system/operation-log.vue'),
        meta: { title: '操作日志', icon: 'Document', group: 'system' }
      },
      {
        path: 'system/api-monitor',
        name: 'ApiMonitor',
        component: () => import('@/views/system/api-monitor.vue'),
        meta: { title: '接口监控', icon: 'DataLine', group: 'system' }
      },
      {
        path: 'appeal/list',
        name: 'AppealList',
        component: () => import('@/views/appeal/list.vue'),
        meta: { title: '申诉管理', icon: 'Warning' }
      },
      {
        path: 'backup',
        name: 'Backup',
        component: () => import('@/views/backup/index.vue'),
        meta: { title: '备份恢复', icon: 'DataBoard' }
      },
      {
        path: 'document',
        name: 'DocumentManage',
        component: () => import('@/views/document/index.vue'),
        meta: { title: '平台文档管理', icon: 'Document' }
      },
      {
        path: 'gray-release',
        name: 'GrayRelease',
        component: () => import('@/views/gray/index.vue'),
        meta: { title: '灰度发布', icon: 'Connection' }
      },
      {
        path: 'platform/accounts',
        name: 'PlatformAccounts',
        component: () => import('@/views/platform/accounts.vue'),
        meta: { title: '平台官方账号', icon: 'UserFilled', group: 'platform' }
      },
      {
        path: 'platform/subscribe',
        name: 'PlatformSubscribe',
        component: () => import('@/views/platform/subscribe.vue'),
        meta: { title: '订阅消息模板', icon: 'Message', group: 'platform' }
      },
      {
        path: 'platform/up-master',
        name: 'PlatformUpMaster',
        component: () => import('@/views/platform/up-master.vue'),
        meta: { title: 'UP主认证', icon: 'Medal', group: 'platform' }
      },
      {
        path: 'up-master/index',
        name: 'UpMasterIndex',
        component: () => import('@/views/up-master/index.vue'),
        meta: { title: 'UP主管理', icon: 'Cpu' }
      },
      {
        path: 'after-sale/keywords',
        name: 'AfterSaleKeywords',
        component: () => import('@/views/after-sale/keywords.vue'),
        meta: { title: '售后关键词', icon: 'Warning', group: 'platform' }
      },
      {
        path: 'group-monitor',
        name: 'GroupMonitor',
        component: () => import('@/views/group-monitor/index.vue'),
        meta: { title: '群聊监察', icon: 'ChatDotRound', group: 'chat' }
      },
      {
        path: 'after-sale/manage',
        name: 'AfterSaleManage',
        component: () => import('@/views/after-sale/manage.vue'),
        meta: { title: '售后介入', icon: 'Service', group: 'chat' }
      },
      {
        path: 'punishment',
        name: 'PunishmentRecords',
        component: () => import('@/views/punishment/index.vue'),
        meta: { title: '处罚记录', icon: 'Lock', group: 'security' }
      },
      {
        path: 'risk/alert-center',
        name: 'RiskAlertCenter',
        component: () => import('@/views/risk/alert-center.vue'),
        meta: { title: '风险预警', icon: 'Warning', group: 'security' }
      },
      {
        path: 'activity',
        name: 'Activity',
        component: () => import('@/views/activity/index.vue'),
        meta: { title: '活动运营', icon: 'Present', group: 'activity' }
      },
      {
        path: 'marketing/coupon',
        name: 'MarketingCoupon',
        component: () => import('@/views/marketing/coupon.vue'),
        meta: { title: '优惠券管理', icon: 'Discount', group: 'marketing' }
      },
      {
        path: 'marketing/recharge',
        name: 'MarketingRecharge',
        component: () => import('@/views/marketing/recharge.vue'),
        meta: { title: '充值活动', icon: 'Wallet', group: 'marketing' }
      },
      {
        path: 'marketing/lottery',
        name: 'MarketingLottery',
        component: () => import('@/views/marketing/lottery.vue'),
        meta: { title: '抽奖活动', icon: 'Promotion', group: 'marketing' }
      },
      {
        path: 'marketing/group-buy',
        name: 'MarketingGroupBuy',
        component: () => import('@/views/marketing/group-buy.vue'),
        meta: { title: '拼团活动', icon: 'UserFilled', group: 'marketing' }
      },
      {
        path: 'marketing/invite-reward',
        name: 'MarketingInviteReward',
        component: () => import('@/views/marketing/invite-reward.vue'),
        meta: { title: '邀请奖励', icon: 'Share', group: 'marketing' }
      },
      {
        path: 'club/list',
        name: 'ClubList',
        component: () => import('@/views/club/list.vue'),
        meta: { title: '俱乐部列表', icon: 'List', group: 'club' }
      },
      {
        path: 'club/detail/:id',
        name: 'ClubDetail',
        component: () => import('@/views/club/detail.vue'),
        meta: { title: '俱乐部详情', hidden: true, group: 'club' }
      },
      {
        path: 'club/deposit',
        name: 'ClubDeposit',
        component: () => import('@/views/club/deposit.vue'),
        meta: { title: '保证金管理', icon: 'Money', group: 'club' }
      },
      {
        path: 'club/transfer',
        name: 'ClubTransfer',
        component: () => import('@/views/club/transfer.vue'),
        meta: { title: '对公打款验证', icon: 'BankCard', group: 'club' }
      },
      {
        path: 'club/transfer-ledger',
        name: 'ClubTransferLedger',
        component: () => import('@/views/club/transfer-ledger.vue'),
        meta: { title: '小额打款台账', icon: 'Wallet', group: 'club' }
      },
      {
        path: 'club/fine-rules',
        name: 'ClubFineRules',
        component: () => import('@/views/club/fine-rules.vue'),
        meta: { title: '罚款规则备案', icon: 'DocumentChecked', group: 'club' }
      },
      {
        path: 'arbitration/case-list',
        name: 'ArbitrationCaseList',
        component: () => import('@/views/arbitration/case-list.vue'),
        meta: { title: '仲裁案件', icon: 'Document', group: 'arbitration' }
      },
      {
        path: 'arbitration/rule-library',
        name: 'ArbitrationRuleLibrary',
        component: () => import('@/views/arbitration/rule-library.vue'),
        meta: { title: '判责规则库', icon: 'Collection', group: 'arbitration' }
      },
      {
        path: 'arbitration/evidence-tpl',
        name: 'ArbitrationEvidenceTpl',
        component: () => import('@/views/arbitration/evidence-tpl.vue'),
        meta: { title: '举证模板', icon: 'Picture', group: 'arbitration' }
      },
      {
        path: 'finance/service-deposit',
        name: 'ServiceDeposit',
        component: () => import('@/views/finance/service-deposit.vue'),
        meta: { title: '服务保证金', icon: 'Money', group: 'finance' }
      },
      {
        path: 'compliance/anti-boosting',
        name: 'AntiBoosting',
        component: () => import('@/views/compliance/anti-boosting.vue'),
        meta: { title: '代练拦截', icon: 'Warning', group: 'compliance' }
      },
      {
        path: 'compliance/agreement-version',
        name: 'AgreementVersion',
        component: () => import('@/views/compliance/agreement-version.vue'),
        meta: { title: '协议版本', icon: 'Document', group: 'compliance' }
      },
      {
        path: 'minor/curfew-config',
        name: 'MinorCurfewConfig',
        component: () => import('@/views/minor/curfew-config.vue'),
        meta: { title: '宵禁配置', icon: 'Clock', group: 'minor' }
      },
      {
        path: 'minor/warning-log',
        name: 'MinorWarningLog',
        component: () => import('@/views/minor/warning-log.vue'),
        meta: { title: '消费预警日志', icon: 'Warning', group: 'minor' }
      },
      {
        path: 'minor/guardian-list',
        name: 'MinorGuardianList',
        component: () => import('@/views/minor/guardian-list.vue'),
        meta: { title: '监护绑定列表', icon: 'UserFilled', group: 'minor' }
      },
      {
        path: 'club/deposit-tier',
        name: 'ClubDepositTier',
        component: () => import('@/views/club/deposit-tier.vue'),
        meta: { title: '保证金阶梯', icon: 'TrendCharts', group: 'club' }
      },
      {
        path: 'club/operation-data',
        name: 'ClubOperationData',
        component: () => import('@/views/club/operation-data.vue'),
        meta: { title: '运营数据看板', icon: 'DataAnalysis', group: 'club' }
      },
      {
        path: 'club/internal-order',
        name: 'ClubInternalOrder',
        component: () => import('@/views/club/internal-order.vue'),
        meta: { title: '内部订单监控', icon: 'Tickets', group: 'club' }
      }
    ]
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '404', noAuth: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404'
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 超级管理后台` : '超级管理后台'

  const token = getToken()
  if (token) {
    // 已登录但账号尚未完成首次登录初始化，强制进入初始化页
    if (getNeedInit() && to.path !== '/init') {
      next({ path: '/init' })
      return
    }
    // 已完成初始化时，访问登录页则跳转仪表盘
    if (to.path === '/login' && !getNeedInit()) {
      next({ path: '/dashboard' })
      return
    }
    next()
  } else {
    if (to.meta.noAuth) {
      next()
    } else {
      next({ path: '/login', query: { redirect: to.fullPath } })
    }
  }
})

export default router