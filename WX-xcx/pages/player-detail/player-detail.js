/**
 * 架构规则：小程序前端不包含任何业务逻辑。
 * 前端仅：1) 通过 request.get/post/put/del 调用后端接口
 *         2) 渲染后端返回字段（relative_time_text 等）
 *         3) 提供纯 UI 反馈（toast/loading/非空提示/进度条）
 * 禁止：mock 数据、状态/类型文本映射、金额换算、时间格式化、脱敏、业务校验、随机数。
 * 约定：request.get(url, data) 直接 resolve 出内层 data.data，res 即数据对象本身。
 */
const request = require('../../utils/request');

Page({
  data: {
    playerId: '',
    player: {},
    reviews: [],
    selectedServiceIndex: -1,
    isFavorited: false,
    playerTags: {
      game: [],
      position: [],
      voice: [],
      rank: [],
      skill: []
    }
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ playerId: options.id });
      this.loadPlayerDetail();
      this.loadReviews();
      this.loadPlayerTags();
      this.checkFavoriteStatus();
    }
  },

  onPullDownRefresh() {
    this.loadPlayerDetail();
    this.loadReviews();
    this.loadPlayerTags();
    this.checkFavoriteStatus();
    wx.stopPullDownRefresh();
  },

  loadPlayerTags() {
    request.get(`/player/tags`, {
      player_id: this.data.playerId
    }).then((res) => {
      this.setData({ playerTags: res || {} });
    }).catch(() => {});
  },

  checkFavoriteStatus() {
    request.get('/player/favorite/list').then((res) => {
      const list = res.list || [];
      const isFavorited = list.some(item => item.player_user_id == this.data.playerId);
      this.setData({ isFavorited });
    }).catch(() => {});
  },

  onToggleFavorite() {
    const { isFavorited, playerId } = this.data;

    if (isFavorited) {
      request.post('/player/favorite/cancel', {
        player_user_id: playerId
      }).then(() => {
        this.setData({ isFavorited: false });
        wx.showToast({ title: '已取消收藏', icon: 'none' });
      }).catch(() => {
        wx.showToast({ title: '操作失败', icon: 'none' });
      });
    } else {
      request.post('/player/favorite/add', {
        player_user_id: playerId
      }).then(() => {
        this.setData({ isFavorited: true });
        wx.showToast({ title: '收藏成功', icon: 'success' });
      }).catch(() => {
        wx.showToast({ title: '操作失败', icon: 'none' });
      });
    }
  },

  loadPlayerDetail() {
    request.get(`/players/${this.data.playerId}`).then((res) => {
      this.setData({
        player: res
      });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ player: {} });
    });
  },

  loadReviews() {
    request.get(`/players/${this.data.playerId}/reviews`, {
      page: 1,
      page_size: 5
    }).then((res) => {
      this.setData({ reviews: res.list || [] });
    }).catch(() => {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ reviews: [] });
    });
  },

  onSelectService(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ selectedServiceIndex: index });
  },

  onOrder() {
    const player = this.data.player;
    if (!player.id) return;

    const app = getApp();
    if (!app.globalData.isLogin) {
      wx.navigateTo({
        url: '/pages/login/login'
      });
      return;
    }

    let serviceId = '';
    if (this.data.selectedServiceIndex >= 0) {
      const service = player.services[this.data.selectedServiceIndex];
      if (service) {
        serviceId = service.id;
      }
    }

    wx.navigateTo({
      url: `/package-order/order-create/order-create?player_id=${player.id}&service_id=${serviceId}`
    });
  },

  onMoreReview() {
    wx.navigateTo({
      url: `/package-player/player-list/player-list`
    });
  }
});
