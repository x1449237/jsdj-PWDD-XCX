const request = require('../../utils/request');
const auth = require('../../utils/auth');
const app = getApp();

Page({
  data: {
    isLogin: false,
    userInfo: {},
    balance: '0.00',
    joinTapCount: 0,
    joinTapTimer: null,
    clubJoinOpen: false  // 俱乐部入驻开关状态
  },

  onLoad() {
    this.refreshUserInfo();
    this.checkClubSwitch();
  },

  onShow() {
    this.refreshUserInfo();
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 3
      });
    }
  },

  async checkClubSwitch() {
    try {
      const res = await request.get('/club/check_switch');
      this.setData({ clubJoinOpen: res.data?.club_join_open === true });
    } catch (e) {
      this.setData({ clubJoinOpen: false });
    }
  },

  refreshUserInfo() {
    const isLogin = app.globalData.isLogin;
    const userInfo = app.globalData.userInfo || auth.getStoredUserInfo() || {};

    this.setData({
      isLogin: isLogin,
      userInfo: userInfo
    });

    if (isLogin) {
      this.loadBalance();
    }
  },

  loadBalance() {
    request.get('/wallet/balance').then((res) => {
      this.setData({
        balance: (res.balance / 100).toFixed(2)
      });
    }).catch(() => {
      this.setData({ balance: '0.00' });
    });
  },

  onChooseAvatar(e) {
    const { avatarUrl } = e.detail;
    request.put('/user/profile', { avatar: avatarUrl }).then(() => {
      app.globalData.userInfo.avatar = avatarUrl;
      auth.setStoredUserInfo(app.globalData.userInfo);
      this.setData({
        'userInfo.avatar': avatarUrl
      });
    });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  },

  onRealNameAuth() {
    wx.navigateTo({
      url: '/package-settings/real-name-auth/real-name-auth'
    });
  },

  onBalance() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/package-wallet/balance/balance'
    });
  },

  onWithdraw() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/package-wallet/withdraw/withdraw'
    });
  },

  onMyOrders() {
    wx.switchTab({
      url: '/pages/my-orders/my-orders'
    });
  },

  onProfitShare() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/profit-share/list'
    });
  },

  onMyCoupons() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/marketing/coupon-list'
    });
  },

  onRecharge() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/marketing/recharge'
    });
  },

  onInvite() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/marketing/invite'
    });
  },

  onLottery() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/marketing/lottery'
    });
  },

  onGroupBuy() {
    wx.navigateTo({
      url: '/pages/marketing/group-buy'
    });
  },

  onMyReviews() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.showToast({
      title: '评价管理开发中',
      icon: 'none'
    });
  },

  onAppealCenter() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/appeal-list/appeal-list'
    });
  },

  onSettings() {
    wx.navigateTo({
      url: '/package-settings/settings/settings'
    });
  },

  onAbout() {
    wx.navigateTo({
      url: '/package-settings/about/about'
    });
  },

  onJoinUs() {
    this.data.joinTapCount++;

    if (this.data.joinTapTimer) {
      clearTimeout(this.data.joinTapTimer);
    }

    if (this.data.joinTapCount >= 3) {
      this.data.joinTapCount = 0;
      wx.navigateTo({
        url: '/package-player/player-apply/player-apply'
      });
    } else {
      this.data.joinTapTimer = setTimeout(() => {
        this.data.joinTapCount = 0;
      }, 2000);
    }
  },

  onClubJoin() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/club/join/join'
    });
  },

  onClubList() {
    wx.navigateTo({
      url: '/pages/club/list/list'
    });
  },

  onParentGuardian() {
    if (!this.data.isLogin) {
      this.onLogin();
      return;
    }
    wx.navigateTo({
      url: '/pages/parent-guardian/home/home'
    });
  },

  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout();
        }
      }
    });
  }
});