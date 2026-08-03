const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    dashboard: null,
    loading: true,
    myRole: '',
    myRoleName: '',
    menus: []
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadDashboard();
  },

  onShow() {
    if (this.data.clubId > 0) {
      this.loadDashboard();
    }
  },

  // 菜单列表由后端按当前用户角色下发,前端不做角色→菜单过滤
  async loadDashboard() {
    try {
      const res = await request.get('/club/manage/dashboard', {
        club_id: this.data.clubId
      });
      const data = res.data || {};
      this.setData({
        dashboard: data,
        myRole: data.my_role || '',
        myRoleName: data.my_role_name || '',
        menus: data.menus || [],
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
  },

  goToMenu(e) {
    const path = e.currentTarget.dataset.path;
    if (path) {
      wx.navigateTo({ url: path });
    }
  },

  goPublish() {
    wx.navigateTo({ url: '/pages/club/dynamic/publish?id=' + this.data.clubId });
  }
});
