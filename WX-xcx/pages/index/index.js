const request = require('../../utils/request');
const util = require('../../utils/util');

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
      this.setData({
        banners: [
          { id: 1, image: '/assets/images/banner1.png' },
          { id: 2, image: '/assets/images/banner2.png' },
          { id: 3, image: '/assets/images/banner3.png' }
        ]
      });
    });
  },

  loadCategories() {
    request.get('/categories').then((res) => {
      this.setData({ categories: res.list || [] });
    }).catch(() => {
      this.setData({
        categories: [
          { id: 1, name: '王者荣耀', icon: '/assets/icons/game-wzry.png' },
          { id: 2, name: '英雄联盟', icon: '/assets/icons/game-lol.png' },
          { id: 3, name: '和平精英', icon: '/assets/icons/game-hpjy.png' },
          { id: 4, name: '原神', icon: '/assets/icons/game-ys.png' },
          { id: 5, name: '无畏契约', icon: '/assets/icons/game-valorant.png' }
        ]
      });
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
      const list = res.list || res.data || [];
      const players = this.data.players.concat(list);

      this.setData({
        players: players,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    }).catch(() => {
      this.setData({
        players: this.data.players.length === 0 ? this.getMockPlayers() : this.data.players,
        loading: false
      });
    });
  },

  getMockPlayers() {
    return [
      {
        id: 1,
        nickname: '王者大神',
        avatar: '/assets/images/default-avatar.png',
        gender: 1,
        is_online: true,
        rank: '王者',
        good_rate: 98,
        order_count: 1560,
        price: 50
      },
      {
        id: 2,
        nickname: '吃鸡少女',
        avatar: '/assets/images/default-avatar.png',
        gender: 2,
        is_online: true,
        rank: '超级王牌',
        good_rate: 95,
        order_count: 890,
        price: 40
      },
      {
        id: 3,
        nickname: 'LOL王者',
        avatar: '/assets/images/default-avatar.png',
        gender: 1,
        is_online: false,
        rank: '钻石',
        good_rate: 92,
        order_count: 2300,
        price: 35
      }
    ];
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