const request = require('../../utils/request');

Page({
  data: {
    role: '',
    agreementType: '',
    agreementContent: '',
    agreementVersion: '',
    agreementTitle: '',
    agreed: false,
    loading: true,
    forceSign: false
  },

  onLoad(options) {
    const { role, type, force } = options;
    this.setData({
      role: role || '',
      agreementType: type || 'user_service',
      forceSign: force === '1'
    });
    this.loadAgreement();
  },

  // 协议标题/内容/版本均由后端下发,前端不做类型→标题映射
  loadAgreement() {
    request.get('/compliance/agreement/latest', {
      role: this.data.role,
      agreement_type: this.data.agreementType
    }).then(res => {
      this.setData({
        agreementContent: res.content || '',
        agreementVersion: res.version || '',
        agreementTitle: res.title || '',
        loading: false
      });
    }).catch(err => {
      wx.showToast({ title: err.msg || '加载失败', icon: 'none' });
      this.setData({ loading: false });
    });
  },

  onAgreeChange(e) {
    this.setData({ agreed: e.detail.value });
  },

  onSign() {
    if (!this.data.agreed) {
      wx.showToast({ title: '请先阅读并同意协议', icon: 'none' });
      return;
    }

    wx.showLoading({ title: '签署中...', mask: true });
    request.post('/compliance/agreement/sign', {
      role: this.data.role,
      agreement_type: this.data.agreementType,
      version: this.data.agreementVersion
    }).then(() => {
      wx.hideLoading();
      
      const signedKey = 'agreement_' + this.data.role + '_' + this.data.agreementType + '_version';
      wx.setStorageSync(signedKey, this.data.agreementVersion);
      
      wx.showToast({ title: '签署成功', icon: 'success' });
      
      const pages = getCurrentPages();
      if (pages.length > 1) {
        wx.navigateBack();
      } else {
        wx.switchTab({
          url: '/pages/index/index'
        });
      }
    }).catch(err => {
      wx.hideLoading();
      wx.showToast({ title: err.msg || '签署失败', icon: 'none' });
    });
  },

  onBack() {
    if (this.data.forceSign) {
      wx.showModal({
        title: '提示',
        content: '您必须阅读并同意协议后才能继续使用',
        showCancel: false,
        confirmText: '我知道了'
      });
      return;
    }
    wx.navigateBack();
  }
});
