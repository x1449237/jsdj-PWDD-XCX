/**
 * 架构规则：小程序前端不含任何业务逻辑。
 * 仅负责：调用后端 API（request）、渲染后端返回字段（*_text/*_color/*_masked/amount_text/time_text 等）、纯 UI 反馈（toast/loading/非空提示）。
 * 禁止：状态/类型硬编码映射、金额换算、时间格式化、脱敏、按月分组、权限判断。
 */
const request = require('../../utils/request');

Page({
  data: {
    totalCommission: '0.00',
    pendingCount: 0,
    settledCount: 0,
    firstOrderCount: 0,
    commissionList: [],
    statusFilter: ['全部', '待结算', '已结算'],
    statusIndex: 0,
    page: 1,
    pageSize: 20,
    hasMore: true,
    loading: true,
    loadingMore: false
  },

  onLoad() {
    this.loadCommissionSummary();
    this.loadCommissionList();
  },

  async loadCommissionSummary() {
    try {
      const res = await request.get('/distributor/commission/summary');
      this.setData({
        totalCommission: res.total_commission_text || res.totalCommission || '0.00',
        pendingCount: res.pendingCount || 0,
        settledCount: res.settledCount || 0,
        firstOrderCount: res.firstOrderCount || 0
      });
    } catch (err) {
      // 忽略
    }
  },

  async loadCommissionList() {
    this.setData({ loading: true });
    try {
      const res = await request.get('/distributor/commission/list', {
        page: this.data.page,
        pageSize: this.data.pageSize,
        status: this.data.statusIndex === 0 ? '' : this.data.statusIndex
      });

      const list = (res.list || []).map(item => this.formatCommissionItem(item));
      this.setData({
        commissionList: list,
        hasMore: res.hasMore !== false,
        loading: false
      });
    } catch (err) {
      this.setData({ loading: false });
    }
  },

  async onReachBottom() {
    if (!this.data.hasMore || this.data.loadingMore) return;
    this.setData({ loadingMore: true });
    try {
      const nextPage = this.data.page + 1;
      const res = await request.get('/distributor/commission/list', {
        page: nextPage,
        pageSize: this.data.pageSize,
        status: this.data.statusIndex === 0 ? '' : this.data.statusIndex
      });

      const list = (res.list || []).map(item => this.formatCommissionItem(item));
      this.setData({
        commissionList: [...this.data.commissionList, ...list],
        page: nextPage,
        hasMore: res.hasMore !== false,
        loadingMore: false
      });
    } catch (err) {
      this.setData({ loadingMore: false });
    }
  },

  formatCommissionItem(item) {
    return {
      ...item,
      sourceAvatar: item.sourceAvatar || '/assets/images/default-avatar.png',
      sourceName: item.source_name_masked || item.sourceName || '',
      orderNo: item.orderNo || item.order_no || '',
      amountText: item.amount_text || '',
      rateText: item.rate_text || '',
      timeText: item.time_text || '',
      statusText: item.status_text || '',
      isFirstOrder: item.isFirstOrder || false
    };
  },

  /* ========== 状态筛选 ========== */
  onStatusChange(e) {
    const index = parseInt(e.detail.value);
    this.setData({
      statusIndex: index,
      page: 1,
      commissionList: [],
      hasMore: true
    });
    this.loadCommissionList();
  }
});
