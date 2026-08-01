const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    list: [],
    page: 1,
    limit: 10,
    total: 0,
    loading: false,
    noMore: false,
    type: '',
    typeMap: {
      achievement: '战绩',
      daily: '日常',
      event: '活动'
    }
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadList(true);
  },

  onShow() {
    if (this.data.clubId > 0) {
      this.loadList(true);
    }
  },

  onPullDownRefresh() {
    this.loadList(true);
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadList(false);
    }
  },

  async loadList(refresh) {
    if (this.data.loading) return;
    this.setData({ loading: true });

    try {
      const page = refresh ? 1 : this.data.page;
      const res = await request.get('/club/dynamic/list', {
        club_id: this.data.clubId,
        page,
        limit: this.data.limit,
        type: this.data.type
      });

      const list = refresh ? res.list : [...this.data.list, ...res.list];
      this.setData({
        list,
        total: res.total || 0,
        page: page + 1,
        noMore: list.length >= (res.total || 0),
        loading: false
      });
    } catch (e) {
      wx.showToast({ title: '加载失败', icon: 'none' });
      this.setData({ loading: false });
    }

    if (refresh) {
      wx.stopPullDownRefresh();
    }
  },

  onTypeChange(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({ type, page: 1, list: [], noMore: false });
    this.loadList(true);
  },

  goPublish() {
    wx.navigateTo({ url: '/pages/club/dynamic/publish?id=' + this.data.clubId });
  },

  handleLike(e) {
    const id = e.currentTarget.dataset.id;
    const index = e.currentTarget.dataset.index;
    const list = this.data.list;
    if (list[index]) {
      list[index].like_count = (list[index].like_count || 0) + 1;
      this.setData({ list });
    }
  },

  previewImage(e) {
    const urls = e.currentTarget.dataset.urls;
    const current = e.currentTarget.dataset.current;
    wx.previewImage({ urls, current });
  }
});
