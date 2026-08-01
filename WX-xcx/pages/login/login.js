const auth = require('../../utils/auth');
const request = require('../../utils/request');
const app = getApp();

Page({
  data: {
    agreed: false,
    loading: false,
    showAgreementModal: false,
    isFirstLogin: true
  },

  onLoad(options) {
    if (auth.isLogin()) {
      wx.switchTab({
        url: '/pages/index/index'
      });
      return;
    }

    if (!auth.isAgreementAccepted()) {
      this.setData({ showAgreementModal: true });
    }
  },

  onToggleAgreement() {
    this.setData({ agreed: !this.data.agreed });
  },

  onShowAgreement(e) {
    const type = e.currentTarget.dataset.type;
    if (type === 'service') {
      wx.showModal({
        title: '用户服务协议',
        content: '用户服务协议详情...',
        showCancel: false
      });
    } else {
      wx.showModal({
        title: '隐私政策',
        content: '隐私政策详情...',
        showCancel: false
      });
    }
  },

  onGetPhoneNumber(e) {
    if (!this.data.agreed) {
      wx.showToast({
        title: '请先同意服务协议',
        icon: 'none'
      });
      return;
    }

    auth.getPhoneNumber(e).then((detail) => {
      return this.doLogin(detail);
    }).catch((err) => {
      console.error('获取手机号失败:', err);
      wx.showToast({
        title: '获取手机号失败，请重试',
        icon: 'none'
      });
    });
  },

  doLogin(phoneDetail) {
    this.setData({ loading: true });

    auth.loginWithWx(phoneDetail).then((res) => {
      let userInfo = {};
      if (res.user_info) {
        userInfo = res.user_info;
      }

      app.setLoginState(res.token, userInfo);

      if (res.is_new_user) {
        wx.redirectTo({
          url: '/pages/register/register'
        });
      } else {
        wx.switchTab({
          url: '/pages/index/index'
        });
      }
    }).catch((err) => {
      console.error('登录失败:', err);
      this.setData({ loading: false });
    });
  },

  onVisitorLogin() {
    wx.switchTab({
      url: '/pages/index/index'
    });
  },

  onCloseAgreementModal() {
    // 首次登录不允许关闭
  },

  onAcceptAgreement() {
    auth.acceptAgreement();
    this.setData({
      showAgreementModal: false,
      agreed: true
    });
  },

  onRejectAgreement() {
    wx.showModal({
      title: '提示',
      content: '需要同意协议才能使用服务',
      showCancel: false,
      success: () => {
        wx.exitMiniProgram();
      }
    });
  }
});