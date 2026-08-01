const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    dashboard: null,
    loading: true,
    myRole: '',
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

  async loadDashboard() {
    try {
      const res = await request.get('/club/manage/dashboard', {
        club_id: this.data.clubId
      });
      const data = res.data;
      const role = data.my_role || '';

      const allMenus = [
        { id: 'members', name: '成员管理', icon: 'team', path: '/pages/club/manage/members?id=' + this.data.clubId, roles: ['founder', 'manager'] },
        { id: 'internal-order', name: '内部订单', icon: 'order', path: '/pages/club/manage/internal-order?id=' + this.data.clubId, roles: ['founder', 'manager', 'member'] },
        { id: 'coupon', name: '优惠券管理', icon: 'coupon', path: '/pages/club/manage/coupon?id=' + this.data.clubId, roles: ['founder', 'manager'] },
        { id: 'dashboard', name: '数据看板', icon: 'chart', path: '/pages/club/manage/dashboard?id=' + this.data.clubId, roles: ['founder', 'manager'] },
        { id: 'dynamic', name: '动态墙', icon: 'image', path: '/pages/club/dynamic/list?id=' + this.data.clubId, roles: ['founder', 'manager', 'member'] },
      ];

      const menus = allMenus.filter(m => m.roles.includes(role));

      this.setData({
        dashboard: data,
        myRole: role,
        myRoleName: data.my_role_name || '',
        menus,
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
