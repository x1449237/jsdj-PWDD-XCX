const request = require('../../utils/request');
const auth = require('../../utils/auth');
const app = getApp();

Page({
  data: {
    avatarUrl: '',
    nickname: '',
    inviteCode: '',
    role: 'user',
    phone: '',
    submitting: false,
    canSubmit: false
  },

  onLoad(options) {
    const userInfo = app.globalData.userInfo || {};
    const phone = userInfo.phone || '';
    this.setData({
      phone: phone,
      nickname: userInfo.nickname || '',
      avatarUrl: userInfo.avatar || ''
    });
    this.checkCanSubmit();
  },

  onChooseAvatar(e) {
    const { avatarUrl } = e.detail;
    this.setData({ avatarUrl });
    this.checkCanSubmit();
  },

  onNicknameInput(e) {
    this.setData({ nickname: e.detail.value });
    this.checkCanSubmit();
  },

  onInviteCodeInput(e) {
    this.setData({ inviteCode: e.detail.value });
  },

  onSelectRole(e) {
    const role = e.currentTarget.dataset.role;
    this.setData({ role });
  },

  checkCanSubmit() {
    const canSubmit = !!this.data.nickname.trim() && !!this.data.avatarUrl;
    this.setData({ canSubmit });
  },

  onSubmit() {
    if (!this.data.canSubmit) {
      wx.showToast({
        title: '请完善信息',
        icon: 'none'
      });
      return;
    }

    this.setData({ submitting: true });

    auth.register({
      nickname: this.data.nickname.trim(),
      avatar: this.data.avatarUrl,
      invite_code: this.data.inviteCode.trim(),
      role: this.data.role
    }).then((res) => {
      app.globalData.userInfo = res.user_info || {};
      auth.setStoredUserInfo(app.globalData.userInfo);

      wx.showToast({
        title: '注册成功',
        icon: 'success',
        duration: 1500
      });

      setTimeout(() => {
        wx.switchTab({
          url: '/pages/index/index'
        });
      }, 1500);
    }).catch((err) => {
      this.setData({ submitting: false });
    });
  },

  onSkip() {
    wx.switchTab({
      url: '/pages/index/index'
    });
  }
});