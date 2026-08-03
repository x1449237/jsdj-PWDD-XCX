const request = require('../../../utils/request');
const auth = require('../../../utils/auth');
const app = getApp();

Page({
  data: {
    isLogin: false,
    tabIndex: 0,
    // tab 选项及对应 status 由后端下发 [{label, status}],前端不硬编码
    tabs: [],
    couponList: [],
    loading: false,
    page: 1,
    pageSize: 20,
    total: 0,
    hasMore: true
  },

  onLoad() {
    this.checkLogin();
    this.loadTabs();
  },

  onShow() {
    this.checkLogin();
    if (this.data.isLogin) {
      this.setData({ page: 1, couponList: [], hasMore: true });
      this.loadCoupons();
    }
  },

  checkLogin() {
    const isLogin = app.globalData.isLogin;
    this.setData({ isLogin });
  },

  // 加载 tab 配置(后端下发 label + status)
  loadTabs() {
    request.get('/coupon/tabs').then((res) => {
      this.setData({ tabs: (res && res.list) || [] });
    }).catch(() => {});
  },

  onTabTap(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ tabIndex: index, page: 1, couponList: [], hasMore: true });
    this.loadCoupons();
  },

  loadCoupons() {
    if (!this.data.isLogin) return;
    if (this.data.loading || !this.data.hasMore) return;
    if (this.data.tabs.length === 0) return;

    this.setData({ loading: true });
    // status 直接取后端下发的 tab.status,前端不做 index→status 映射
    const tab = this.data.tabs[this.data.tabIndex] || {};
    const status = tab.status || '';

    request.get('/coupon/my', {
      status,
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = res.data?.list || [];
      const newList = this.data.page === 1 ? list : this.data.couponList.concat(list);
      this.setData({
        couponList: newList,
        total: res.data?.total || 0,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  onReachBottom() {
    if (this.data.hasMore) {
      this.loadCoupons();
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, couponList: [], hasMore: true });
    this.loadCoupons();
    wx.stopPullDownRefresh();
  },

  // 是否可使用由后端 can_use 字段决定,前端不比较 status 字符串
  onUseCoupon(e) {
    const coupon = e.currentTarget.dataset.item;
    if (!coupon || coupon.can_use !== true) return;
    wx.switchTab({
      url: '/pages/index/index'
    });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
