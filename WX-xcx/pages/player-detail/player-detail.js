const request = require('../../utils/request');
const util = require('../../utils/util');

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
      this.setData({
        player: {
          id: this.data.playerId,
          nickname: '大神玩家',
          avatar: '/assets/images/default-avatar.png',
          gender: 1,
          is_online: true,
          rank: '王者',
          server: '微信区',
          tags: ['技术过硬', '耐心', '声音好听'],
          good_rate: 98,
          order_count: 1560,
          online_hours: 320,
          fans_count: 5200,
          intro: '资深游戏陪玩，多年游戏经验，擅长各种英雄，可教学可陪玩。',
          min_price: 30,
          services: [
            { id: 1, name: '排位上分', description: '王者荣耀排位赛，包上分', estimated_duration: 30, price: 50, unit: '局' },
            { id: 2, name: '娱乐陪玩', description: '轻松娱乐，边玩边聊', estimated_duration: 30, price: 30, unit: '局' },
            { id: 3, name: '教学指导', description: '一对一教学，讲解英雄技巧', estimated_duration: 60, price: 80, unit: '小时' }
          ]
        }
      });
    });
  },

  loadReviews() {
    request.get(`/players/${this.data.playerId}/reviews`, {
      page: 1,
      page_size: 5
    }).then((res) => {
      const list = (res.list || []).map((item) => ({
        ...item,
        create_time: util.formatRelativeTime(item.create_time)
      }));
      this.setData({ reviews: list });
    }).catch(() => {
      this.setData({
        reviews: [
          {
            id: 1,
            user_avatar: '/assets/images/default-avatar.png',
            user_name: '用户****',
            rating: 5,
            content: '技术很厉害，态度也很好，非常满意！',
            tags: ['技术好', '态度好'],
            create_time: '2小时前'
          },
          {
            id: 2,
            user_avatar: '/assets/images/default-avatar.png',
            user_name: '用户****',
            rating: 5,
            content: '声音好听，打得也好，下次还来~',
            tags: ['声音好听', '技术好'],
            create_time: '1天前'
          }
        ]
      });
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