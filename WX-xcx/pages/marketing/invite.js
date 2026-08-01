const request = require('../../../utils/request');
const auth = require('../../../utils/auth');
const app = getApp();

Page({
  data: {
    isLogin: false,
    userInfo: {},
    inviteCode: '',
    inviteQrCode: '',
    stats: {
      invite_count: 0,
      reward_amount: '0.00'
    },
    rewardList: [],
    showShare: false
  },

  onLoad() {
    this.checkLogin();
  },

  onShow() {
    this.checkLogin();
    if (this.data.isLogin) {
      this.loadInviteInfo();
      this.loadInviteStats();
    }
  },

  checkLogin() {
    const isLogin = app.globalData.isLogin;
    const userInfo = app.globalData.userInfo || auth.getStoredUserInfo() || {};
    this.setData({ isLogin, userInfo });
  },

  loadInviteInfo() {
    request.get('/invite/info').then((res) => {
      this.setData({
        inviteCode: res.data?.invite_code || '',
        inviteQrCode: res.data?.qr_code || ''
      });
    }).catch(() => {});
  },

  loadInviteStats() {
    request.get('/invite/stats').then((res) => {
      this.setData({
        stats: res.data || {
          invite_count: 0,
          reward_amount: '0.00'
        }
      });
    }).catch(() => {});

    request.get('/invite/rewards', { limit: 10 }).then((res) => {
      this.setData({
        rewardList: res.data?.list || []
      });
    }).catch(() => {});
  },

  onCopyCode() {
    if (!this.data.inviteCode) return;
    wx.setClipboardData({
      data: this.data.inviteCode,
      success: () => {
        wx.showToast({ title: '邀请码已复制', icon: 'success' });
      }
    });
  },

  onShareAppMessage() {
    const inviteCode = this.data.inviteCode;
    return {
      title: '快来一起开黑吧！输入我的邀请码有惊喜',
      path: `/pages/register/register?invite_code=${inviteCode}`,
      imageUrl: '/assets/images/share-cover.jpg'
    };
  },

  onShareTimeline() {
    const inviteCode = this.data.inviteCode;
    return {
      title: '快来一起开黑吧！',
      query: `invite_code=${inviteCode}`
    };
  },

  onSharePoster() {
    wx.showToast({
      title: '生成分享海报开发中',
      icon: 'none'
    });
  },

  onRewardRule() {
    wx.showModal({
      title: '邀请奖励规则',
      content: '1. 邀请好友注册并完成首单，您将获得奖励\n2. 奖励金额以实际到账为准\n3. 活动最终解释权归平台所有',
      showCancel: false,
      confirmText: '我知道了'
    });
  },

  onLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    });
  }
});
