const request = require('../../utils/request');
const auth = require('../../utils/auth');
const validator = require('../../utils/validator');

Page({
  data: {
    account: '',
    password: '',
    shopCode: '',
    showPassword: true,
    loading: false,
    isFirstLogin: true
  },

  onLoad(options) {
    const shopAdminInfo = wx.getStorageSync('shop_admin_info');
    if (shopAdminInfo && shopAdminInfo.token) {
      wx.redirectTo({
        url: '/shop-admin/orders/orders'
      });
      return;
    }

    if (options && options.from === 'expired') {
      wx.showToast({
        title: '登录已过期，请重新登录',
        icon: 'none',
        duration: 2000
      });
    }
  },

  onAccountInput(e) {
    this.setData({ account: e.detail.value });
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value });
  },

  onShopCodeInput(e) {
    this.setData({ shopCode: e.detail.value });
  },

  onTogglePassword() {
    this.setData({ showPassword: !this.data.showPassword });
  },

  onLogin() {
    const { account, password, shopCode } = this.data;

    if (!account.trim()) {
      wx.showToast({ title: '请输入管理员账号', icon: 'none' });
      return;
    }

    if (!password.trim()) {
      wx.showToast({ title: '请输入管理员密码', icon: 'none' });
      return;
    }

    const pwdResult = validator.validatePassword(password);
    if (!pwdResult.valid) {
      wx.showToast({ title: pwdResult.message, icon: 'none' });
      return;
    }

    if (!shopCode.trim()) {
      wx.showToast({ title: '请输入店铺安全码', icon: 'none' });
      return;
    }

    this.setData({ loading: true });

    request.post('/shop-admin/login', {
      account: account.trim(),
      password: password.trim(),
      shop_code: shopCode.trim()
    }).then((res) => {
      wx.setStorageSync('shop_admin_info', {
        token: res.token,
        shop_id: res.shop_id,
        shop_name: res.shop_name,
        account: account.trim()
      });

      wx.showToast({
        title: '登录成功',
        icon: 'success',
        duration: 1500
      });

      setTimeout(() => {
        wx.redirectTo({
          url: '/shop-admin/orders/orders'
        });
      }, 1500);
    }).catch((err) => {
      console.error('管理端登录失败:', err);
      this.setData({ loading: false });
    });
  }
});