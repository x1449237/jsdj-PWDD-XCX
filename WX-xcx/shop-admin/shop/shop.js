const request = require('../../utils/request');

Page({
  data: {
    shopInfo: {
      logo: '',
      banner: '',
      shopName: '',
      description: '',
      themeColor: '#e94560',
      notice: ''
    },
    originalShopInfo: {},
    colorOptions: [
      { value: '#e94560' },
      { value: '#0f3460' },
      { value: '#07c160' },
      { value: '#ff976a' },
      { value: '#ffd700' },
      { value: '#1a1a2e' },
      { value: '#6c5ce7' },
      { value: '#00b894' },
      { value: '#0984e3' },
      { value: '#d63031' }
    ],
    saving: false
  },

  onLoad() {
    this.checkAuth();
    this.loadShopInfo();
  },

  checkAuth() {
    const shopAdminInfo = wx.getStorageSync('shop_admin_info');
    if (!shopAdminInfo || !shopAdminInfo.token) {
      wx.redirectTo({
        url: '/shop-admin/login/login'
      });
    }
  },

  loadShopInfo() {
    request.get('/shop-admin/shop-info').then((res) => {
      const shopInfo = {
        logo: res.logo || '',
        banner: res.banner || '',
        shopName: res.shop_name || '',
        description: res.description || '',
        themeColor: res.theme_color || '#e94560',
        notice: res.notice || ''
      };

      this.setData({
        shopInfo: shopInfo,
        originalShopInfo: { ...shopInfo }
      });
    }).catch((err) => {
      console.error('加载店铺信息失败:', err);
    });
  },

  onChooseLogo() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.uploadImage(tempFilePath, 'logo');
      }
    });
  },

  onChooseBanner() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.uploadImage(tempFilePath, 'banner');
      }
    });
  },

  uploadImage(filePath, type) {
    wx.showLoading({ title: '上传中...' });

    wx.uploadFile({
      url: request.baseURL + '/upload/image',
      filePath: filePath,
      name: 'file',
      header: {
        'Authorization': 'Bearer ' + wx.getStorageSync('shop_admin_info').token
      },
      success: (res) => {
        wx.hideLoading();
        try {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            const key = type === 'logo' ? 'logo' : 'banner';
            this.setData({
              ['shopInfo.' + key]: data.data.url
            });
            wx.showToast({ title: '上传成功', icon: 'success' });
          } else {
            wx.showToast({ title: data.msg || '上传失败', icon: 'none' });
          }
        } catch (e) {
          wx.showToast({ title: '上传失败', icon: 'none' });
        }
      },
      fail: () => {
        wx.hideLoading();
        wx.showToast({ title: '网络异常', icon: 'none' });
      }
    });
  },

  onShopNameInput(e) {
    this.setData({ ['shopInfo.shopName']: e.detail.value });
  },

  onDescriptionInput(e) {
    this.setData({ ['shopInfo.description']: e.detail.value });
  },

  onNoticeInput(e) {
    this.setData({ ['shopInfo.notice']: e.detail.value });
  },

  onSelectColor(e) {
    const color = e.currentTarget.dataset.color;
    this.setData({ ['shopInfo.themeColor']: color });
  },

  onSave() {
    const { shopInfo } = this.data;

    if (!shopInfo.shopName.trim()) {
      wx.showToast({ title: '请输入店铺名称', icon: 'none' });
      return;
    }

    this.setData({ saving: true });

    request.post('/shop-admin/shop-info', {
      shop_name: shopInfo.shopName.trim(),
      logo: shopInfo.logo,
      banner: shopInfo.banner,
      description: shopInfo.description.trim(),
      theme_color: shopInfo.themeColor,
      notice: shopInfo.notice.trim()
    }).then(() => {
      wx.showToast({
        title: '保存成功',
        icon: 'success',
        duration: 2000
      });

      this.setData({
        saving: false,
        originalShopInfo: { ...shopInfo }
      });

      const shopAdminInfo = wx.getStorageSync('shop_admin_info');
      if (shopAdminInfo) {
        shopAdminInfo.shop_name = shopInfo.shopName;
        wx.setStorageSync('shop_admin_info', shopAdminInfo);
      }
    }).catch((err) => {
      console.error('保存失败:', err);
      this.setData({ saving: false });
    });
  }
});