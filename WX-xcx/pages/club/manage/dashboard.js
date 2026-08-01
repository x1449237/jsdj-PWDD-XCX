const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    dashboard: null,
    trendData: [],
    days: 7,
    loading: true
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadData();
  },

  onPullDownRefresh() {
    this.loadData();
  },

  async loadData() {
    this.setData({ loading: true });
    try {
      const [dashRes, trendRes] = await Promise.all([
        request.get('/club/manage/dashboard', { club_id: this.data.clubId }),
        request.get('/club/manage/trend', { club_id: this.data.clubId, days: this.data.days })
      ]);

      this.setData({
        dashboard: dashRes.data,
        trendData: trendRes.data.daily_stats || [],
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }
    wx.stopPullDownRefresh();
  },

  onDaysChange(e) {
    const days = parseInt(e.currentTarget.dataset.days);
    this.setData({ days });
    this.loadTrend();
  },

  async loadTrend() {
    try {
      const res = await request.get('/club/manage/trend', {
        club_id: this.data.clubId,
        days: this.data.days
      });
      this.setData({ trendData: res.data.daily_stats || [] });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }
  }
});
