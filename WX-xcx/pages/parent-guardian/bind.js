const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    childUserId: '',
    verifyCode: '',
    sending: false,
    countdown: 0,
    submitting: false
  },

  onLoad(options) {
    if (options.child_user_id) {
      this.setData({ childUserId: options.child_user_id });
    }
  },

  onChildUserIdInput(e) {
    this.setData({ childUserId: e.detail.value });
  },

  onVerifyCodeInput(e) {
    this.setData({ verifyCode: e.detail.value });
  },

  onSendCode() {
    const { childUserId } = this.data;
    if (!childUserId) {
      wx.showToast({ title: '请先输入孩子用户ID', icon: 'none' });
      return;
    }

    this.setData({ sending: true });

    request.post('/guardian/send_bind_code', {
      child_user_id: childUserId
    }).then(() => {
      wx.showToast({ title: '验证码已发送', icon: 'success' });
      this.startCountdown();
    }).catch((err) => {
      wx.showToast({ title: err.message || '发送失败', icon: 'none' });
    }).finally(() => {
      this.setData({ sending: false });
    });
  },

  startCountdown() {
    let countdown = 60;
    this.setData({ countdown });
    const timer = setInterval(() => {
      countdown--;
      this.setData({ countdown });
      if (countdown <= 0) {
        clearInterval(timer);
      }
    }, 1000);
  },

  onBind() {
    const { childUserId, verifyCode } = this.data;
    if (!childUserId) {
      wx.showToast({ title: '请输入孩子用户ID', icon: 'none' });
      return;
    }
    if (!verifyCode) {
      wx.showToast({ title: '请输入验证码', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    request.post('/guardian/bind', {
      child_user_id: childUserId,
      verify_code: verifyCode
    }).then(() => {
      wx.showToast({ title: '绑定成功', icon: 'success' });
      setTimeout(() => {
        wx.redirectTo({ url: '/pages/parent-guardian/home/home' });
      }, 1500);
    }).catch((err) => {
      wx.showToast({ title: err.message || '绑定失败', icon: 'none' });
    }).finally(() => {
      this.setData({ submitting: false });
    });
  }
});
