const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    realName: '',
    idCard: '',
    livenessStatus: 'idle', // idle | scanning | success | fail
    livenessLoading: false,
    livenessFailReason: '',
    isMinor: false,
    overLimit: false,
    submitting: false
  },

  onLoad(options) {
    // 可传入是否超额消费标记
    if (options && options.overLimit) {
      this.setData({ overLimit: true });
    }
  },

  onNameInput(e) {
    this.setData({ realName: e.detail.value });
  },

  onIdCardInput(e) {
    this.setData({ idCard: e.detail.value });
  },

  onStartLiveness() {
    const { realName, idCard } = this.data;
    if (!realName || !idCard) {
      wx.showToast({ title: '请先填写姓名和身份证号', icon: 'none' });
      return;
    }

    this.setData({
      livenessStatus: 'scanning',
      livenessLoading: true,
      livenessFailReason: ''
    });

    // 调用活体检测接口
    request.post('/user/realname/face-verify', {
      real_name: realName,
      id_card: idCard
    }).then((res) => {
      this.setData({
        livenessStatus: 'success',
        livenessLoading: false
      });
    }).catch((err) => {
      this.setData({
        livenessStatus: 'fail',
        livenessLoading: false,
        livenessFailReason: err.message || '活体检测未通过，请重试'
      });
    });
  },

  onSubmit() {
    const { realName, idCard, livenessStatus } = this.data;

    if (livenessStatus !== 'success') {
      wx.showToast({ title: '请先完成活体检测', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });

    request.post('/user/realname/submit', {
      real_name: realName,
      id_card: idCard
    }).then((res) => {
      const isMinor = res.is_minor || false;
      const overLimit = res.over_limit || false;

      this.setData({ submitting: false });

      if (isMinor) {
        this.setData({ isMinor: true, overLimit: overLimit });
        wx.showToast({ title: '认证成功，您是未成年人', icon: 'none' });
      } else {
        wx.showToast({ title: '实名认证成功', icon: 'success' });
        setTimeout(() => {
          wx.navigateBack();
        }, 1500);
      }
    }).catch((err) => {
      this.setData({ submitting: false });
      wx.showToast({ title: err.message || '认证失败，请重试', icon: 'none' });
    });
  },

  onGoGuardian() {
    wx.navigateTo({
      url: '/pages/guardian/guardian'
    });
  }
});