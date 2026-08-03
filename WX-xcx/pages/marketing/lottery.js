/**
 * lottery.js — 小程序前端零逻辑架构
 *
 * 架构铁律：抽奖算法、概率、中奖结果全部由 Go 后端 POST /lottery/draw 执行。
 * 前端禁止有任何抽奖逻辑、概率配置、兜底 mock 数据。
 * 前端只负责：调 API → 拿后端返回的 prize_index/is_win/prize_name → 渲染转盘动画。
 */
const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    isLogin: false,
    activity: null,       // 后端下发活动配置
    prizes: [],           // 后端下发奖品列表
    isDrawing: false,
    result: null,         // 后端返回的中奖结果
    showResult: false,
    rotateDegree: 0,
    records: [],
    orderId: '',
    source: ''
  },

  onLoad(options) {
    const { orderId = '', source = '' } = options || {};
    this.setData({ orderId, source });
    this.checkLogin();
    this.loadActivity();
  },

  onShow() {
    this.checkLogin();
  },

  checkLogin() {
    const isLogin = !!(app.globalData && app.globalData.isLogin);
    this.setData({ isLogin });
  },

  // 加载活动配置(后端权威返回,前端不兜底)
  loadActivity() {
    request.get('/lottery/activities').then((res) => {
      this.setData({
        activity: res.activity || null,
        prizes: res.prizes || []
      });
      if (this.data.isLogin) {
        this.loadRecords();
      }
    }).catch(() => {
      // 后端失败:不伪造任何数据,提示用户
      wx.showToast({ title: '活动加载失败,请重试', icon: 'none' });
    });
  },

  loadRecords() {
    request.get('/lottery/records', { limit: 10 }).then((res) => {
      this.setData({ records: res.list || [] });
    }).catch(() => {
      this.setData({ records: [] });
    });
  },

  // 抽奖:仅调后端 API,前端不做任何随机/概率计算
  onDraw() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    if (this.data.isDrawing) return;
    if (!this.data.activity || !this.data.prizes.length) {
      wx.showToast({ title: '活动未开始', icon: 'none' });
      return;
    }

    this.setData({ isDrawing: true });
    const { orderId, source } = this.data;

    request.post('/lottery/draw', {
      activity_id: this.data.activity.id,
      order_id: orderId,
      source
    }).then((res) => {
      // 后端返回: prize_index(转盘位置), is_win, prize_name, prize_id
      const prizeIndex = res.prize_index != null ? res.prize_index : 0;
      const baseDegree = 360 * 5;
      const eachDegree = 360 / this.data.prizes.length;
      const targetDegree = baseDegree + (prizeIndex * eachDegree) + (eachDegree / 2);

      this.setData({
        rotateDegree: this.data.rotateDegree + targetDegree,
        result: {
          is_win: !!res.is_win,
          prize_id: res.prize_id,
          prize_name: res.prize_name || ''
        }
      });

      setTimeout(() => {
        this.setData({ isDrawing: false, showResult: true });
        this.loadRecords();
      }, 3000);
    }).catch(() => {
      // 后端失败:不伪造结果,直接提示
      this.setData({ isDrawing: false });
      wx.showToast({ title: '抽奖失败,请重试', icon: 'none' });
    });
  },

  closeResult() {
    this.setData({ showResult: false, result: null });
  },

  onLogin() {
    wx.navigateTo({ url: '/pages/login/login' });
  }
});
