// 架构规则：小程序前端不得包含任何业务逻辑。
// 前端只负责：1) 调用后端 API（request.get/post/put/del）
// 2) 渲染后端返回的字段（status_text、amount_text、*_masked、*_color、time_text 等）
// 3) 提供纯 UI 反馈（toast、loading、非空提示、进度条）
// 后端负责：状态/类型文本映射、金额转换、时间格式化、脱敏、业务校验、权限控制、折扣计算。
const request = require('../../utils/request');

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
      // 直接使用后端返回的 price_text / original_price_text / discount_label，前端不做分转元与折扣计算
      const list = (res.list || []).map(item => ({
        ...item,
        price_text: item.price_text || '',
        original_price_text: item.original_price_text || '',
        discount: item.discount_label || 0
      }));
      this.setData({ packageList: list });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ packageList: [] });
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
