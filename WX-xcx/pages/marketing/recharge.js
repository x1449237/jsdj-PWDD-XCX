const request = require('../../../utils/request');
const auth = require('../../../utils/auth');
const app = getApp();

Page({
  data: {
    isLogin: false,
    balance: '0.00',
    activityList: [],
    selectedActivity: null,
    loading: false,
    paying: false
  },

  onLoad() {
    this.checkLogin();
  },

  onShow() {
    this.checkLogin();
    if (this.data.isLogin) {
      this.loadBalance();
      this.loadActivities();
    }
  },

  checkLogin() {
    const isLogin = app.globalData.isLogin;
    this.setData({ isLogin });
  },

  loadBalance() {
    // 余额文案由后端返回 balance_text，前端不做分转元换算
    request.get('/wallet/balance').then((res) => {
      this.setData({
        balance: res.balance_text || '0.00'
      });
    }).catch(() => {
      this.setData({ balance: '0.00' });
    });
  },

  loadActivities() {
    this.setData({ loading: true });
    request.get('/recharge/activities').then((res) => {
      const list = res.data?.list || [];
      this.setData({
        activityList: list,
        selectedActivity: list.length > 0 ? list[0] : null
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  onSelectActivity(e) {
    const activity = e.currentTarget.dataset.item;
    this.setData({ selectedActivity: activity });
  },

  onRecharge() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    if (!this.data.selectedActivity) {
      wx.showToast({ title: '请选择充值档位', icon: 'none' });
      return;
    }
    if (this.data.paying) return;

    this.setData({ paying: true });
    const activityId = this.data.selectedActivity.id;

    request.post('/recharge', {
      activity_id: activityId
    }).then((res) => {
      if (res.data?.pay_params) {
        const payParams = res.data.pay_params;
        wx.requestPayment({
          timeStamp: payParams.timeStamp,
          nonceStr: payParams.nonceStr,
          package: payParams.package,
          signType: payParams.signType,
          paySign: payParams.paySign,
          success: () => {
            wx.showToast({ title: '充值成功', icon: 'success' });
            this.loadBalance();
          },
          fail: () => {
            wx.showToast({ title: '支付取消', icon: 'none' });
          }
        });
      } else {
        wx.showToast({ title: '充值成功', icon: 'success' });
        this.loadBalance();
      }
    }).catch((err) => {
      wx.showToast({
        title: err.msg || '充值失败',
        icon: 'none'
      });
    }).finally(() => {
      this.setData({ paying: false });
    });
  },

  onRechargeRecords() {
    wx.navigateTo({
      url: '/pages/marketing/recharge-records'
    });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
