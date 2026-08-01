const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    club: null,
    loading: true,
    activeTab: 'announcement',
    tabs: [
      { key: 'announcement', label: '公告' },
      { key: 'coupon', label: '优惠券' },
      { key: 'dynamic', label: '动态墙' },
      { key: 'branch', label: '分店' }
    ]
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadDetail();
  },

  async loadDetail() {
    try {
      const res = await request.get('/club/detail', {
        id: this.data.clubId
      });
      this.setData({ club: res.data, loading: false });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key;
    this.setData({ activeTab: key });
  },

  goManage() {
    if (this.data.club && this.data.club.my_role) {
      wx.navigateTo({ url: '/pages/club/manage/index?id=' + this.data.clubId });
    }
  },

  goJoinApply() {
    wx.showModal({
      title: '加入俱乐部',
      content: '确定申请加入该俱乐部吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await request.post('/club/member/join-apply', { club_id: this.data.clubId });
            wx.showToast({ title: '申请已提交', icon: 'success' });
          } catch (e) {
            wx.showToast({ title: e.message || '申请失败', icon: 'none' });
          }
        }
      }
    });
  },

  goDynamicDetail(e) {
    // 可以跳转到动态详情，这里先保留
  },

  previewImage(e) {
    const urls = e.currentTarget.dataset.urls;
    const current = e.currentTarget.dataset.current;
    if (urls && urls.length > 0) {
      wx.previewImage({ urls, current });
    }
  },

  receiveCoupon(e) {
    const id = e.currentTarget.dataset.id;
    wx.showModal({
      title: '领取优惠券',
      content: '确定领取这张优惠券吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await request.post('/club/coupon/receive', { id });
            wx.showToast({ title: '领取成功', icon: 'success' });
            this.loadDetail();
          } catch (e) {
            wx.showToast({ title: e.message || '领取失败', icon: 'none' });
          }
        }
      }
    });
  }
});
