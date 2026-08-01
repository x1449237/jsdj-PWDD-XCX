const request = require('../../../utils/request');
const app = getApp();

Page({
  data: {
    childUserId: '',
    verifyCode: '',
    sending: false,
    countdown: 0,
    submitting: false,
    childPreview: {}
  },

  onLoad(options) {
    if (options.child_user_id) {
      this.setData({ childUserId: options.child_user_id });
      this.previewChildInfo(options.child_user_id);
    }
  },

  onChildUserIdInput(e) {
    const val = e.detail.value;
    this.setData({ childUserId: val });
    if (val && val.length >= 5) {
      clearTimeout(this._previewTimer);
      this._previewTimer = setTimeout(() => {
        this.previewChildInfo(val);
      }, 500);
    } else {
      this.setData({ childPreview: {} });
    }
  },

  onVerifyCodeInput(e) {
    this.setData({ verifyCode: e.detail.value });
  },

  onScanQrCode() {
    wx.scanCode({
      onlyFromCamera: false,
      scanType: ['qrCode'],
      success: (res) => {
        try {
          let result = res.result || '';
          let childUserId = '';
          if (result.startsWith('http')) {
            const match = result.match(/child_user_id=(\d+)/) || result.match(/uid[=:](\d+)/);
            childUserId = match ? match[1] : '';
          } else if (/^\d+$/.test(result)) {
            childUserId = result;
          } else {
            try {
              const parsed = JSON.parse(result);
              childUserId = parsed.child_user_id || parsed.uid || '';
            } catch (e) {
              childUserId = '';
            }
          }
          if (childUserId) {
            wx.showToast({ title: '扫码成功', icon: 'success' });
            this.setData({ childUserId });
            this.previewChildInfo(childUserId);
          } else {
            wx.showToast({ title: '未识别到用户ID', icon: 'none' });
          }
        } catch (err) {
          wx.showToast({ title: '二维码解析失败', icon: 'none' });
        }
      },
      fail: (err) => {
        if (err && err.errMsg && err.errMsg.indexOf('cancel') === -1) {
          wx.showToast({ title: '扫码失败，请重试', icon: 'none' });
        }
      }
    });
  },

  previewChildInfo(userId) {
    request.get('/guardian/child_preview', {
      child_user_id: userId
    }).then((res) => {
      this.setData({ childPreview: res.data || {} });
    }).catch(() => {
      this.setData({ childPreview: {} });
    });
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
      wx.showToast({ title: (err && err.message) || '发送失败', icon: 'none' });
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
