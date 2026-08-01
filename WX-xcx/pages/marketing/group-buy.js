const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    isLogin: false,
    tabIndex: 0,
    tabs: ['热门拼团', '我的拼团'],
    activityList: [],
    myGroupList: [],
    loading: false,
    page: 1,
    pageSize: 20,
    hasMore: true
  },

  onLoad() {
    this.checkLogin();
    this.loadActivities();
  },

  onShow() {
    this.checkLogin();
    if (this.data.isLogin && this.data.tabIndex === 1) {
      this.loadMyGroups();
    }
  },

  checkLogin() {
    const isLogin = app.globalData.isLogin;
    this.setData({ isLogin });
  },

  onTabTap(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ tabIndex: index, page: 1, hasMore: true });
    if (index === 0) {
      this.setData({ activityList: [] });
      this.loadActivities();
    } else {
      if (!this.data.isLogin) {
        this.onLogin();
        return;
      }
      this.setData({ myGroupList: [] });
      this.loadMyGroups();
    }
  },

  loadActivities() {
    if (this.data.loading || !this.data.hasMore) return;
    this.setData({ loading: true });

    request.get('/group-buy/activities', {
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = res.data?.list || [];
      const newList = this.data.page === 1 ? list : this.data.activityList.concat(list);
      this.setData({
        activityList: newList,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  loadMyGroups() {
    if (!this.data.isLogin) return;
    if (this.data.loading || !this.data.hasMore) return;
    this.setData({ loading: true });

    request.get('/group-buy/my', {
      page: this.data.page,
      limit: this.data.pageSize
    }).then((res) => {
      const list = res.data?.list || [];
      const newList = this.data.page === 1 ? list : this.data.myGroupList.concat(list);
      this.setData({
        myGroupList: newList,
        hasMore: list.length === this.data.pageSize,
        page: this.data.page + 1
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  onJoinGroup(e) {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    const groupId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/marketing/group-buy-detail?id=${groupId}`
    });
  },

  onOpenGroup(e) {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    const activity = e.currentTarget.dataset.item;
    wx.showModal({
      title: '确认开团',
      content: `确定要开团「${activity.name}」吗？拼团价¥${activity.group_price}`,
      success: (res) => {
        if (res.confirm) {
          this.createGroup(activity.id);
        }
      }
    });
  },

  createGroup(activityId) {
    wx.showLoading({ title: '创建中...' });
    request.post('/group-buy/join', {
      activity_id: activityId
    }).then((res) => {
      wx.hideLoading();
      wx.showToast({ title: '开团成功', icon: 'success' });
      setTimeout(() => {
        wx.navigateTo({
          url: `/pages/marketing/group-buy-detail?id=${res.data.id}`
        });
      }, 1000);
    }).catch((err) => {
      wx.hideLoading();
      wx.showToast({
        title: err.msg || '开团失败',
        icon: 'none'
      });
    });
  },

  onReachBottom() {
    if (this.data.hasMore) {
      if (this.data.tabIndex === 0) {
        this.loadActivities();
      } else {
        this.loadMyGroups();
      }
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true });
    if (this.data.tabIndex === 0) {
      this.setData({ activityList: [] });
      this.loadActivities();
    } else {
      this.setData({ myGroupList: [] });
      this.loadMyGroups();
    }
    wx.stopPullDownRefresh();
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
