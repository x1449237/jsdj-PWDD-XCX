const request = require('../../../utils/request');

Page({
  data: {
    clubs: [],
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false,
    clubJoinOpen: true
  },

  onLoad() {
    this.loadClubs();
    this.checkSwitch();
  },

  onPullDownRefresh() {
    this.setData({ page: 1, clubs: [], noMore: false });
    this.loadClubs();
    wx.stopPullDownRefresh();
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadClubs();
    }
  },

  async checkSwitch() {
    try {
      const res = await request.get('/club/check_switch');
      this.setData({ clubJoinOpen: res.data?.club_join_open === true });
    } catch (e) {
      // 忽略
    }
  },

  async loadClubs() {
    if (this.data.loading || this.data.noMore) return;

    this.setData({ loading: true });

    try {
      const res = await request.get('/club/list', {
        page: this.data.page,
        page_size: this.data.pageSize
      });

      const list = res.data?.list || res.data || [];
      const clubs = this.data.clubs.concat(list);

      this.setData({
        clubs: clubs,
        page: this.data.page + 1,
        loading: false,
        noMore: list.length < this.data.pageSize
      });
    } catch (e) {
      this.setData({ loading: false });
    }
  },

  onClubTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/club/detail/detail?id=${id}`
    });
  },

  onJoinClub() {
    if (!this.data.clubJoinOpen) {
      wx.showToast({ title: '俱乐部入驻功能暂未开放', icon: 'none' });
      return;
    }
    wx.navigateTo({
      url: '/pages/club/join/join'
    });
  }
});