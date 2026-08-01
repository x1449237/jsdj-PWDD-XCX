const request = require('../../../utils/request');

Page({
  data: {
    clubId: 0,
    list: [],
    page: 1,
    limit: 20,
    total: 0,
    loading: false,
    noMore: false,
    role: '',
    keyword: ''
  },

  onLoad(options) {
    const id = parseInt(options.id) || 0;
    this.setData({ clubId: id });
    this.loadList(true);
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
      const res = await request.get('/club/member/list', {
        club_id: this.data.clubId,
        page,
        limit: this.data.limit,
        role: this.data.role,
        keyword: this.data.keyword
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

  onRoleChange(e) {
    const role = e.currentTarget.dataset.role;
    this.setData({ role, page: 1, list: [], noMore: false });
    this.loadList(true);
  },

  onSearch(e) {
    const keyword = e.detail.value;
    this.setData({ keyword, page: 1, list: [], noMore: false });
    this.loadList(true);
  }
});
