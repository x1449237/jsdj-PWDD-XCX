const request = require('../../../utils/request');
const auth = require('../../../utils/auth');
const app = getApp();

Page({
  data: {
    isLogin: false,
    tabIndex: 0,
    tabs: ['可使用', '已使用', '已过期'],
    couponList: [],
    loading: false,
    page: 1,
    pageSize: 20,
    total: 0,
    hasMore: true
  },

  onLoad() {
    this.checkLogin();
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

  onTabTap(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ tabIndex: index, page: 1, couponList: [], hasMore: true });
    this.loadCoupons();
  },

  loadCoupons() {
    if (!this.data.isLogin) return;
    if (this.data.loading || !this.data.hasMore) return;

    this.setData({ loading: true });
    const statusMap = ['unused', 'used', 'expired'];
    const status = statusMap[this.data.tabIndex];

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

  onUseCoupon(e) {
    const coupon = e.currentTarget.dataset.item;
    if (coupon.status !== 'unused') return;
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
