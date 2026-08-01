const request = require('../../utils/request');
const util = require('../../utils/util');

Page({
  data: {
    gameList: [],
    currentGameId: '',
    packageList: [],
    loading: false
  },

  onLoad(options) {
    const gameId = options.gameId || '';
    this.setData({ currentGameId: gameId });
    this.loadGameList();
    this.loadPackageList();
  },

  loadGameList() {
    request.get('/config/service-types').then((res) => {
      this.setData({
        gameList: res.games || []
      });
    }).catch(() => {});
  },

  loadPackageList() {
    this.setData({ loading: true });
    const params = {};
    if (this.data.currentGameId) {
      params.game_id = this.data.currentGameId;
    }
    request.get('/orders/packages', params).then((res) => {
      const list = (res.list || []).map(item => ({
        ...item,
        price_text: util.fenToYuan(item.price),
        original_price_text: util.fenToYuan(item.original_price),
        discount: item.original_price > 0 
          ? Math.round((item.price / item.original_price) * 10) / 10 
          : 0
      }));
      this.setData({ packageList: list });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  onGameChange(e) {
    const gameId = e.currentTarget.dataset.id;
    this.setData({ currentGameId: gameId });
    this.loadPackageList();
  },

  onPackageSelect(e) {
    const pkg = e.currentTarget.dataset.pkg;
    const pages = getCurrentPages();
    if (pages.length > 1) {
      const prevPage = pages[pages.length - 2];
      if (prevPage && prevPage.selectPackage) {
        prevPage.selectPackage(pkg);
      }
    }
    wx.navigateBack();
  }
});
