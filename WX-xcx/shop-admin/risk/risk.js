const request = require('../../utils/request');
const util = require('../../utils/util');

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

    const levelTextMap = {
      high: '高',
      medium: '中',
      low: '低'
    };

    const handleStatusMap = {
      0: '未处理',
      1: '处理中',
      2: '已处理'
    };

    const handleStatusColorMap = {
      0: '#e94560',
      1: '#ff976a',
      2: '#07c160'
    };

    request.get('/shop-admin/risk-users', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        risk_level_text: levelTextMap[item.risk_level] || '未知',
        handle_status_text: handleStatusMap[item.handle_status] || '未知',
        handle_status_color: handleStatusColorMap[item.handle_status] || '#999999',
        trigger_time: util.formatTime(item.trigger_time, 'YYYY-MM-DD HH:mm')
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