/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
const request = require('../../utils/request');

Page({
  data: {
    banners: [],
    categories: [],
    players: [],
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false,
    keyword: '',
    subscribeTmplIds: 'TEMPLATE_ID_PLACEHOLDER_01,TEMPLATE_ID_PLACEHOLDER_03,TEMPLATE_ID_PLACEHOLDER_04'
  },

  onLoad() {
    this.loadBanners();
    this.loadCategories();
    this.loadPlayers();
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 0
      });
    }
  },

  onPullDownRefresh() {
    this.setData({
      page: 1,
      players: [],
      noMore: false
    });
    this.loadBanners();
    this.loadPlayers();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadPlayers();
    }
  },

  loadBanners() {
    request.get('/banners').then((res) => {
      this.setData({ banners: res.list || [] });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ banners: [] });
    });
  },

  loadCategories() {
    request.get('/categories').then((res) => {
      this.setData({ categories: res.list || [] });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ categories: [] });
    });
  },

  loadPlayers() {
    if (this.data.loading || this.data.noMore) return;

    this.setData({ loading: true });

    request.get('/players', {
      page: this.data.page,
      page_size: this.data.pageSize,
      keyword: this.data.keyword
    }).then((res) => {
      const list = res.list || [];
      const players = this.data.players.concat(list);

      this.setData({
        players: players,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      if (this.data.page === 1) {
        this.setData({ players: [], loading: false, noMore: true });
      } else {
        this.setData({ loading: false });
      }
    });
  },

  onSearchTap() {
    wx.navigateTo({
      url: '/package-game/game-list/game-list'
    });
  },

  onBannerTap(e) {
    const id = e.currentTarget.dataset.id;
    console.log('Banner tapped:', id);
  },

  onCategoryTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/package-game/game-category/game-category?id=${id}`
    });
  },

  onPlayerTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/player-detail/player-detail?id=${id}`
    });
  },

  onMoreCategory() {
    wx.navigateTo({
      url: '/package-game/game-list/game-list'
    });
  },

  onMorePlayer() {
    wx.navigateTo({
      url: '/package-player/player-list/player-list'
    });
  },

  onSubscribeResult(e) {
    console.log('订阅消息授权结果:', e.detail);
  }
});
