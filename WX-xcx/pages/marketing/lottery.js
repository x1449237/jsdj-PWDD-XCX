const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    isLogin: false,
    activity: null,
    prizes: [],
    isDrawing: false,
    result: null,
    showResult: false,
    rotateDegree: 0,
    canDraw: 0,
    records: []
  },

  onLoad() {
    this.checkLogin();
    this.loadActivity();
  },

  onShow() {
    this.checkLogin();
  },

  checkLogin() {
    const isLogin = app.globalData.isLogin;
    this.setData({ isLogin });
  },

  loadActivity() {
    request.get('/lottery/activities').then((res) => {
      const activity = res.data?.activity;
      const prizes = res.data?.prizes || [];
      this.setData({
        activity,
        prizes
      });
      if (activity && this.data.isLogin) {
        this.loadRecords();
      }
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    });
  },

  loadRecords() {
    request.get('/lottery/records', { limit: 10 }).then((res) => {
      this.setData({
        records: res.data?.list || []
      });
    }).catch(() => {});
  },

  onDraw() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    if (this.isDrawing) return;
    if (!this.data.activity) {
      wx.showToast({ title: '活动未开始', icon: 'none' });
      return;
    }

    this.setData({ isDrawing: true });

    request.post('/lottery/draw', {
      activity_id: this.data.activity.id
    }).then((res) => {
      const result = res.data;
      const prizeIndex = this.data.prizes.findIndex(p => p.id === result.prize_id);
      const baseDegree = 360 * 5;
      const eachDegree = 360 / this.data.prizes.length;
      const targetDegree = baseDegree + (prizeIndex * eachDegree) + (eachDegree / 2);

      this.setData({
        rotateDegree: this.data.rotateDegree + targetDegree,
        result
      });

      setTimeout(() => {
        this.setData({
          isDrawing: false,
          showResult: true
        });
        this.loadRecords();
      }, 3000);
    }).catch((err) => {
      this.setData({ isDrawing: false });
      wx.showToast({
        title: err.msg || '抽奖失败',
        icon: 'none'
      });
    });
  },

  closeResult() {
    this.setData({ showResult: false, result: null });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
