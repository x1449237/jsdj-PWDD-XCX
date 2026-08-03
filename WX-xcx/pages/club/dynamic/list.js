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
    type: ''
    // 动态类型文案由后端返回 type_text,前端不硬编码 typeMap
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

  // 点赞:更新后的 like_count/liked 由后端返回,前端不本地自增
  handleLike(e) {
    const id = e.currentTarget.dataset.id;
    const index = e.currentTarget.dataset.index;
    if (!id || index == null) return;
    request.post('/club/dynamic/like', { id }).then((res) => {
      const list = this.data.list;
      if (list[index]) {
        list[index].like_count = res.like_count != null ? res.like_count : list[index].like_count;
        list[index].liked = !!res.liked;
        this.setData({ list });
      }
    }).catch(() => {
      wx.showToast({ title: '操作失败', icon: 'none' });
    });
  },

  previewImage(e) {
    const urls = e.currentTarget.dataset.urls;
    const current = e.currentTarget.dataset.current;
    wx.previewImage({ urls, current });
  }
});
