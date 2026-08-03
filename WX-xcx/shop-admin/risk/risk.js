/**
 * 架构规则：前端零业务逻辑。
 * 本页面仅负责：调用后端 API、setData 渲染后端返回字段、纯 UI 反馈。
 * 风险等级/处理状态文案与颜色、时间均由后端返回：
 *   risk_level_text / handle_status_text / handle_status_color / time_text
 */
const request = require('../../utils/request');

Page({
  data: {
    userList: [],
    currentRiskType: '',
    currentRiskLevel: '',
    riskTypeList: [
      { label: '恶意退款', value: 'refund_abuse' },
      { label: '虚假交易', value: 'fake_order' },
      { label: '刷单', value: 'brush_order' },
      { label: '恶意投诉', value: 'malicious_complaint' },
      { label: '信用异常', value: 'credit_abnormal' }
    ],
    riskLevelList: [
      { label: '高', value: 'high' },
      { label: '中', value: 'medium' },
      { label: '低', value: 'low' }
    ],
    loading: false,
    page: 1,
    pageSize: 20,
    hasMore: true
  },

  onLoad() {
    this.checkAuth();
    this.loadRiskUsers();
  },

  checkAuth() {
    const shopAdminInfo = wx.getStorageSync('shop_admin_info');
    if (!shopAdminInfo || !shopAdminInfo.token) {
      wx.redirectTo({
        url: '/shop-admin/login/login'
      });
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true, userList: [] });
    this.loadRiskUsers();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadRiskUsers();
    }
  },

  onRiskTypeFilter(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({
      currentRiskType: type,
      page: 1,
      hasMore: true,
      userList: []
    });
    this.loadRiskUsers();
  },

  onRiskLevelFilter(e) {
    const level = e.currentTarget.dataset.level;
    this.setData({
      currentRiskLevel: level,
      page: 1,
      hasMore: true,
      userList: []
    });
    this.loadRiskUsers();
  },

  loadRiskUsers() {
    if (this.data.loading) return;
    this.setData({ loading: true });

    const params = {
      page: this.data.page,
      page_size: this.data.pageSize
    };

    if (this.data.currentRiskType) {
      params.risk_type = this.data.currentRiskType;
    }
    if (this.data.currentRiskLevel) {
      params.risk_level = this.data.currentRiskLevel;
    }

    request.get('/shop-admin/risk-users', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        risk_level_text: item.risk_level_text || '',
        handle_status_text: item.handle_status_text || '',
        handle_status_color: item.handle_status_color || '',
        trigger_time: item.time_text || ''
      }));

      this.setData({
        userList: this.data.page === 1 ? list : [...this.data.userList, ...list],
        loading: false,
        hasMore: list.length >= this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      this.setData({ loading: false });
    });
  }
});