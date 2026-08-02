const extApi = require('../../utils/ext-api');
const app = getApp();

const toBool = (v) => v === true || v === 1 || v === '1';

Page({
  data: {
    loaded: false,
    form: {
      order_notify: false,
      after_sale_notify: false,
      marketing_notify: false
    }
  },

  onLoad() {
    this.loadSettings();
  },

  onShow() {
    if (!app.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/login' });
    }
  },

  loadSettings() {
    extApi.getNotificationSettings().then((res) => {
      const settings = res || {};
      this.setData({
        loaded: true,
        form: {
          order_notify: toBool(settings.order_notify),
          after_sale_notify: toBool(settings.after_sale_notify),
          marketing_notify: toBool(settings.marketing_notify)
        }
      });
    }).catch(() => {
      this.setData({ loaded: true });
    });
  },

  onSwitchChange(e) {
    const field = e.currentTarget.dataset.field;
    this.setData({ [`form.${field}`]: e.detail.value });
  },

  onSave() {
    const { order_notify, after_sale_notify, marketing_notify } = this.data.form;
    extApi.updateNotificationSettings({
      order_notify: order_notify ? 1 : 0,
      after_sale_notify: after_sale_notify ? 1 : 0,
      marketing_notify: marketing_notify ? 1 : 0
    }).then(() => {
      wx.showToast({ title: '已保存', icon: 'success' });
    }).catch(() => {});
  }
});
